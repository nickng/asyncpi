package name_test

import (
	"testing"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/internal/name"
)

func TestBaseName(t *testing.T) {
	var n asyncpi.Name
	n = name.New("abc")
	if _, ok := n.(name.Setter); !ok {
		t.Fatalf("%s should have SetName()", n.Ident())
	}
}

func TestHuntedName(t *testing.T) {
	var n asyncpi.Name
	n = name.NewHinted("hinted", "int")
	if _, ok := n.(name.Setter); !ok {
		t.Fatalf("%v should have SetName()", n)
	}
	if _, ok := n.(name.TypeHinter); !ok {
		t.Fatalf("%v should have TypeHint()", n)
	}
}
