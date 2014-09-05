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
	Add(x, y *big.Int) *big.Int
	Cmp(y *big.Int) (r int)
	Int64() int64
	Mul(x, y *big.Int) *big.Int
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
	cm := false
	for i := 0; i < len(code); i++ {
		if code[i] == ' ' || code[i] == '\t' || code[i] == '\n' {
			if cm {
				if tok == "'#" { // end comment
					cm = false
				}
			} else if tok == "#'" { // begin comment
				cm = true
			} else if tok == ")" { // end list
				Assert(Ps_.Cdr() != nil, "Parse WTF! Too many )s")
				Ps_ = Ps_.Cdr()
			} else if len(tok) > 0 {
				var ls Lister
				if tok == "(" { // begin list
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
		f, ok := C_.Car().(Lister)
		Assert(ok, "WTF! Bad stack frame")
		e, ok := f.Car().(Lister)
		if !ok {
			fmt.Print("0 ")
			Ret(f.Car())
		} else {
			switch t := e.Car().(type) {
			case nil:
				Assert(false, "WTF! Can't call the empty list")
			case Inter:
				Assert(false, "WTF! Can't call an int")
			case Lister:
				Assert(false, "WTF! Can't call a list")
				// I kind of like the behavior below, but it causes strange error messages if there's a bug
				/*if f.Cdr() == nil {
					fmt.Println("a")
					C_ = &List{&List{t, nil}, C_}
				} else {
					fmt.Println("b")
					f.SetCar(f.Cdr().Car())
				}*/
			case string:
				switch t {
				case "sx'": // sx', arg, ret
					if f.Cdr() == nil {
						fmt.Print("{0 ")
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.Cdr().Car() != nil {
						fmt.Print("{1 ")
						C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
						f.SetCdr(&List{NCarL(f, 1).Cdr(), nil})
					} else {
						fmt.Print("{2 ")
						Ret(f.Last().Car())
					}
				case "q'":
					fmt.Print("' ")
					Assert(e.Cdr() != nil, "WTF! Missing argument to quote")
					Ret(e.Cdr().Car())
				case ":^'", ":>'", ":|'": // op, ret
					if f.Cdr() == nil {
						fmt.Print(":0 ")
						Assert(e.Cdr() != nil, "WTF! Missing argument to "+t)
						C_ = &List{&List{e.Cdr().Car(), nil}, C_}
					} else if f.Cdr().Car() == nil {
						fmt.Print(":1 ")
						Ret(nil)
					} else {
						fmt.Print(":2 ")
						arg := NCarLA(f, 1, "WTF! "+t+" takes a list")
						switch t {
						case ":^'":
							Ret(arg.Car())
						case ":>'":
							Ret(arg.Cdr())
						case ":|'":
							Ret(arg.Last())
						}
					}
				case "lt'": // lt', arg, ret...
					if f.Cdr() == nil {
						fmt.Print("lt0 ")
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.Cdr().Car() != nil {
						fmt.Print("lt1 ")
						C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
						f.Cdr().SetCar(NCarL(f, 1).Cdr())
					} else {
						fmt.Print("lt2 ")
						SetCdrA(NCdr(f, -2), f.Last().Car(), "WTF! Last argument to lt' must be a list")
						Ret(NCdr(f, 2))
					}
				case "?'":
					// ?', if part, then part, ret
					// ?', then part, nil, ret
					if e.Cdr() == nil {
						fmt.Print("?0 ")
						Ret(nil)
					} else if NCdr(f, 1) == nil {
						fmt.Print("?1 ")
						f.SetCdr(&List{e.Cdr(), &List{e.Cdr().Cdr(), nil}})
					} else if NCdr(f, 3) == nil {
						fmt.Print("?2 ")
						C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
					} else if NCar(f, 2) == nil {
						fmt.Print("?3 ")
						Ret(NCar(f, 3))
					} else if NCar(f, 3) != nil {
						fmt.Print("?4 ")
						f.SetCdr(&List{NCarL(f, 1).Cdr(), &List{nil, nil}})
					} else if NCarL(f, 2).Cdr() == nil {
						fmt.Print("?5 ")
						Ret(nil)
					} else {
						fmt.Print("?6 ")
						f.SetCdr(&List{NCarL(f, 2).Cdr(), &List{NCarL(f, 2).Cdr().Cdr(), nil}})
					}
				case "+'", "*'": // op, arg, sum, ret
					if f.Cdr() == nil {
						fmt.Print("+0 ")
						var bi *big.Int
						if t == "+'" {
							bi = big.NewInt(0)
						} else if t == "*'" {
							bi = big.NewInt(1)
						}
						f.SetCdr(&List{e.Cdr(), &List{bi, nil}})
					} else if NCdr(f, 3) != nil {
						fmt.Print("+1 ")
						if t == "+'" {
							NCarI(f, 2).Add(NCarI(f, 2), NCarIA(f, 3, "WTF! +' takes numbers"))
						} else if t == "*'" {
							NCarI(f, 2).Mul(NCarI(f, 2), NCarIA(f, 3, "WTF! *' takes numbers"))
						}
						f.Cdr().Cdr().SetCdr(nil)
					} else if f.Cdr().Car() != nil {
						fmt.Print("+2 ")
						C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
						f.Cdr().SetCar(NCarL(f, 1).Cdr())
					} else {
						fmt.Print("+3 ")
						Ret(NCar(f, 2))
					}
				case "pr'": // pr', arg, ret...
					if f.Cdr() == nil {
						fmt.Print("pr0 ")
						f.SetCdr(&List{e.Cdr(), nil})
					} else {
						if f.Cdr().Cdr() != nil {
							fmt.Print("pr1 ")
							SetCdrA(NCdr(f, -2), f.Last().Car(), "WTF! pr' takes a string")
						}
						if f.Cdr().Car() != nil {
							fmt.Print("pr2 ")
							C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
							f.Cdr().SetCar(NCarL(f, 1).Cdr())
						} else {
							fmt.Print("pr3 ")
							s := make([]uint8, Len(f.Cdr().Cdr()))
							for i, arg := 0, f.Cdr().Cdr(); arg != nil; i, arg = i+1, arg.Cdr() {
								c, ok := arg.Car().(Inter)
								Assert(ok && c.Cmp(big.NewInt(-1)) == 1 && c.Cmp(big.NewInt(256)) == -1,
									"WTF! Bad byte passed to pr'")
								s[i] = uint8(c.Int64())
							}
							fmt.Print(string(s))
							Ret(f.Cdr().Cdr())
						}
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

func NCar(ls Lister, n int) interface{} {
	nCdr := NCdr(ls, n)
	Assert(nCdr != nil, "WTF! Out of bounds when calling :'")
	return nCdr.Car()
}

func NCarL(ls Lister, n int) Lister {
	return NCarLA(ls, n, "WTF! Requested list element isn't a list")
}

func NCarLA(ls Lister, n int, msg string) Lister {
	nCar, ok := NCar(ls, n).(Lister)
	Assert(ok, msg)
	return nCar
}

func NCarI(ls Lister, n int) *big.Int {
	return NCarIA(ls, n, "WTF! Requested list element isn't an int")
}

func NCarIA(ls Lister, n int, msg string) *big.Int {
	nCar, ok := NCar(ls, n).(*big.Int)
	Assert(ok, msg)
	return nCar
}

func NCdr(ls Lister, n int) Lister {
	if ls == nil {
		return nil
	}
	if n > 0 {
		return NCdr(ls.Cdr(), n-1)
	}
	if n < 0 {
		return NCdr(ls, Len(ls)+n)
	}
	return ls
}

func SetCdrA(ls Lister, v interface{}, msg string) Lister {
	switch t := v.(type) {
	case nil:
		return ls.SetCdr(nil)
	case Lister:
		return ls.SetCdr(t)
	}
	Assert(false, msg)
	return nil
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
