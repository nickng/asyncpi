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

// Package name contains support functions for working with Name.
package name

import "go.nickng.io/asyncpi"

// IsSame is the equality operator for Name.
//
// A Name m equals another Name n if m and n has the same Ident.
// The comparison ignores the underlying represention
// (sortedname, typedname, etc.)
func IsSame(m, n asyncpi.Name) bool {
	return m.Ident() == n.Ident()
}

// IsFreeName returns true if a given Name n is Free.
//
// This is just a convenient wrapper for FreeNames(n) which return a slice.
func IsFreeName(n asyncpi.Name) bool {
	return len(asyncpi.FreeNames(n)) == 1 && asyncpi.FreeNames(n)[0].Ident() == n.Ident()
}
