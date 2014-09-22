package anylisp

import (
	"fmt"
	"math/big"
	"os"
	"strings"
)

type (
	Lister interface {
		Car() interface{}
		SetCar(v interface{}) interface{}
		Cdr() Lister
		SetCdr(v Lister) Lister
		Last() Lister
	}

	List struct {
		car interface{}
		cdr Lister
	}

	Set map[interface{}]bool

	Inter interface {
		Add(x, y *big.Int) *big.Int
		Cmp(y *big.Int) (r int)
		Int64() int64
		Mul(x, y *big.Int) *big.Int
		Sign() int
		Sub(x, y *big.Int) *big.Int
	}

	OpSx     struct{}
	OpQ      struct{}
	OpCar    struct{}
	OpCdr    struct{}
	OpLast   struct{}
	OpNCar   struct{}
	OpNCdr   struct{}
	OpSetCar struct{}
	OpSetCdr struct{}
	OpLt     struct{}
	OpSetAdd struct{}
	OpIf     struct{}
	OpAdd    struct{}
	OpSub    struct{}
	OpMul    struct{}
	OpIntDiv struct{}
	OpEval   struct{}
	OpPr     struct{}
)

var (
	P        Lister
	C        Lister
	E        Lister
	TempRoot Lister
)

func Parse(code string) {
	E = &List{&Set{ // TODO: store values directly in cdr
		&List{OpSx{}.String(), &List{OpSx{}, nil}}: true,
		&List{OpQ{}.String(), &List{OpQ{}, nil}}: true,
		&List{OpCar{}.String(), &List{OpCar{}, nil}}: true,
		&List{OpCdr{}.String(), &List{OpCdr{}, nil}}: true,
		&List{OpLast{}.String(), &List{OpLast{}, nil}}: true,
		&List{OpNCar{}.String(), &List{OpNCar{}, nil}}: true,
		&List{OpNCdr{}.String(), &List{OpNCdr{}, nil}}: true,
		&List{OpSetCar{}.String(), &List{OpSetCar{}, nil}}: true,
		&List{OpSetCdr{}.String(), &List{OpSetCdr{}, nil}}: true,
		&List{OpLt{}.String(), &List{OpLt{}, nil}}: true,
		&List{OpSetAdd{}.String(), &List{OpSetAdd{}, nil}}: true,
		&List{OpIf{}.String(), &List{OpIf{}, nil}}: true,
		&List{OpAdd{}.String(), &List{OpAdd{}, nil}}: true,
		&List{OpSub{}.String(), &List{OpSub{}, nil}}: true,
		&List{OpMul{}.String(), &List{OpMul{}, nil}}: true,
		&List{OpIntDiv{}.String(), &List{OpIntDiv{}, nil}}: true,
		&List{OpEval{}.String(), &List{OpEval{}, nil}}: true,
		&List{OpPr{}.String(), &List{OpPr{}, nil}}: true,
	}, nil}
	(*E.Car().(*Set))[&List{"e'", &List{E, nil}}] = true
	TempRoot = &List{OpSx{}, nil}
	P = &List{TempRoot, nil}
	C = &List{&List{TempRoot, nil}, nil}
	tok := ""
	cm := false
	for i := 0; i < len(code); i++ {
		if strings.IndexByte(" \t\n", code[i]) == -1 {
			tok += string(code[i])
		}
		if strings.IndexByte(" \t\n()[]", code[i]) != -1 ||
			(i+1 < len(code) && strings.IndexByte("()[]", code[i+1]) != -1) {
			if cm {
				if tok == "'#" {
					cm = false
				}
			} else if tok == "#'" {
				cm = true
			} else if tok == ")" {
				Assert(P.Cdr() != nil, "Parse WTF! Too many )s")
				if P.Car() != nil {
					_, ok := P.Car().(Lister)
					Assert(ok, "Parse WTF! Unexpected )")
				} else if set, ok := P.Cdr().Car().(*Set); ok {
					(*set)[nil] = true
				}
				P = P.Cdr()
			} else if tok == "]" {
				Assert(P.Cdr() != nil, "Parse WTF! Too many ]s")
				_, ok := P.Car().(*Set)
				Assert(ok, "Parse WTF! Unexpected ]")
				P = P.Cdr()
			} else if len(tok) > 0 {
				var a interface{}
				if tok == "(" {
					a = nil
				} else if tok == "[" {
					a = &Set{}
				} else if tok[0] == '\'' && len(tok) > 1 && func() bool {
					for j := 1; j < len(tok); j++ {
						if !(tok[j] == '-' || (tok[j] >= '0' && tok[j] <= '9') || (tok[j] >= 'a' && tok[j] <= 'f')) {
							return false
						}
					}
					return true
				}() { // number
					bi := new(big.Int)
					_, err := fmt.Sscanf(tok[1:len(tok)], "%x", bi)
					Assert(err == nil, "Parse WTF! Bad number")
					a = bi
				} else { // symbol
					a = tok
				}
				ls := &List{a, nil}
				switch t := P.Car().(type) {
				case nil:
					switch t2 := P.Cdr().Car().(type) {
					case Lister:
						t2.SetCar(ls) // 1st token in list
					case *Set:
						(*t2)[ls] = true
					default:
						Panic("Parse WTF! Bad stack (probably an interpreter bug)")
					}
					P.SetCar(ls)
				case Lister:
					t.SetCdr(ls)
					P.SetCar(ls)
				case *Set:
					if tok != "(" {
						(*t)[a] = true
					}
				default:
					Panic("Parse WTF! Bad stack (probably an interpreter bug)")
				}
				if tok == "(" {
					P = &List{nil, P}
				} else if tok == "[" {
					P = &List{a, P}
				}
			}
			tok = ""
		}
	}
	Assert(P.Cdr() == nil, "Parse WTF! Too few )s")
}

func Run() {
	for C != nil {
		f, ok := C.Car().(Lister)
		Assert(ok, "WTF! Bad stack frame")
		switch e := f.Car().(type) {
		case string:
			fmt.Print("e ")
			v, ok := Lookup(E, e)
			Assert(ok, "WTF! \"" + e + "\" not defined")
			Ret(v)
		case Lister:
			op := e.Car()
			sym, ok := op.(string)
			if ok {
				op, ok = Lookup(E, sym)
				Assert(ok, "WTF! Can't call undefined function \"" + sym + "\"")
			}
			switch t := op.(type) {
			case nil:
				Panic("WTF! Can't call the empty list")
			case Inter:
				Panic("WTF! Can't call an int")
			case Lister:
				Panic("WTF! Can't call a list")
			case *Set:
				Panic("WTF! Can't call a set")
			case fmt.Stringer:
				switch t.(type) {
				case OpSx, OpPr: // op, arg, ret
					if _, ok = t.(OpPr); ok && NCdr(f, 2) != nil {
						s := make([]uint8, Len(NCarLA(f, 2, "WTF! "+t.String()+" takes a string")))
						for i, arg := 0, NCarL(f, 2); arg != nil; i, arg = i+1, arg.Cdr() {
							c, ok := arg.Car().(Inter)
							Assert(ok && c.Sign() >= 0 && c.Cmp(big.NewInt(256)) == -1,
								"WTF! Bad byte passed to "+t.String())
							s[i] = uint8(c.Int64())
						}
						fmt.Print(string(s))
					}
					if f.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.Cdr().Car() != nil {
						fmt.Print(t.String()+"1 ")
						C = &List{&List{NCarL(f, 1).Car(), nil}, C}
						f.SetCdr(&List{NCarL(f, 1).Cdr(), nil})
					} else {
						fmt.Print(t.String()+"2 ")
						Ret(f.Last().Car())
					}
				case OpQ:
					fmt.Print(t.String()+" ")
					Assert(e.Cdr() != nil, "WTF! Missing argument to quote")
					Ret(e.Cdr().Car())
				case OpCar, OpCdr, OpLast: // op, ret
					if f.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						Assert(e.Cdr() != nil, "WTF! Missing argument to "+t.String())
						C = &List{&List{e.Cdr().Car(), nil}, C}
					} else if f.Cdr().Car() == nil {
						fmt.Print(t.String()+"1 ")
						Ret(nil)
					} else {
						fmt.Print(t.String()+"2 ")
						arg := NCarLA(f, 1, "WTF! "+t.String()+" takes a list")
						switch t.(type) {
						case OpCar:
							Ret(arg.Car())
						case OpCdr:
							Ret(arg.Cdr())
						case OpLast:
							Ret(arg.Last())
						}
					}
				case OpSetCar, OpSetCdr: // op, dest, src
					if Len(f) < 3 {
						fmt.Print(t.String()+"0 ")
						Assert(Len(e) > Len(f), fmt.Sprintf("WTF! %s takes 2 arguments but you gave it %d", t.String(), Len(f)-1))
						C = &List{&List{NCar(e, Len(f)), nil}, C}
					} else {
						fmt.Print(t.String()+"1 ")
						x := NCarLA(f, 1, "WTF! 1st argument to "+t.String()+" must be a list")
						switch t.(type) {
						case OpSetCar:
							Ret(x.SetCar(NCar(f, 2)))
						case OpSetCdr:
							Ret(SetCdrA(x, NCar(f, 2), "WTF! 2nd argument to "+t.String()+" must be a list"))
						}
					}
				case OpLt: // op, arg, ret...
					if f.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.Cdr().Car() != nil {
						fmt.Print(t.String()+"1 ")
						C = &List{&List{NCarL(f, 1).Car(), nil}, C}
						f.Cdr().SetCar(NCarL(f, 1).Cdr())
					} else {
						fmt.Print(t.String()+"2 ")
						SetCdrA(NCdr(f, -2), f.Last().Car(), "WTF! Last argument to "+t.String()+" must be a list")
						Ret(NCdr(f, 2))
					}
				case OpIf:
					// op, if part, then part, ret
					// op, then part, nil, ret
					if e.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						Ret(nil)
					} else if NCdr(f, 1) == nil {
						fmt.Print(t.String()+"1 ")
						f.SetCdr(&List{e.Cdr(), &List{e.Cdr().Cdr(), nil}})
					} else if NCdr(f, 3) == nil {
						fmt.Print(t.String()+"2 ")
						C = &List{&List{NCarL(f, 1).Car(), nil}, C}
					} else if NCar(f, 2) == nil {
						fmt.Print(t.String()+"3 ")
						Ret(NCar(f, 3))
					} else if NCar(f, 3) != nil {
						fmt.Print(t.String()+"4 ")
						f.SetCdr(&List{NCarL(f, 1).Cdr(), &List{}})
					} else if NCarL(f, 2).Cdr() == nil {
						fmt.Print(t.String()+"5 ")
						Ret(nil)
					} else {
						fmt.Print(t.String()+"6 ")
						f.SetCdr(&List{NCarL(f, 2).Cdr(), &List{NCarL(f, 2).Cdr().Cdr(), nil}})
					}
				case OpNCar, OpNCdr, OpSetAdd, OpAdd, OpSub, OpMul, OpIntDiv: // op, arg, sum, ret
					if f.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						var cdr Lister
						switch t.(type) {
						case OpAdd:
							cdr = &List{big.NewInt(0), nil}
						case OpMul:
							cdr = &List{big.NewInt(1), nil}
						}
						f.SetCdr(&List{e.Cdr(), cdr})
					} else if NCdr(f, 3) != nil {
						fmt.Print(t.String()+"1 ")
						switch t.(type) {
						case OpNCar:
							x := NCarLA(f, 2, "WTF! "+t.String()+" takes a list")
							y := int(NCarIA(f, 3, "WTF! "+t.String()+" index must be an int").Int64())
							NCdr(f, 2).SetCar(NCar(x, y))
						case OpNCdr:
							x := NCarLA(f, 2, "WTF! "+t.String()+" takes a list")
							y := int(NCarIA(f, 3, "WTF! "+t.String()+" index must be an int").Int64())
							NCdr(f, 2).SetCar(NCdr(x, y))
						case OpSetAdd:
							x := NCarSA(f, 2, "WTF! 1st argument to "+t.String()+" must be a set")
							(*x)[NCar(f, 3)] = true
						default:
							x := NCarIA(f, 2, "WTF! "+t.String()+" takes numbers")
							y := NCarIA(f, 3, "WTF! "+t.String()+" takes numbers")
							switch t.(type) {
							case OpAdd:
								x.Add(x, y)
							case OpSub:
								NCdr(f, 2).SetCar(new(big.Int).Sub(x, y))
							case OpMul:
								x.Mul(x, y)
							case OpIntDiv:
								Assert(y.Sign() != 0, "WTF! Int division by 0")
								// this does Euclidean division (like Python and unlike C), and I like that
								NCdr(f, 2).SetCar(new(big.Int).Div(x, y))
							}
						}
						f.Cdr().Cdr().SetCdr(nil)
					} else if f.Cdr().Car() != nil {
						fmt.Print(t.String()+"2 ")
						C = &List{&List{NCarL(f, 1).Car(), nil}, C}
						f.Cdr().SetCar(NCarL(f, 1).Cdr())
					} else {
						fmt.Print(t.String()+"3 ")
						Assert(f.Cdr().Cdr() != nil, "WTF! Missing argument to "+t.String())
						Ret(NCar(f, 2))
					}
				case OpEval:
					if f.Cdr() == nil {
						fmt.Print(t.String()+"0 ")
						Assert(e.Cdr() != nil, "WTF! Missing argument to "+t.String())
						C = &List{&List{e.Cdr().Car(), nil}, C}
					} else if f.Cdr().Cdr() == nil {
						fmt.Print(t.String()+"1 ")
						C = &List{&List{f.Cdr().Car(), nil}, C}
					} else {
						fmt.Print(t.String()+"2 ")
						Ret(NCar(f, 2))
					}
				default:
					Panic("WTF! Unrecognized function (probably an interpreter bug)")
				}
			default:
				Panic("WTF! Unrecognized function type (probably an interpreter bug)")
			}
		case *Set:
			Panic("TODO: evaluate the set")
		default:
			fmt.Print("r ")
			Ret(f.Car())
		}
	}
}

func PrintTree(ls interface{}) {
	switch t := ls.(type) {
	case nil:
		fmt.Print("()")
	case Inter:
		fmt.Printf("'%x ", t)
	case Lister:
		fmt.Print("(")
		for ls != nil {
			PrintTree(ls.(Lister).Car())
			ls = ls.(Lister).Cdr()
		}
		fmt.Print(")")
	case *Set:
		fmt.Print("[")
		for e := range *t {
			PrintTree(e)
		}
		fmt.Print("]")
	case fmt.Stringer:
		fmt.Print(t.String() + " ")
	case string:
		fmt.Print(t + " ")
	default:
		Panic("Unrecognized object in tree")
	}
}

func Ret(v interface{}) {
	if C.Cdr() != nil {
		C.Cdr().Car().(Lister).Last().SetCdr(&List{v, nil})
	}
	C = C.Cdr()
}

func Lookup(ns interface{}, k string) (interface{}, bool) {
	switch t := ns.(type) {
	case Lister:
		k2, ok := t.Car().(string)
		if ok {
			if k == k2 {
				return t.Cdr().Car(), true
			}
		} else {
			v, ok := Lookup(t.Car(), k)
			if ok {
				return v, true
			}
			if t.Cdr() != nil {
				return Lookup(t.Cdr(), k)
			}
		}
	case *Set:
		for k2 := range *t {
			v, ok := Lookup(k2, k)
			if ok {
				return v, true
			}
		}
	}
	return nil, false
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

func NCarSA(ls Lister, n int, msg string) *Set {
	nCar, ok := NCar(ls, n).(*Set)
	Assert(ok, msg)
	return nCar
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
		n2 := Len(ls)+n
		if n2 < 0 {
			return nil
		}
		return NCdr(ls, n2)
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
	Panic(msg)
	return nil
}

func (ls *List) Car() interface{} {
	return ls.car
}

// TODO: should this return ls or v?
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

func (o OpSx) String() string     { return "sx'" }
func (o OpQ) String() string      { return "q'" }
func (o OpCar) String() string    { return ":^'" }
func (o OpCdr) String() string    { return ":>'" }
func (o OpLast) String() string   { return ":|'" }
func (o OpNCar) String() string   { return ":'" }
func (o OpNCdr) String() string   { return ":@'" }
func (o OpSetCar) String() string { return "=:^'" }
func (o OpSetCdr) String() string { return "=:>'" }
func (o OpLt) String() string     { return "lt'" }
func (o OpSetAdd) String() string { return "$+'" }
func (o OpIf) String() string     { return "?'" }
func (o OpAdd) String() string    { return "+'" }
func (o OpSub) String() string    { return "-'" }
func (o OpMul) String() string    { return "*'" }
func (o OpIntDiv) String() string { return "//'" }
func (o OpEval) String() string   { return "ev'" }
func (o OpPr) String() string     { return "pr'" }

func Assert(cond bool, msg string) {
	if !cond {
		Panic(msg)
	}
}

func Panic(msg string) {
	fmt.Println(msg)
	os.Exit(2)
}
