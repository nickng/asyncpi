# asyncpi [![Build Status](https://travis-ci.org/nickng/asyncpi.svg?branch=master)](https://travis-ci.org/nickng/asyncpi) [![GoDoc](https://godoc.org/github.com/nickng/asyncpi?status.svg)](http://godoc.org/github.com/nickng/asyncpi)

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

    go get github.com/nickng/asyncpi

## License

asyncpi is licensed under the [Apache License](http://www.apache.org/licenses/LICENSE-2.0)
