// Copyright 2018 Nicholas Ng <nickng@nickng.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"go.nickng.io/asyncpi"
)

// TypedName is a typed wrapper for Name.
// A TypedName can be used as a Name.
type TypedName interface {
	// Name is the wrapped Name.
	asyncpi.Name

	// Type returns the type of the wrapped Name.
	Type() Type

	// setType replaces the type of the TypedName
	// with the parameter.
	setType(Type)
}

// typedName is a concrete typed Name.
type typedName struct {
	// name is the wrapped Name.
	// The Name field shadows the Name() so all
	// Name methods must be implemented.
	name asyncpi.Name

	// t is the type for the wrapped Name.
	t Type
}

// newTypedName returns a new typed Name for the given Name.
func newTypedName(n asyncpi.Name) *typedName {
	return &typedName{n, newAnyType()}
}

func (n *typedName) FreeNames() []asyncpi.Name {
	return n.name.FreeNames()
}

func (n *typedName) FreeVars() []asyncpi.Name {
	return n.name.FreeVars()
}

func (n *typedName) Name() string {
	return n.name.Name()
}

// Type returns the underlying Type of the wrapped Name.
func (n *typedName) Type() Type {
	return n.t
}

var _ TypedName = (*typedName)(nil)

// setType replaces the type of n with t.
func (n *typedName) setType(t Type) {
	n.t = t
}

// AttachType wraps the given n with types.
// The default type is unconstrained.
func AttachType(n asyncpi.Name) TypedName {
	if tn, alreadyTyped := n.(TypedName); alreadyTyped {
		return tn
	}
	// Use type hint
	if th, hasHint := n.(asyncpi.TypeHinter); hasHint {
		tn := newTypedName(n)
		tn.setType(NewBase(th.TypeHint()))
		return tn
	}
	return newTypedName(n)
}
