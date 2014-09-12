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
	OpLt     struct{}
	OpIf     struct{}
	OpAdd    struct{}
	OpSub    struct{}
	OpMul    struct{}
	OpIntDiv struct{}
	OpPr     struct{}
)

var (
	Ps_      Lister
	C_       Lister
	TempRoot Lister
)

func Parse(code string) {
	TempRoot = &List{OpSx{}, nil}
	Ps_ = &List{TempRoot, nil}
	C_ = &List{&List{TempRoot, nil}, nil}
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
				Assert(Ps_.Cdr() != nil, "Parse WTF! Too many )s")
				if Ps_.Car() != nil {
					_, ok := Ps_.Car().(Lister)
					Assert(ok, "Parse WTF! Unexpected )")
				} else if set, ok := Ps_.Cdr().Car().(*Set); ok {
					(*set)[nil] = true
				}
				Ps_ = Ps_.Cdr()
			} else if tok == "]" {
				Assert(Ps_.Cdr() != nil, "Parse WTF! Too many ]s")
				_, ok := Ps_.Car().(*Set)
				Assert(ok, "Parse WTF! Unexpected ]")
				Ps_ = Ps_.Cdr()
			} else if len(tok) > 0 {
				var a interface{}
				if tok == "(" {
					a = nil
				} else if tok == "[" {
					a = &Set{}
				} else if tok == (OpSx{}.String()) {
					a = OpSx{}
				} else if tok == (OpQ{}.String()) {
					a = OpQ{}
				} else if tok == (OpCar{}.String()) {
					a = OpCar{}
				} else if tok == (OpCdr{}.String()) {
					a = OpCdr{}
				} else if tok == (OpLast{}.String()) {
					a = OpLast{}
				} else if tok == (OpLt{}.String()) {
					a = OpLt{}
				} else if tok == (OpIf{}.String()) {
					a = OpIf{}
				} else if tok == (OpAdd{}.String()) {
					a = OpAdd{}
				} else if tok == (OpSub{}.String()) {
					a = OpSub{}
				} else if tok == (OpMul{}.String()) {
					a = OpMul{}
				} else if tok == (OpIntDiv{}.String()) {
					a = OpIntDiv{}
				} else if tok == (OpPr{}.String()) {
					a = OpPr{}
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
				switch t := Ps_.Car().(type) {
				case nil:
					switch t2 := Ps_.Cdr().Car().(type) {
					case Lister:
						t2.SetCar(ls) // 1st token in list
					case *Set:
						(*t2)[ls] = true
					default:
						Panic("Parse WTF! Bad stack (probably an interpreter bug)")
					}
					Ps_.SetCar(ls)
				case Lister:
					t.SetCdr(ls)
					Ps_.SetCar(ls)
				case *Set:
					if tok != "(" {
						(*t)[a] = true
					}
				default:
					Panic("Parse WTF! Bad stack (probably an interpreter bug)")
				}
				if tok == "(" {
					Ps_ = &List{nil, Ps_}
				} else if tok == "[" {
					Ps_ = &List{a, Ps_}
				}
			}
			tok = ""
		}
	}
	Assert(Ps_.Cdr() == nil, "Parse WTF! Too few )s")
}

func Run() {
	for C_ != nil {
		f, ok := C_.Car().(Lister)
		Assert(ok, "WTF! Bad stack frame")
		switch e := f.Car().(type) {
		case Lister:
			switch t := e.Car().(type) {
			case nil:
				Panic("WTF! Can't call the empty list")
			case Inter:
				Panic("WTF! Can't call an int")
			case Lister:
				Panic("WTF! Can't call a list")
			case *Set:
				Panic("WTF! Can't call a set")
			case OpSx: // op, arg, ret
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
			case OpQ:
				fmt.Print("' ")
				Assert(e.Cdr() != nil, "WTF! Missing argument to quote")
				Ret(e.Cdr().Car())
			case OpCar, OpCdr, OpLast: // op, ret
				if f.Cdr() == nil {
					fmt.Print(":0 ")
					Assert(e.Cdr() != nil, "WTF! Missing argument to "+t.(fmt.Stringer).String())
					C_ = &List{&List{e.Cdr().Car(), nil}, C_}
				} else if f.Cdr().Car() == nil {
					fmt.Print(":1 ")
					Ret(nil)
				} else {
					fmt.Print(":2 ")
					arg := NCarLA(f, 1, "WTF! "+t.(fmt.Stringer).String()+" takes a list")
					switch t.(type) {
					case OpCar:
						Ret(arg.Car())
					case OpCdr:
						Ret(arg.Cdr())
					case OpLast:
						Ret(arg.Last())
					}
				}
			case OpLt: // op, arg, ret...
				if f.Cdr() == nil {
					fmt.Print("lt0 ")
					f.SetCdr(&List{e.Cdr(), nil})
				} else if f.Cdr().Car() != nil {
					fmt.Print("lt1 ")
					C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
					f.Cdr().SetCar(NCarL(f, 1).Cdr())
				} else {
					fmt.Print("lt2 ")
					SetCdrA(NCdr(f, -2), f.Last().Car(), "WTF! Last argument to "+t.String()+" must be a list")
					Ret(NCdr(f, 2))
				}
			case OpIf:
				// op, if part, then part, ret
				// op, then part, nil, ret
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
					f.SetCdr(&List{NCarL(f, 1).Cdr(), &List{}})
				} else if NCarL(f, 2).Cdr() == nil {
					fmt.Print("?5 ")
					Ret(nil)
				} else {
					fmt.Print("?6 ")
					f.SetCdr(&List{NCarL(f, 2).Cdr(), &List{NCarL(f, 2).Cdr().Cdr(), nil}})
				}
			case OpAdd, OpSub, OpMul, OpIntDiv: // op, arg, sum, ret
				if f.Cdr() == nil {
					fmt.Print("+0 ")
					var cdr Lister
					switch t.(type) {
					case OpAdd:
						cdr = &List{big.NewInt(0), nil}
					case OpMul:
						cdr = &List{big.NewInt(1), nil}
					}
					f.SetCdr(&List{e.Cdr(), cdr})
				} else if NCdr(f, 3) != nil {
					fmt.Print("+1 ")
					x := NCarI(f, 2)
					y := NCarIA(f, 3, "WTF! "+t.(fmt.Stringer).String()+" takes numbers")
					switch t.(type) {
					case OpAdd:
						NCarI(f, 2).Add(x, y)
					case OpSub:
						NCdr(f, 2).SetCar(big.NewInt(0).Sub(x, y))
					case OpMul:
						NCarI(f, 2).Mul(x, y)
					case OpIntDiv:
						Assert(y.Sign() != 0, "WTF! Int division by 0")
						// this does Euclidean division (like Python and unlike C), and I like that
						NCdr(f, 2).SetCar(new(big.Int).Div(x, y))
					}
					f.Cdr().Cdr().SetCdr(nil)
				} else if f.Cdr().Car() != nil {
					fmt.Print("+2 ")
					C_ = &List{&List{NCarL(f, 1).Car(), nil}, C_}
					f.Cdr().SetCar(NCarL(f, 1).Cdr())
				} else {
					fmt.Print("+3 ")
					Assert(f.Cdr().Cdr() != nil, "WTF! Missing argument to "+t.(fmt.Stringer).String())
					Ret(NCar(f, 2))
				}
			case OpPr: // op, arg, ret...
				if f.Cdr() == nil {
					fmt.Print("pr0 ")
					f.SetCdr(&List{e.Cdr(), nil})
				} else {
					if f.Cdr().Cdr() != nil {
						fmt.Print("pr1 ")
						SetCdrA(NCdr(f, -2), f.Last().Car(), "WTF! "+t.String()+" takes a string")
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
							Assert(ok && c.Sign() >= 0 && c.Cmp(big.NewInt(256)) == -1,
								"WTF! Bad byte passed to "+t.String())
							s[i] = uint8(c.Int64())
						}
						fmt.Print(string(s))
						Ret(f.Cdr().Cdr())
					}
				}
			case string:
				Panic("WTF! Can't call undefined function \"" + t + "\"")
			default:
				Panic("WTF! Unrecognized function type (probably an interpreter bug)")
			}
		case *Set:
			Panic("TODO: evaluate the set")
		default:
			fmt.Print("0 ")
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
	Panic(msg)
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

func (o OpSx) String() string     { return "sx'" }
func (o OpQ) String() string      { return "q'" }
func (o OpCar) String() string    { return ":^'" }
func (o OpCdr) String() string    { return ":>'" }
func (o OpLast) String() string   { return ":|'" }
func (o OpLt) String() string     { return "lt'" }
func (o OpIf) String() string     { return "?'" }
func (o OpAdd) String() string    { return "+'" }
func (o OpSub) String() string    { return "-'" }
func (o OpMul) String() string    { return "*'" }
func (o OpIntDiv) String() string { return "//'" }
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
