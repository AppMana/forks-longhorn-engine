//go:build !windows

package controller

import lhns "github.com/longhorn/go-common-libs/ns"

func syncFilesystems() error {
	return lhns.Sync()
}
