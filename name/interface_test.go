package name

import (
	"testing"

	"go.nickng.io/asyncpi/internal/name"
)

// This test ensures the Setter is in sync with internal version.
func TestSetterSync(t *testing.T) {
	var n Setter = name.New("base")
	if _, ok := n.(name.Setter); !ok {
		t.Fatalf("asyncpi/name.Setter and asyncpi/internal/name.Setter out of sync")
	}
}
