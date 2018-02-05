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

// Package sortedname provides a specialised Name implementation
// associating sorts with a Name. Sorts are used for distinguishing
// between normal Names as constants/values and variables.
package sortedname

import (
	"go.nickng.io/asyncpi"
)

// Sorts is the type for a sort.
type Sorts int

const (
	// NameSort is a sort for names (values or constants).
	// A Name is of NameSort by default.
	NameSort Sorts = iota

	// VarSort is a sort for variables.
	VarSort
)

// setter is an interface to test if a Name has a sort and can be set.
type setter interface {
	SetSort(Sorts)
}

// SortedName implements a Name with sort.
type SortedName struct {
	asyncpi.Name
	s Sorts
}

// New returns a SortedName wrapping Name n.
func New(n asyncpi.Name) *SortedName {
	return &SortedName{Name: n}
}

// NewWithSort returns a SortedName wrapping Name n with a given sort s.
func NewWithSort(n asyncpi.Name, s Sorts) *SortedName {
	return &SortedName{Name: n, s: s}
}

// Sort returns the sort of the SortedName n.
func (n *SortedName) Sort() Sorts {
	return n.s
}

// SetSort sets the sort of the given SortedName n.
func (n *SortedName) SetSort(s Sorts) {
	n.s = s
}

func (n *SortedName) FreeNames() []asyncpi.Name {
	if n.s == NameSort {
		return []asyncpi.Name{n}
	}
	return nil
}

func (n *SortedName) FreeVars() []asyncpi.Name {
	if n.s == VarSort {
		return []asyncpi.Name{n}
	}
	return nil
}
