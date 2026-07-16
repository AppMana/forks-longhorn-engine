//go:build !windows

package cmd

func frontendListenAddress(_, explicit string) (string, error) {
	return explicit, nil
}
