package socket

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"github.com/longhorn/longhorn-engine/pkg/dataconn"
	"github.com/longhorn/longhorn-engine/pkg/types"
)

const (
	frontendName = "socket"

	SocketDirectory = "/var/run"
	DevPath         = "/dev/longhorn/"
)

func New() *Socket {
	return &Socket{}
}

type Socket struct {
	Volume      string
	Size        int64
	SectorSize  int
	ScsiTimeout int

	isUp          bool
	network       string
	listenAddress string
	socketPath    string
	listener      net.Listener
	socketServer  *dataconn.Server
}

// NewNetwork creates a data socket frontend on an explicit network endpoint.
// The default New frontend remains a Unix socket so the Linux target path and
// upgrade behaviour are unchanged.
func NewNetwork(network, listenAddress string) *Socket {
	return &Socket{network: network, listenAddress: listenAddress}
}

func (t *Socket) FrontendName() string {
	return frontendName
}

func (t *Socket) Init(name string, size, sectorSize int64) error {
	t.Volume = name
	t.Size = size
	t.SectorSize = int(sectorSize)

	return t.Shutdown()
}

func (t *Socket) Startup(rwu types.ReaderWriterUnmapperAt) error {
	if err := t.startSocketServer(rwu); err != nil {
		return err
	}

	t.isUp = true

	return nil
}

func (t *Socket) Shutdown() error {
	if t.Volume != "" {
		if t.socketServer != nil {
			logrus.Infof("Shutting down TGT socket server for %v", t.Volume)
			t.socketServer.Stop()
			t.socketServer = nil
		}
		if t.listener != nil {
			if err := t.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				return err
			}
			t.listener = nil
		}
	}
	t.isUp = false

	return nil
}

func (t *Socket) State() types.State {
	if t.isUp {
		return types.StateUp
	}
	return types.StateDown
}

func (t *Socket) Endpoint() string {
	if t.isUp {
		if t.listenAddress != "" {
			return t.listenAddress
		}
		return t.GetSocketPath()
	}
	return ""
}

func (t *Socket) ListenAddress() string {
	return t.listenAddress
}

func (t *Socket) GetSocketPath() string {
	if t.Volume == "" {
		panic("Invalid volume name")
	}
	return filepath.Join(SocketDirectory, "longhorn-"+t.Volume+".sock")
}

func (t *Socket) startSocketServer(rwu types.ReaderWriterUnmapperAt) error {
	if t.network != "" {
		return t.startNetworkServer(rwu)
	}
	socketPath := t.GetSocketPath()
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return errors.Wrapf(err, "cannot create directory %v", filepath.Dir(socketPath))
	}
	// Check and remove existing socket
	if st, err := os.Stat(socketPath); err == nil && !st.IsDir() {
		if err := os.Remove(socketPath); err != nil {
			return err
		}
	}

	t.socketPath = socketPath
	return t.startNetworkServer(rwu)
}

func (t *Socket) startNetworkServer(rwu types.ReaderWriterUnmapperAt) error {
	network, address := t.network, t.listenAddress
	if network == "" {
		network, address = "unix", t.socketPath
	}
	ln, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	t.listener = ln
	go t.acceptConnections(ln, rwu)
	return nil
}

func (t *Socket) acceptConnections(ln net.Listener, rwu types.ReaderWriterUnmapperAt) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			logrus.WithError(err).Error("Failed to accept socket connection")
			continue
		}
		go t.handleServerConnection(conn, rwu)
	}
}

func (t *Socket) handleServerConnection(c net.Conn, rwu types.ReaderWriterUnmapperAt) {
	defer func() {
		if errClose := c.Close(); errClose != nil {
			logrus.WithError(errClose).Error("Failed to close socket connection")
		}
	}()

	server := dataconn.NewServer(c, NewDataProcessorWrapper(rwu))
	logrus.Info("New data socket connection established")
	if err := server.Handle(); err != nil && err != io.EOF {
		logrus.WithError(err).Errorf("Failed to handle socket server connection")
	} else if err == io.EOF {
		logrus.Warn("Socket server connection closed")
	}
}

type DataProcessorWrapper struct {
	rwu types.ReaderWriterUnmapperAt
}

func NewDataProcessorWrapper(rwu types.ReaderWriterUnmapperAt) DataProcessorWrapper {
	return DataProcessorWrapper{
		rwu: rwu,
	}
}

func (d DataProcessorWrapper) ReadAt(p []byte, off int64) (n int, err error) {
	return d.rwu.ReadAt(p, off)
}

func (d DataProcessorWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	return d.rwu.WriteAt(p, off)
}

func (d DataProcessorWrapper) UnmapAt(length uint32, off int64) (n int, err error) {
	return d.rwu.UnmapAt(length, off)
}

func (d DataProcessorWrapper) PingResponse() error {
	return nil
}

func (t *Socket) Upgrade(name string, size, sectorSize int64, rwu types.ReaderWriterUnmapperAt) error {
	return fmt.Errorf("upgrade is not supported")
}

func (t *Socket) Expand(size int64) error {
	return fmt.Errorf("expand is not supported")
}
