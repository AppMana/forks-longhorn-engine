//go:build !windows

package namedpipe

import (
	"fmt"
	"net"
)

func Listen(path string) (net.Listener, error) {
	return nil, fmt.Errorf("Windows named pipe %q is unavailable on this platform", path)
}

func Dial(path string) (net.Conn, error) {
	return nil, fmt.Errorf("Windows named pipe %q is unavailable on this platform", path)
}
