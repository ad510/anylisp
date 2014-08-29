package anylisp

import (
	"fmt"
	"math/big"
	"os"
)

type AnyLister interface {
	Fi() interface{}
	Sfi(v interface{}) interface{}
	Bf() AnyLister
	Sbf(v AnyLister) AnyLister
	La() AnyLister
}

type AnyList struct {
	fi interface{}
	bf AnyLister
}

type AnyInter interface {
	Cmp(y *big.Int) (r int)
	Int64() int64
}

var (
	Ps_      AnyLister
	C_       AnyLister
	TempRoot AnyLister
)

func Parse(code string) {
	TempRoot = &AnyList{"sx'", nil}
	Ps_ = &AnyList{TempRoot, nil}
	C_ = &AnyList{&AnyList{TempRoot, nil}, nil}
	tok := ""
	for i := 0; i < len(code); i++ {
		if code[i] == ' ' || code[i] == '\t' || code[i] == '\n' {
			if tok == ")" {
				Assert(Ps_.Bf() != nil, "Parse WTF! Too many )s")
				Ps_ = Ps_.Bf()
			} else if len(tok) > 0 {
				var ls AnyLister
				if tok == "(" { // list
					ls = &AnyList{nil, nil}
				} else if tok[0] == '[' && tok[len(tok)-1] == ']' { // number
					for j := 1; j < len(tok)-1; j++ {
						Assert(tok[j] == '-' || (tok[j] >= '0' && tok[j] <= '9') || (tok[j] >= 'a' && tok[j] <= 'f'),
							"Parse WTF! Bad character in number")
					}
					bi := new(big.Int)
					_, err := fmt.Sscanf(tok[1:len(tok)-1], "%x", bi)
					Assert(err == nil, "Parse WTF! Bad number")
					ls = &AnyList{bi, nil}
				} else { // symbol
					ls = &AnyList{tok, nil}
				}
				if Ps_.Fi() == nil {
					Ps_.Bf().Fi().(AnyLister).Sfi(ls) // 1st token in list
				} else {
					Ps_.Fi().(AnyLister).Sbf(ls)
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
	Assert(Ps_.Bf() == nil, "Parse WTF! Too few )s")
}

func Run() {
	for C_ != nil {
		frm, ok := C_.Fi().(AnyLister)
		Assert(ok, "WTF! Bad stack frame")
		HasArg := func(n int) bool {
			return Nth(frm, n-1).Bf() != nil
		}
		Arg := func(n int) interface{} {
			return Nth(frm, n).Fi()
		}
		ArgL := func(n int) AnyLister {
			arg, ok := Arg(n).(AnyLister)
			Assert(ok, fmt.Sprintf("WTF! Stack frame argument %d isn't a list", n))
			return arg
		}
		exp, ok := frm.Fi().(AnyLister)
		if !ok {
			fmt.Print("0 ")
			Ret(frm.Fi())
		} else {
			switch t := exp.Fi().(type) {
			case nil:
				Assert(false, "WTF! Can't call the empty list")
			case AnyInter:
				Assert(false, "WTF! Can't call an int")
			case AnyLister:
				Assert(false, "WTF! Can't call a list")
				// I kind of like the behavior below, but it causes strange error messages if there's a bug
				/*if frm.Bf() == nil {
					fmt.Println("a")
					C_ = &AnyList{&AnyList{t, nil}, C_}
				} else {
					fmt.Println("b")
					frm.Sfi(frm.Bf().Fi())
				}*/
			case string:
				switch t {
				case "sx'": // sx', arg, ret
					if exp.Bf() == nil {
						fmt.Print("{0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("{1 ")
						frm.Sbf(&AnyList{exp.Bf(), nil})
					} else if Arg(1) == nil {
						fmt.Print("{2 ")
						Ret(Arg(2))
					} else {
						fmt.Print("{3 ")
						C_ = &AnyList{&AnyList{ArgL(1).Fi(), nil}, C_}
						frm.Sbf(&AnyList{ArgL(1).Bf(), nil})
					}
				case "q'":
					if exp.Bf() == nil {
						fmt.Print("'0 ")
						Ret(nil)
					} else {
						fmt.Print("'1 ")
						Ret(exp.Bf().Fi())
					}
				case "?'":
					// ?', if part, then part, ret
					// ?', then part, nil, ret
					if exp.Bf() == nil {
						fmt.Print("?0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("?1 ")
						frm.Sbf(&AnyList{exp.Bf(), &AnyList{exp.Bf().Bf(), nil}})
					} else if HasArg(3) {
						if Arg(2) == nil {
							fmt.Print("?2 ")
							Ret(Arg(3))
						} else if Arg(3) != nil {
							fmt.Print("?3 ")
							frm.Sbf(&AnyList{ArgL(1).Bf(), &AnyList{nil, nil}})
						} else if ArgL(2).Bf() == nil {
							fmt.Print("?4 ")
							Ret(nil)
						} else {
							fmt.Print("?5 ")
							frm.Sbf(&AnyList{ArgL(2).Bf(), &AnyList{ArgL(2).Bf().Bf(), nil}})
						}
					} else {
						fmt.Print("?6 ")
						C_ = &AnyList{&AnyList{ArgL(1).Fi(), nil}, C_}
					}
				case "pr'":
					// pr', ret
					if exp.Bf() == nil {
						fmt.Print("pr0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("pr1 ")
						C_ = &AnyList{&AnyList{exp.Bf().Fi(), nil}, C_}
					} else {
						fmt.Print("pr2 ")
						s := make([]uint8, Ln(ArgL(1)))
						for i, arg := 0, ArgL(1); arg != nil; i, arg = i+1, arg.Bf() {
							c, ok := arg.Fi().(AnyInter)
							Assert(ok && c.Cmp(big.NewInt(-1)) == 1 && c.Cmp(big.NewInt(256)) == -1,
								"WTF! pr' takes a string")
							s[i] = uint8(c.Int64())
						}
						fmt.Print(string(s))
						Ret(Arg(1))
					}
				default:
					Assert(false, "WTF! Can't call undefined function \""+t+"\"")
				}
			default:
				Assert(false, "WTF! Unrecognized function type (probably an interpreter bug)")
			}
		}
	}
}

func PrintTree(ls interface{}) {
	switch t := ls.(type) {
	case nil:
		fmt.Print("( ) ")
	case AnyLister:
		fmt.Print("( ")
		for ls != nil {
			PrintTree(ls.(AnyLister).Fi())
			ls = ls.(AnyLister).Bf()
		}
		fmt.Print(") ")
	case string:
		fmt.Print(t + " ")
	}
}

func Ret(v interface{}) {
	if C_.Bf() != nil {
		C_.Bf().Fi().(AnyLister).La().Sbf(&AnyList{v, nil})
	}
	C_ = C_.Bf()
}

func Ln(ls AnyLister) int {
	if ls == nil {
		return 0
	}
	if ls.Bf() == nil {
		return 1
	}
	return Ln(ls.Bf()) + 1
}

func Nth(ls AnyLister, n int) AnyLister {
	Assert(ls != nil, "WTF! Out of bounds when calling nth'.")
	if n > 0 {
		return Nth(ls.Bf(), n-1)
	}
	if n < 0 {
		return Nth(ls, Ln(ls)-n)
	}
	return ls
}

func (ls *AnyList) Fi() interface{} {
	return ls.fi
}

func (ls *AnyList) Sfi(v interface{}) interface{} {
	ls.fi = v
	return v
}

func (ls *AnyList) Bf() AnyLister {
	return ls.bf
}

func (ls *AnyList) Sbf(v AnyLister) AnyLister {
	ls.bf = v
	return v
}

func (ls *AnyList) La() AnyLister {
	if ls.bf == nil {
		return ls
	}
	return ls.bf.La()
}

func Assert(cond bool, msg string) {
	if !cond {
		fmt.Println(msg)
		os.Exit(2)
	}
}
