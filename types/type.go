// Copyright 2018 Nicholas Ng
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

// Package types declares the types and implements the algorithms for
// type inference for programs written using the asyncpi package.
package types

import (
	"bytes"
	"fmt"

	"go.nickng.io/asyncpi"
)

// A Type represents a type in asyncpi.
// All types implement the Type interface.
type Type interface {
	// Underlying returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

// anyType represents an unconstrained type.
// Equivalent to empty interface in Go.
type anyType struct{}

// newAnyType returns a new unconstrained type.
// This is not exported as it should not be used outside of
// conversion from untyped to typed.
func newAnyType() *anyType {
	return &anyType{}
}

// Underlying returns itself as the underlying type of t.
func (t *anyType) Underlying() Type {
	return t
}

func (t *anyType) String() string {
	return "interface{}"
}

// Base represents a base (concrete) type.
// Base types are 'abstract' type only known by their name.
type Base struct {
	name string
}

// NewBase returns a new base type for the given base type name.
func NewBase(name string) *Base {
	return &Base{name}
}

// Underlying returns itself as the underlying type of b.
func (b *Base) Underlying() Type {
	return b
}

func (b *Base) String() string {
	return b.name
}

// Chan represents a channel type.
type Chan struct {
	// elem is the type of the elements that
	// can be transmitted through the channel.
	elem Type
}

// NewChan returns a new channel type for the given element type.
func NewChan(elem Type) *Chan {
	return &Chan{elem}
}

// Elem returns the element type of channel c.
func (c *Chan) Elem() Type {
	return c.elem
}

// Underlying returns itself as the underlying type of c.
func (c *Chan) Underlying() Type {
	return c
}

func (c *Chan) String() string {
	return fmt.Sprintf("chan %s", c.Elem().String())
}

// Composite represents a composite type of multiple types.
// The type is mainly used for representing polyadic parameters of a channel.
type Composite struct {
	// elems is the list of parameter types that made up the composite type.
	elems []Type
}

// NewComposite returns a new Composite type for the given types.
func NewComposite(t ...Type) *Composite {
	comp := &Composite{}
	comp.elems = append(comp.elems, t...)
	return comp
}

// Elems returns the element types of c.
func (c *Composite) Elems() []Type {
	return c.elems
}

// Underlying returns itself as the underlying type of c.
func (c *Composite) Underlying() Type {
	return c
}

// String returns a struct of the composed types.
func (c *Composite) String() string {
	var buf bytes.Buffer
	buf.WriteString("struct{")
	for i, e := range c.elems {
		if i != 0 {
			buf.WriteRune(';')
		}
		buf.WriteString(fmt.Sprintf("e%d %s", i, e.String()))
	}
	buf.WriteString("}")
	return buf.String()
}

// Reference is a reference to the type of a given Name.
type Reference struct {
	ref TypedName
}

// NewReference returns a new Name reference to the given Name.
func NewReference(n asyncpi.Name) *Reference {
	return &Reference{AttachType(n)}
}

// Underlying returns the type of the referenced Name as the underlying type of r.
func (r *Reference) Underlying() Type {
	return r.ref.Type()
}

func (r *Reference) String() string {
	return r.ref.Type().String()
}

// deref peels off layers of Reference from a given type
// and returns the underlying type.
func deref(t Type) Type {
	if refType, ok := t.(*Reference); ok {
		return deref(refType.ref.Type())
	}
	return t
}

// IsEqual compare types.
func IsEqual(t, u Type) bool {
	if t == u {
		return true
	}
	if baseT, tok := deref(t).(*Base); tok {
		if baseU, uok := deref(u).(*Base); uok {
			return baseT.name == baseU.name
		}
	}
	if compT, tok := deref(t).(*Composite); tok {
		if compU, uok := deref(u).(*Composite); uok {
			if len(compT.elems) == 0 && len(compU.elems) == 0 {
				return true
			}
			compEqual := len(compT.elems) == len(compU.elems)
			for i := range compT.elems {
				compEqual = compEqual && IsEqual(compT.elems[i], compU.elems[i])
			}
			return compEqual
		}
	}
	if chanT, tok := deref(t).(*Chan); tok {
		if chanU, uok := deref(u).(*Chan); uok {
			return IsEqual(chanT.elem, chanU.elem)
		}
	}
	return false
}
