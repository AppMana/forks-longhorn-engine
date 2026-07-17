//go:build windows

package namedpipe

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func Listen(path string) (net.Listener, error) {
	return winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;BA)",
	})
}

func Dial(path string) (net.Conn, error) {
	return winio.DialPipe(path, nil)
}
