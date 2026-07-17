//go:build windows

package namedpipe

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestListenAndDial(t *testing.T) {
	listener, err := Listen(Path(fmt.Sprintf("test-%d", os.Getpid())))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan error, 1)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			accepted <- err
			return
		}
		defer connection.Close()
		buffer := make([]byte, 4)
		_, err = io.ReadFull(connection, buffer)
		if err == nil && string(buffer) != "ping" {
			err = fmt.Errorf("unexpected payload %q", buffer)
		}
		accepted <- err
	}()

	connection, err := Dial(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	_ = connection.Close()
	if err := <-accepted; err != nil {
		t.Fatal(err)
	}
}
