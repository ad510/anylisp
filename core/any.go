package anylisp

import (
	"fmt"
	"math/big"
	"os"
)

type Lister interface {
	Car() interface{}
	SetCar(v interface{}) interface{}
	Cdr() Lister
	SetCdr(v Lister) Lister
	Last() Lister
}

type List struct {
	car interface{}
	cdr Lister
}

type Inter interface {
	Cmp(y *big.Int) (r int)
	Int64() int64
}

var (
	Ps_      Lister
	C_       Lister
	TempRoot Lister
)

func Parse(code string) {
	TempRoot = &List{"sx'", nil}
	Ps_ = &List{TempRoot, nil}
	C_ = &List{&List{TempRoot, nil}, nil}
	tok := ""
	for i := 0; i < len(code); i++ {
		if code[i] == ' ' || code[i] == '\t' || code[i] == '\n' {
			if tok == ")" {
				Assert(Ps_.Cdr() != nil, "Parse WTF! Too many )s")
				Ps_ = Ps_.Cdr()
			} else if len(tok) > 0 {
				var ls Lister
				if tok == "(" { // list
					ls = &List{nil, nil}
				} else if tok[0] == '[' && tok[len(tok)-1] == ']' { // number
					for j := 1; j < len(tok)-1; j++ {
						Assert(tok[j] == '-' || (tok[j] >= '0' && tok[j] <= '9') || (tok[j] >= 'a' && tok[j] <= 'f'),
							"Parse WTF! Bad character in number")
					}
					bi := new(big.Int)
					_, err := fmt.Sscanf(tok[1:len(tok)-1], "%x", bi)
					Assert(err == nil, "Parse WTF! Bad number")
					ls = &List{bi, nil}
				} else { // symbol
					ls = &List{tok, nil}
				}
				if Ps_.Car() == nil {
					Ps_.Cdr().Car().(Lister).SetCar(ls) // 1st token in list
				} else {
					Ps_.Car().(Lister).SetCdr(ls)
				}
				Ps_.SetCar(ls)
				if tok == "(" {
					Ps_ = &List{nil, Ps_}
				}
			}
			tok = ""
		} else {
			tok += string(code[i])
		}
	}
	Assert(Ps_.Cdr() == nil, "Parse WTF! Too few )s")
}

func Run() {
	for C_ != nil {
		frm, ok := C_.Car().(Lister)
		Assert(ok, "WTF! Bad stack frame")
		HasArg := func(n int) bool {
			return Nth(frm, n-1).Cdr() != nil
		}
		Arg := func(n int) interface{} {
			return Nth(frm, n).Car()
		}
		ArgL := func(n int) Lister {
			arg, ok := Arg(n).(Lister)
			Assert(ok, fmt.Sprintf("WTF! Stack frame argument %d isn't a list", n))
			return arg
		}
		exp, ok := frm.Car().(Lister)
		if !ok {
			fmt.Print("0 ")
			Ret(frm.Car())
		} else {
			switch t := exp.Car().(type) {
			case nil:
				Assert(false, "WTF! Can't call the empty list")
			case Inter:
				Assert(false, "WTF! Can't call an int")
			case Lister:
				Assert(false, "WTF! Can't call a list")
				// I kind of like the behavior below, but it causes strange error messages if there's a bug
				/*if frm.Cdr() == nil {
					fmt.Println("a")
					C_ = &List{&List{t, nil}, C_}
				} else {
					fmt.Println("b")
					frm.SetCar(frm.Cdr().Car())
				}*/
			case string:
				switch t {
				case "sx'": // sx', arg, ret
					if exp.Cdr() == nil {
						fmt.Print("{0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("{1 ")
						frm.SetCdr(&List{exp.Cdr(), nil})
					} else if Arg(1) != nil {
						fmt.Print("{2 ")
						C_ = &List{&List{ArgL(1).Car(), nil}, C_}
						frm.SetCdr(&List{ArgL(1).Cdr(), nil})
					} else {
						fmt.Print("{3 ")
						Ret(Arg(2))
					}
				case "q'":
					fmt.Print("' ")
					Assert(exp.Cdr() != nil, "WTF! Missing argument to quote")
					Ret(exp.Cdr().Car())
				case ":^'", ":>'", ":|'": // op, ret
					if !HasArg(1) {
						fmt.Print(":1 ")
						Assert(exp.Cdr() != nil, "WTF! Missing argument to "+t)
						C_ = &List{&List{exp.Cdr().Car(), nil}, C_}
					} else if Arg(1) == nil {
						fmt.Print(":2 ")
						Ret(nil)
					} else { // TODO: better error message if not a list
						fmt.Print(":3 ")
						switch t {
						case ":^'":
							Ret(ArgL(1).Car())
						case ":>'":
							Ret(ArgL(1).Cdr())
						case ":|'":
							Ret(ArgL(1).Last())
						}
					}
				case "lt'": // lt', args...
					if exp.Cdr() == nil {
						fmt.Print("lt0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("lt1 ")
						frm.SetCdr(&List{exp.Cdr(), nil})
					} else if Arg(1) != nil {
						fmt.Print("lt2 ")
						C_ = &List{&List{ArgL(1).Car(), nil}, C_}
						frm.Cdr().SetCar(ArgL(1).Cdr())
					} else {
						fmt.Print("lt3 ")
						switch t2 := frm.Last().Car().(type) {
						case nil:
							Nth(frm, -2).SetCdr(nil)
						case Lister:
							Nth(frm, -2).SetCdr(t2)
						default:
							Assert(false, "WTF! Last argument to lt' must be a list")
						}
						Ret(Nth(frm, 2))
					}
				case "?'":
					// ?', if part, then part, ret
					// ?', then part, nil, ret
					if exp.Cdr() == nil {
						fmt.Print("?0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("?1 ")
						frm.SetCdr(&List{exp.Cdr(), &List{exp.Cdr().Cdr(), nil}})
					} else if !HasArg(3) {
						fmt.Print("?2 ")
						C_ = &List{&List{ArgL(1).Car(), nil}, C_}
					} else {
						if Arg(2) == nil {
							fmt.Print("?3 ")
							Ret(Arg(3))
						} else if Arg(3) != nil {
							fmt.Print("?4 ")
							frm.SetCdr(&List{ArgL(1).Cdr(), &List{nil, nil}})
						} else if ArgL(2).Cdr() == nil {
							fmt.Print("?5 ")
							Ret(nil)
						} else {
							fmt.Print("?6 ")
							frm.SetCdr(&List{ArgL(2).Cdr(), &List{ArgL(2).Cdr().Cdr(), nil}})
						}
					}
				case "pr'":
					// pr', ret
					// TODO: print all arguments
					if exp.Cdr() == nil {
						fmt.Print("pr0 ")
						Ret(nil)
					} else if !HasArg(1) {
						fmt.Print("pr1 ")
						C_ = &List{&List{exp.Cdr().Car(), nil}, C_}
					} else if Arg(1) == nil {
						fmt.Print("pr2 ")
						Ret(nil)
					} else {
						fmt.Print("pr3 ")
						s := make([]uint8, Len(ArgL(1))) // TODO: better error message if not a list
						for i, arg := 0, ArgL(1); arg != nil; i, arg = i+1, arg.Cdr() {
							c, ok := arg.Car().(Inter)
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
	case Inter:
		fmt.Printf("[%x] ", t)
	case Lister:
		fmt.Print("( ")
		for ls != nil {
			PrintTree(ls.(Lister).Car())
			ls = ls.(Lister).Cdr()
		}
		fmt.Print(") ")
	case string:
		fmt.Print(t + " ")
	default:
		Assert(false, "Unrecognized object in tree")
	}
}

func Ret(v interface{}) {
	if C_.Cdr() != nil {
		C_.Cdr().Car().(Lister).Last().SetCdr(&List{v, nil})
	}
	C_ = C_.Cdr()
}

func Len(ls Lister) int {
	if ls == nil {
		return 0
	}
	if ls.Cdr() == nil {
		return 1
	}
	return Len(ls.Cdr()) + 1
}

func Nth(ls Lister, n int) Lister {
	if ls == nil {
		return nil
	}
	if n > 0 {
		return Nth(ls.Cdr(), n-1)
	}
	if n < 0 {
		return Nth(ls, Len(ls)+n)
	}
	return ls
}

func (ls *List) Car() interface{} {
	return ls.car
}

func (ls *List) SetCar(v interface{}) interface{} {
	ls.car = v
	return v
}

func (ls *List) Cdr() Lister {
	return ls.cdr
}

func (ls *List) SetCdr(v Lister) Lister {
	ls.cdr = v
	return v
}

func (ls *List) Last() Lister {
	if ls.cdr == nil {
		return ls
	}
	return ls.cdr.Last()
}

func Assert(cond bool, msg string) {
	if !cond {
		fmt.Println(msg)
		os.Exit(2)
	}
}
