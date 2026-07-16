package controller

// Windows has no system-wide sync operation. Replica layer transitions close
// and sync their file handles before publishing new metadata, so there is no
// additional host-global flush to perform here.
func syncFilesystems() error {
	return nil
}
