package main

import (
	"github.com/ad510/anylisp/core"
	"io/ioutil"
	"os"
)

func main() {
	anylisp.Assert(len(os.Args) >= 2, "Usage: anylisp program.any [args]")
	file, err := ioutil.ReadFile(os.Args[1])
	anylisp.Assert(err == nil, "'"+os.Args[1]+"' not found.")
	anylisp.Parse(string(file))
}
