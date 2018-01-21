// Copyright 2017 Nicholas Ng <nickng@nickng.io>
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

// Package asyncpi provides a simple library to work with π-calculus.
//
// The asyncpi package consists of a parser of asynchronous π-calculus and Go
// code generator to execute the calculus.
//
// Syntax
//
// The basic syntax of the input language is as follows:
//
//   P,Q ::= 0           nil process
//         | P|Q         parallel composition of P and Q
//         | (new a)P    generation of a with scope P
//         | !P          replication of P, i.e. infinite parallel composition  P|P|P...
//         | u<v>        output of v on channel u
//         | u(x).P      input of distinct variables x on u, with continuation P
//
// The input language accepted is slightly more flexible with some syntactic
// sugar.
// For instance, consecutive new scope can be grouped together, for example:
//
//   (new a,b,c)P
//
// Which is equivalent to:
//
//   (new a)(new b)(new c)P
//
// Another optional syntax accepted in the input language is type annotations on
// the names being created. This feature is mainly designed for code generation.
// The following annotation attaches the type int to the name i.
//
//   (new i:int)P
//
// However, the annotation is only a hint, the type inference may assign i with
// a different type if its usage does not match the annotation.
//
//  (new i:int)i<>
//
// Since i is used as a channel, i cannot be of type int, the annotation is
// therefore ignored.
//
package asyncpi // import "go.nickng.io/asyncpi"
