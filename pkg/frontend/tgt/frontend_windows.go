//go:build windows

package tgt

import (
	"fmt"
	"net"
	"time"

	"github.com/longhorn/longhorn-engine/pkg/frontend/socket"
	"github.com/longhorn/longhorn-engine/pkg/types"
	"github.com/longhorn/longhorn-engine/pkg/util"
)

const (
	DevPath         = ""
	DefaultTargetID = 1
)

// Tgt is the per-engine half of the Windows frontend. It deliberately serves
// only the engine data protocol. The node-resident instance manager owns the
// long-lived iSCSI target and switches that target between old and replacement
// engine data endpoints before terminating the old engine.
type Tgt struct {
	s            *socket.Socket
	frontendName string
	name         string
	targetName   string
	isUp         bool
}

func New(frontendName string, _, _, _ time.Duration, listenAddress ...string) types.Frontend {
	address := ""
	if len(listenAddress) > 0 {
		address = listenAddress[0]
	}
	return &Tgt{s: socket.NewNetwork("tcp", address), frontendName: frontendName}
}

func (t *Tgt) FrontendName() string { return t.frontendName }

func (t *Tgt) ListenAddress() string { return t.s.ListenAddress() }

func (t *Tgt) Init(name string, size, sectorSize int64) error {
	if t.frontendName != types.EngineFrontendISCSI {
		return fmt.Errorf("Windows supports %s, not %s", types.EngineFrontendISCSI, t.frontendName)
	}
	if _, _, err := net.SplitHostPort(t.s.ListenAddress()); err != nil {
		return fmt.Errorf("Windows engine data frontend requires an IP:port listen address: %w", err)
	}
	t.name = name
	t.targetName = "iqn.2019-10.io.longhorn:" + util.Volume2ISCSIName(name)
	return t.s.Init(name, size, sectorSize)
}

func (t *Tgt) Startup(rwu types.ReaderWriterUnmapperAt) error {
	if err := t.s.Startup(rwu); err != nil {
		return err
	}
	t.isUp = true
	return nil
}

func (t *Tgt) Shutdown() error {
	if err := t.s.Shutdown(); err != nil {
		return err
	}
	t.isUp = false
	return nil
}

func (t *Tgt) State() types.State {
	if t.isUp {
		return types.StateUp
	}
	return types.StateDown
}

func (t *Tgt) Endpoint() string {
	if t.isUp {
		return t.targetName
	}
	return ""
}

func (t *Tgt) Upgrade(name string, size, sectorSize int64, rwu types.ReaderWriterUnmapperAt) error {
	if err := t.Init(name, size, sectorSize); err != nil {
		return err
	}
	return t.Startup(rwu)
}

func (t *Tgt) Expand(size int64) error { return nil }
