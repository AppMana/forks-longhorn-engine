package namedpipe

import "strings"

const prefix = `\\.\pipe\longhorn-`

// Path returns the node-local Windows named pipe used for one volume's data
// plane. Volume names are Kubernetes resource names, but defensively remove
// path separators so a caller cannot escape the Longhorn pipe namespace.
func Path(volumeName string) string {
	name := strings.NewReplacer(`\`, "-", "/", "-").Replace(volumeName)
	return prefix + name
}
