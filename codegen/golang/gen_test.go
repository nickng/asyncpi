package golang_test

import (
	"fmt"
	"os"
	"strings"

	"go.nickng.io/asyncpi"
	"go.nickng.io/asyncpi/codegen/golang"
	"go.nickng.io/asyncpi/types"
)

// This example shows how to generate code from a Process.
func ExampleGenerate() {
	p, err := asyncpi.Parse(strings.NewReader("(new a)(new b)(a<b> | a(x).x<> | b().0)"))
	if err != nil {
		fmt.Println(err) // Parse failed
	}
	p = asyncpi.Bind(p)
	types.Infer(p)
	err = types.Unify(p)
	if err != nil {
		fmt.Println(err) // Unify failed
	}
	golang.Generate(p, os.Stdout)
	// Output: a := make(chan chan struct{}); b := make(chan struct{}); go func(){ go func(){ a <- b; }()
	//x := <-a;x <- struct{}{}; }()
	//<-b;/* end */
}
