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

// Package name provides internal default implementations of the asyncpi Names.
package name

// base is a default Name implementation.
type base struct {
	name string
}

// New returns a new concrete name from a string.
func New(name string) *base {
	return &base{name}
}

// Ident returns the string identifier of the base name n.
func (n *base) Ident() string {
	return n.name
}

// SetName sets the internal name.
func (n *base) SetName(name string) {
	n.name = name
}

func (n *base) String() string {
	return n.name
}

// hinted represents a name with type hint.
type hinted struct {
	name string
	hint string
}

// NewHinted returns a new hinted name from a string name and type hint.
func NewHinted(name, hint string) *hinted {
	return &hinted{name, hint}
}

// Ident returns the string identifier of the hinted name n.
func (n *hinted) Ident() string {
	return n.name
}

// SetName sets the internal name.
func (n *hinted) SetName(name string) {
	n.name = name
}

func (n *hinted) String() string {
	return n.name
}

// TypeHint returns the type hint of hinted name n.
func (n *hinted) TypeHint() string {
	return n.hint
}
