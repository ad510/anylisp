package anylisp

import (
	"fmt"
	"math/big"
	"os"
	"strings"
)

type (
	List struct {
		car interface{}
		cdr interface{}
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

	Sym string

	Op int
)

const (
	OpSv Op = iota
	OpPr
	OpQ
	OpIf
	OpCar
	OpCdr
	OpLast
	OpSetCar
	OpSetCdr
	OpSetPair
	OpList
	OpLen
	OpLookup
	OpSpawn
	OpNCar
	OpNCdr
	OpSet
	OpSetAdd
	OpSetRm
	OpAdd
	OpSub
	OpMul
	OpIntDiv
	OpEq
	OpNe
	OpLt
	OpGt
	OpLte
	OpGte
	NOp
)

var (
	Names    map[Op]Sym
	P        interface{}
	S        interface{}
	TempRoot interface{}
)

func Init() {
	Names = map[Op]Sym{
		OpSv:      "sv'",
		OpPr:      "pr'",
		OpQ:       "q'",
		OpIf:      "?'",
		OpCar:     ":^'",
		OpCdr:     ":>'",
		OpLast:    ":|'",
		OpSetCar:  "=:^'",
		OpSetCdr:  "=:>'",
		OpSetPair: "=:'",
		OpList:    "lt'",
		OpLen:     "ln'",
		OpLookup:  "lu'",
		OpSpawn:   "ps'",
		OpNCar:    ":'",
		OpNCdr:    ":@'",
		OpSet:     "st'",
		OpSetAdd:  "$+'",
		OpSetRm:   "$-'",
		OpAdd:     "+'",
		OpSub:     "-'",
		OpMul:     "*'",
		OpIntDiv:  "//'",
		OpEq:      "=='",
		OpNe:      "!='",
		OpLt:      "<'",
		OpGt:      ">'",
		OpLte:     "<='",
		OpGte:     ">='",
	}
}

func Parse(code string) {
	TempRoot = &List{}
	P = &List{TempRoot, nil}
	S = &List{func() *Set {
		E := Set{}
		for op := Op(0); op < NOp; op++ {
			name := Names[op]
			o := op
			E[&List{&name, &o}] = true
		}
		return &E
	}(), &List{&List{TempRoot, nil}, nil}}
	sv, _, _ := Lookup(Car(S), Names[OpSv])
	SetCar(TempRoot, sv)
	symS := Sym("s'")
	(*Car(S).(*Set))[&List{&symS, S}] = true
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
				Assert(Cdr(P) != nil, "Parse WTF! Too many )s")
				if Car(P) != nil {
					_, ok := Car(P).(*List)
					Assert(ok, "Parse WTF! Unexpected )")
				} else if set, ok := Car(Cdr(P)).(*Set); ok {
					(*set)[nil] = true
				}
				P = Cdr(P)
			} else if tok == "]" {
				Assert(Cdr(P) != nil, "Parse WTF! Too many ]s")
				_, ok := Car(P).(*Set)
				Assert(ok, "Parse WTF! Unexpected ]")
				P = Cdr(P)
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
					sym := Sym(tok)
					a = &sym
				}
				ls := &List{a, nil}
				switch t := Car(P).(type) {
				case nil:
					switch t2 := Car(Cdr(P)).(type) {
					case *List:
						SetCar(t2, ls) // 1st token in list
					case *Set:
						(*t2)[ls] = true
					default:
						Panic("Parse WTF! Bad stack (probably an interpreter bug)")
					}
					SetCar(P, ls)
				case *List:
					SetCdr(t, ls)
					SetCar(P, ls)
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
	Assert(Cdr(P) == nil, "Parse WTF! Too few )s")
}

func Run() {
	for Cdr(S) != nil {
		E := Car(S)
		C := Cdr(S)
		Ret := func(v interface{}) {
			if C == Cdr(S) {
				if Cdr(C) != nil {
					SetCdr(Last(NCar(C, 1)), &List{v, nil})
				}
				SetCdr(S, Cdr(C))
			}
		}
		f := Car(C)
		switch e := Car(f).(type) {
		case *Sym:
			fmt.Print(string(*e) + " ")
			v, _, ok := Lookup(E, *e)
			Assert(ok, "WTF! \""+string(*e)+"\" not defined")
			Ret(v)
		case *List:
			if Cdr(f) == nil {
				SetCdr(S, &List{&List{Car(e), nil}, C})
			} else if op, ok := Car(Cdr(f)).(*Op); ok {
				f = Cdr(f)
				switch *op {
				case OpSv, OpPr: // op, arg, ret
					if *op == OpPr && NCdr(f, 2) != nil {
						fmt.Print(L2Str(NCar(f, 2), "WTF! "+op.String()+" takes a string"))
					}
					if Cdr(e) == nil {
						Ret(nil)
					} else if Cdr(f) == nil {
						SetCdr(f, &List{Cdr(e), nil})
					} else if Car(Cdr(f)) != nil {
						SetCdr(S, &List{&List{Car(NCar(f, 1)), nil}, C})
						SetCdr(f, &List{Cdr(NCar(f, 1)), nil})
					} else {
						Ret(NCar(f, 2))
					}
				case OpQ:
					Ret(Car(Cdr(e)))
				case OpIf:
					// op, if part, then part, ret
					// op, then part, nil, ret
					if Cdr(e) == nil {
						Ret(nil)
					} else if NCdr(f, 1) == nil {
						SetCdr(f, &List{Cdr(e), &List{Cdr(Cdr(e)), nil}})
					} else if NCdr(f, 3) == nil {
						SetCdr(S, &List{&List{Car(NCar(f, 1)), nil}, C})
					} else if NCar(f, 2) == nil {
						Ret(NCar(f, 3))
					} else if NCar(f, 3) != nil {
						SetCdr(f, &List{Cdr(NCar(f, 1)), &List{}})
					} else if Cdr(NCar(f, 2)) == nil {
						Ret(nil)
					} else {
						SetCdr(f, &List{Cdr(NCar(f, 2)), &List{Cdr(Cdr(NCar(f, 2))), nil}})
					}
				case OpCar, OpCdr, OpLast, OpSetCar, OpSetCdr, OpSetPair, OpList, OpLen, OpLookup, OpSpawn: // op, arg, ret...
					if Cdr(f) == nil {
						SetCdr(f, &List{Cdr(e), nil})
					} else if Car(Cdr(f)) != nil {
						SetCdr(S, &List{&List{Car(NCar(f, 1)), nil}, C})
						SetCar(Cdr(f), Cdr(NCar(f, 1)))
					} else {
						AssertArgs := func(n int64) {
							Assert(Len(f) == n+2, fmt.Sprintf("WTF! %s takes %d arguments but you gave it %d", op.String(), n, Len(f)-2))
						}
						switch *op {
						case OpCar, OpCdr, OpLast:
							AssertArgs(1)
							x := NCar(f, 2)
							switch *op {
							case OpCar:
								Ret(Car(x))
							case OpCdr:
								Ret(Cdr(x))
							case OpLast:
								Ret(Last(x))
							default:
								op.Panic()
							}
						case OpSetCar, OpSetCdr, OpSetPair:
							AssertArgs(2)
							x := NCar(f, 2)
							switch *op {
							case OpSetCar:
								Ret(SetCar(x, NCar(f, 3)))
							case OpSetCdr:
								Ret(SetCdr(x, NCar(f, 3)))
							case OpSetPair:
								y := NCarL(f, 3, "WTF! 2nd argument to "+op.String()+" must be a list")
								Ret(SetCdr(SetCar(x, Car(y)), Cdr(y)))
							default:
								op.Panic()
							}
						case OpList:
							SetCdr(NCdr(f, -2), Car(Last(f)))
							Ret(NCdr(f, 2))
						case OpLen:
							AssertArgs(1)
							switch x := NCar(f, 2).(type) {
							case nil:
								Ret(big.NewInt(0))
							case *List:
								Ret(big.NewInt(Len(x)))
							case *Set:
								Ret(big.NewInt(int64(len(*x))))
							default:
								Panic("WTF! " + op.String() + " takes a list or set")
							}
						case OpLookup:
							AssertArgs(2)
							_, s, ok := Lookup(NCar(f, 2), *NCarSym(f, 3, "WTF! 2nd argument to "+op.String()+" must be a symbol"))
							if ok {
								Ret(s)
							} else {
								Ret(nil)
							}
						case OpSpawn:
							AssertArgs(2)
							name := L2Str(NCar(f, 2), "WTF! 1st argument to "+op.String()+" must be a string")
							m := "WTF! 2nd argument to " + op.String() + " must be a list of strings"
							var argv []string
							if NCar(f, 3) != nil {
								argvl := NCarL(f, 3, m)
								argv = make([]string, Len(argvl))
								for i := 0; i < len(argv); i++ {
									argv[i] = L2Str(NCar(argvl, int64(i)), m)
								}
							}
							p, err := os.StartProcess(name, argv, &os.ProcAttr{}) // TODO: pass in attr
							if err == nil {
								Ret(&List{big.NewInt(int64(p.Pid)), &List{}})
							} else {
								Ret(&List{nil, &List{Str2L(err.Error()), nil}})
							}
						default:
							op.Panic()
						}
					}
				case OpNCar, OpNCdr, OpSet, OpSetAdd, OpSetRm, OpAdd, OpSub, OpMul, OpIntDiv: // op, arg, sum, ret
					if Cdr(f) == nil {
						var cdr interface{}
						switch *op {
						case OpSet:
							cdr = &List{&Set{}, nil}
						case OpAdd:
							cdr = &List{big.NewInt(0), nil}
						case OpMul:
							cdr = &List{big.NewInt(1), nil}
						}
						SetCdr(f, &List{Cdr(e), cdr})
					} else if NCdr(f, 3) != nil {
						switch *op {
						case OpNCar:
							x := NCarL(f, 2, "WTF! "+op.String()+" takes a list")
							y := NCarI(f, 3, "WTF! "+op.String()+" index must be an int").Int64()
							SetCar(NCdr(f, 2), NCar(x, y))
						case OpNCdr:
							x := NCarL(f, 2, "WTF! "+op.String()+" takes a list")
							y := NCarI(f, 3, "WTF! "+op.String()+" index must be an int").Int64()
							SetCar(NCdr(f, 2), NCdr(x, y))
						case OpSet, OpSetAdd:
							x := NCarS(f, 2, "WTF! 1st argument to "+op.String()+" must be a set")
							(*x)[NCar(f, 3)] = true
						case OpSetRm:
							x := NCarS(f, 2, "WTF! 1st argument to "+op.String()+" must be a set")
							delete(*x, NCar(f, 3))
						default:
							x := NCarI(f, 2, "WTF! "+op.String()+" takes numbers")
							y := NCarI(f, 3, "WTF! "+op.String()+" takes numbers")
							switch *op {
							case OpAdd:
								x.Add(x, y)
							case OpSub:
								SetCar(NCdr(f, 2), new(big.Int).Sub(x, y))
							case OpMul:
								x.Mul(x, y)
							case OpIntDiv:
								Assert(y.Sign() != 0, "WTF! Int division by 0")
								// this does Euclidean division (like Python and unlike C), and I like that
								SetCar(NCdr(f, 2), new(big.Int).Div(x, y))
							default:
								op.Panic()
							}
						}
						SetCdr(Cdr(Cdr(f)), nil)
					} else if Car(Cdr(f)) != nil {
						SetCdr(S, &List{&List{Car(NCar(f, 1)), nil}, C})
						SetCar(Cdr(f), Cdr(NCar(f, 1)))
					} else {
						retL, ok := Cdr(Cdr(f)).(*List)
						Assert(ok, "WTF! Missing argument to "+op.String())
						Ret(Car(retL))
					}
				case OpEq, OpNe, OpLt, OpGt, OpLte, OpGte: // op, arg, ret1, ret2
					if Cdr(f) == nil {
						SetCdr(f, &List{Cdr(e), nil})
					} else if *op != OpEq && *op != OpNe && NCdr(f, 2) != nil && NCar(f, 2) == nil {
						Ret(nil)
					} else if NCdr(f, 3) != nil {
						c := NCarI(f, 2, "WTF! "+op.String()+" takes numbers").Cmp(NCarI(f, 3, "WTF! "+op.String()+" takes numbers"))
						var b bool
						switch *op {
						case OpEq:
							b = c == 0
						case OpNe:
							b = c != 0
						case OpLt:
							b = c < 0
						case OpGt:
							b = c > 0
						case OpLte:
							b = c <= 0
						case OpGte:
							b = c >= 0
						default:
							op.Panic()
						}
						if b {
							SetCdr(Cdr(f), NCdr(f, 3))
						} else {
							Ret(nil)
						}
					} else if Car(Cdr(f)) != nil {
						SetCdr(S, &List{&List{Car(NCar(f, 1)), nil}, C})
						SetCar(Cdr(f), Cdr(NCar(f, 1)))
					} else if *op != OpEq && *op != OpNe && Cdr(Cdr(f)) != nil {
						Ret(Car(Last(f)))
					} else {
						Ret(true)
					}
				default:
					op.Panic()
				}
			} else if Cdr(Cdr(f)) == nil { // e, op, ret
				SetCdr(S, &List{&List{Car(Cdr(f)), nil}, C})
			} else {
				Ret(NCar(f, 2))
			}
		case *Set:
			Panic("TODO: evaluate the set")
		default:
			Ret(Car(f))
		}
	}
}

func PrintTree(ls interface{}) {
	switch t := ls.(type) {
	case nil:
		fmt.Print("()")
	case Inter:
		fmt.Printf("'%x ", t)
	case *List:
		fmt.Print("(")
		for ls != nil {
			PrintTree(Car(ls))
			ls = Cdr(ls)
		}
		fmt.Print(")")
	case *Set:
		fmt.Print("[")
		for e := range *t {
			PrintTree(e)
		}
		fmt.Print("]")
	case *Op:
		fmt.Print(t.String() + " ")
	case *Sym:
		fmt.Print(string(*t) + " ")
	default:
		Panic("Unrecognized object in tree")
	}
}

func Lookup(ns interface{}, k Sym) (interface{}, interface{}, bool) {
	switch t := ns.(type) {
	case *List:
		k2, ok := Car(t).(*Sym)
		if ok {
			if k == *k2 {
				return Cdr(t), &List{ns, nil}, true
			}
		} else {
			v, s, ok := Lookup(Car(t), k)
			if ok {
				return v, &List{ns, s}, true
			}
			if Cdr(t) != nil {
				v, s, ok = Lookup(Cdr(t), k)
				return v, &List{ns, s}, ok
			}
		}
	case *Set:
		for k2 := range *t {
			v, s, ok := Lookup(k2, k)
			if ok {
				return v, &List{ns, s}, true
			}
		}
	}
	return nil, nil, false
}

func L2Str(ls interface{}, m string) string {
	_, ok := ls.(*List)
	Assert(ok || ls == nil, m)
	s := make([]uint8, Len(ls))
	for i := 0; ls != nil; i++ {
		c, ok := Car(ls).(Inter)
		Assert(ok && c.Sign() >= 0 && c.Cmp(big.NewInt(256)) == -1, m)
		s[i] = uint8(c.Int64())
		ls = Cdr(ls)
	}
	return string(s)
}

func Str2L(s string) interface{} {
	var f, b interface{}
	for i := 0; i < len(s); i++ {
		e := &List{big.NewInt(int64(s[i])), nil}
		if f == nil {
			f, b = e, e
		} else {
			SetCdr(b, e)
			b = e
		}
	}
	return f
}

func Len(ls interface{}) int64 {
	if ls == nil {
		return 0
	}
	cdr, ok := Cdr(ls).(*List)
	if !ok {
		return 1
	}
	return Len(cdr) + 1
}

func NCar(ls interface{}, n int64) interface{} {
	nCdr, ok := NCdr(ls, n).(*List)
	Assert(ok, "WTF! Out of bounds")
	return Car(nCdr)
}

func NCarL(ls interface{}, n int64, m string) *List {
	nCar, ok := NCar(ls, n).(*List)
	Assert(ok, m)
	return nCar
}

func NCarS(ls interface{}, n int64, m string) *Set {
	nCar, ok := NCar(ls, n).(*Set)
	Assert(ok, m)
	return nCar
}

func NCarSym(ls interface{}, n int64, m string) *Sym {
	nCar, ok := NCar(ls, n).(*Sym)
	Assert(ok, m)
	return nCar
}

func NCarI(ls interface{}, n int64, m string) *big.Int {
	nCar, ok := NCar(ls, n).(*big.Int)
	Assert(ok, m)
	return nCar
}

func NCdr(ls interface{}, n int64) interface{} {
	if ls == nil || n == 0 {
		return ls
	}
	if n < 0 {
		n2 := Len(ls) + n
		if n2 < 0 {
			return nil
		}
		return NCdr(ls, n2)
	}
	return NCdr(Cdr(ls), n-1)
}

func Car(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	ls, ok := v.(*List)
	Assert(ok, "WTF! "+OpCar.String()+" takes a list")
	return ls.car
}

func SetCar(v interface{}, car interface{}) interface{} {
	ls, ok := v.(*List)
	Assert(ok, "WTF! 1st argument to "+OpSetCar.String()+" must be a list")
	ls.car = car
	return ls
}

func Cdr(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	ls, ok := v.(*List)
	Assert(ok, "WTF! "+OpCdr.String()+" takes a list")
	return ls.cdr
}

func SetCdr(v interface{}, cdr interface{}) interface{} {
	ls, ok := v.(*List)
	Assert(ok, "WTF! 1st argument to "+OpSetCdr.String()+" must be a list")
	ls.cdr = cdr
	return ls
}

func Last(v interface{}) interface{} {
	ls, ok := v.(*List)
	Assert(ok, "WTF! "+OpLast.String()+" takes a list")
	if cdr, ok := ls.cdr.(*List); ok {
		return Last(cdr)
	}
	return ls
}

func (op Op) String() string {
	return string(Names[op])
}

func (op Op) Panic() {
	Panic("WTF! Unrecognized function \"" + op.String() + "\" (probably an interpreter bug)")
}

func Assert(b bool, m string) {
	if !b {
		Panic(m)
	}
}

func Panic(m string) {
	fmt.Println(m)
	os.Exit(2)
}
