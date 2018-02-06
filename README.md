# asyncpi [![Build Status](https://travis-ci.org/nickng/asyncpi.svg?branch=master)](https://travis-ci.org/nickng/asyncpi) [![GoDoc](https://godoc.org/go.nickng.io/asyncpi?status.svg)](http://godoc.org/go.nickng.io/asyncpi)

## An implementation of asynchronous π-calculus in Go.

The basic syntax accept is given below, for details (including syntactic sugar),
see [godoc](http://godoc.org/github.com/nickng/asyncpi).

    P,Q ::= 0           nil process
          | P|Q         parallel composition of P and Q
          | (new a)P    generation of a with scope P
          | !P          replication of P, i.e. infinite parallel composition  P|P|P...
          | u<v>        output of v on channel u
          | u(x).P      input of distinct variables x on u, with continuation P

## Install

    go get -u go.nickng.io/asyncpi

## Play

`cmd/asyncpi` is a simple REPL front end for the package, with very
basic support for *parsing*,
*free name calculation*, *process reduction* and *code fragment generation*.

### Build and install

    go install go.nickng.io/asyncpi/cmd/asyncpi

### Run

    $ asyncpi
    async-π> parse
    .......> a<b,c,d> | a(x,y,z).x().0 | b<> | c(z).0 | (new c)c<d>
    ((((a<b,c,d> | a(x,y,z).x().0) | b<>) | c(z).0) | (new c)c<d>)
    async-π> reduce
    Reducing: ((((a<b,c,d> | a(x,y,z).x().0) | b<>) | c(z).0) | (new c)c<d>)
    (((b().0 | b<>) | c(z).0) | (new c)c<d>)
    async-π> reduce
    Reducing: (((b().0 | b<>) | c(z).0) | (new c)c<d>)
    (c(z).0 | (new c)c<d>)
    async-π> reduce
    Reducing: (c(z).0 | (new c)c<d>)
    (c(z).0 | (new c)c<d>)
    async-π> codegen
    /* start generated code */

    go func() { z := <-c /* end */ }()
    c := make(chan interface{})
    c <- d

    /* end generated code */
    async-π> exit

## License

asyncpi is licensed under the [Apache License](http://www.apache.org/licenses/LICENSE-2.0)
