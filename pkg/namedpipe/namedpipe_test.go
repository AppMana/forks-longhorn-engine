package namedpipe

import "testing"

func TestPath(t *testing.T) {
	want := `\\.\pipe\longhorn-pvc-123-name`
	if got := Path(`pvc-123/name`); got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}
