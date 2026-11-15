// Command eval is the balance gate: it plays the scripted pilot over a fixed
// run of seeds, with the hunter parked and then against it, and exits non-zero
// when a pass mark in internal/eval is missed. The marks live in code so the
// local run and CI can never disagree.
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/nemesis/internal/eval"
)

func main() {
	r := eval.Run(eval.DefaultConfig())
	fmt.Println(r)
	failures := r.Failures()
	for _, f := range failures {
		fmt.Fprintln(os.Stderr, "eval: FAIL:", f)
	}
	if len(failures) > 0 {
		os.Exit(1)
	}
	fmt.Println("eval: PASS")
}
