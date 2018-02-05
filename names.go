package asyncpi

import (
	"go.nickng.io/asyncpi/internal/name"
)

// newNames is a convenient utility function
// for creating a []Name from given strings.
func newNames(names ...string) []Name {
	pn := make([]Name, len(names))
	for i, n := range names {
		pn[i] = name.New(n)
	}
	return pn
}

type TypeHinter interface {
	TypeHint() string
}

type nameSetter interface {
	SetName(string)
}

// freeNameser is an interface which Name should
// provide to have custom FreeNames implementation.
type freeNameser interface {
	FreeNames() []Name
}

// FreeNames returns the free names in a give Name n.
func FreeNames(n Name) []Name {
	if fn, ok := n.(freeNameser); ok {
		return fn.FreeNames()
	}
	return []Name{n}
}

// freeVarser is an interface which Name should
// provide to have custom FreeVars implementation.
type freeVarser interface {
	FreeVars() []Name
}

// FreeVars returns the free variables in a give Name n.
func FreeVars(n Name) []Name {
	if fv, ok := n.(freeVarser); ok {
		return fv.FreeVars()
	}
	return nil
}

// IsSameName is a simple comparison operator for Name.
// A Name x is equal with another Name y when x and y has the same name.
// This comparison ignores the underlying representation (sort, type, etc.).
func IsSameName(x, y Name) bool {
	return x.Ident() == y.Ident()
}

func IsFreeName(x Name) bool {
	return len(FreeNames(x)) == 1 && FreeNames(x)[0].Ident() == x.Ident()
}
