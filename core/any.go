package anylisp

import (
	"fmt"
	"os"
)

type AnyLister interface {
	Gfi() interface{}
	Sfi(v interface{}) interface{}
	Gbf() AnyLister
	Sbf(v AnyLister) AnyLister
}

type AnyList struct {
	Fi interface{}
	Bf AnyLister
}

var (
	Ps_      AnyLister
	TempRoot AnyLister
)

func Parse(code string) {
	TempRoot = &AnyList{}
	Ps_ = &AnyList{TempRoot, nil}
	tok := ""
	for i := 0; i < len(code); i++ {
		if code[i] == ' ' || code[i] == '\t' || code[i] == '\n' {
			if tok == ")" {
				Assert(Ps_.Gbf() != nil, "Parse WTF! Too many )s.")
				Ps_ = Ps_.Gbf()
			} else if len(tok) > 0 {
				var ls AnyLister
				if tok == "(" {
					ls = &AnyList{nil, nil}
				} else {
					ls = &AnyList{tok, nil}
				}
				if Ps_.Gfi() == nil {
					Ps_.Gbf().Gfi().(AnyLister).Sfi(ls) // 1st token in list
				} else {
					Ps_.Gfi().(AnyLister).Sbf(ls)
				}
				Ps_.Sfi(ls)
				if tok == "(" {
					Ps_ = &AnyList{nil, Ps_}
				}
			}
			tok = ""
		} else {
			tok += string(code[i])
		}
	}
	Assert(Ps_.Gbf() == nil, "Parse WTF! Too few )s.")
}

func PrintTree(ls interface{}) {
	switch t := ls.(type) {
	case string:
		fmt.Print(t + " ")
	case AnyLister:
		fmt.Print("(")
		for ls != nil {
			PrintTree(ls.(AnyLister).Gfi())
			ls = ls.(AnyLister).Gbf()
		}
		fmt.Print(")")
	}
}

func Ln(ls AnyLister) int {
	if ls == nil {
		return 0
	}
	if ls.Gbf() == nil {
		return 1
	}
	return Ln(ls.Gbf()) + 1
}

func Nth(ls AnyLister, n int) AnyLister {
	Assert(ls != nil, "WTF! Out of bounds when calling (nth.")
	if n > 0 {
		return Nth(ls.Gbf(), n-1)
	}
	if n < 0 {
		return Nth(ls, Ln(ls)-n)
	}
	return ls
}

func (ls *AnyList) Gfi() interface{} {
	return ls.Fi
}

func (ls *AnyList) Sfi(v interface{}) interface{} {
	ls.Fi = v
	return v
}

func (ls *AnyList) Gbf() AnyLister {
	return ls.Bf
}

func (ls *AnyList) Sbf(v AnyLister) AnyLister {
	ls.Bf = v
	return v
}

func Assert(cond bool, msg string) {
	if !cond {
		fmt.Println(msg)
		os.Exit(2)
	}
}
