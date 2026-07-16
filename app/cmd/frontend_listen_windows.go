//go:build windows

package cmd

import (
	"fmt"
	"net"
	"strconv"
)

func frontendListenAddress(controllerAddress, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	host, portText, err := net.SplitHostPort(controllerAddress)
	if err != nil {
		return "", fmt.Errorf("invalid controller listen address %q: %w", controllerAddress, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port >= 65535 {
		return "", fmt.Errorf("invalid controller port %q", portText)
	}
	return net.JoinHostPort(host, strconv.Itoa(port+1)), nil
}
