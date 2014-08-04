package anylisp

import (
	"fmt"
	"os"
)

type AnyList struct {
	Fi *interface{}
	Bf *AnyList
}

var (
	ParseStack *AnyList
	TempRoot *AnyList
)

func Parse(code string) {
	var tok string
	ParseStack = &AnyList{}
	TempRoot = ParseStack
	for i := 0; i < len(code); i++ {
		if code[i] == ' ' || code[i] == '\t' || code[i] == '\n' {
			if tok == "(" {
				ParseStack = &AnyList{nil, ParseStack}
			}else if tok == ")"{
				Assert(ParseStack.Bf != nil, "Parse WTF! Too many )s.")
				ParseStack = ParseStack.Bf
			}else if len(tok) > 0 {
				//ParseStack.Fi = &AnyList{&fmt.Stringer(tok), nil/*ParseStack.Fi.(*AnyList)*/}
			}
			tok = ""
		} else {
			tok += string(code[i])
		}
	}
	fmt.Println(ParseStack)
}

func Ln(ls *AnyList) int {
	if ls == nil {return 0}
	if ls.Bf == nil {return 1}
	return Ln(ls.Bf) + 1
}

func Nth(ls *AnyList, n int) *AnyList {
	Assert(ls != nil, "WTF! Out of bounds when calling (nth.")
	if n > 0 {return Nth(ls.Bf, n - 1)}
	if n < 0 {return Nth(ls, Ln(ls) - n)}
	return ls
}

func Assert(cond bool, msg string) {
	if !cond {
		fmt.Println(msg)
		os.Exit(2)
	}
}
