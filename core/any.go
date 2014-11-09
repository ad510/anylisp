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
		SetCar(v interface{}) Lister
		Cdr() interface{}
		CdrL() Lister
		CdrLA(m string) Lister
		SetCdr(v interface{}) Lister
		Last() Lister
	}

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
	P        Lister
	S        Lister
	TempRoot Lister
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
	sv, _, _ := Lookup(S.Car(), Names[OpSv])
	TempRoot.SetCar(sv)
	symS := Sym("s'")
	(*S.Car().(*Set))[&List{&symS, S}] = true
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
				} else if set, ok := P.CdrL().Car().(*Set); ok {
					(*set)[nil] = true
				}
				P = P.CdrL()
			} else if tok == "]" {
				Assert(P.Cdr() != nil, "Parse WTF! Too many ]s")
				_, ok := P.Car().(*Set)
				Assert(ok, "Parse WTF! Unexpected ]")
				P = P.CdrL()
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
				switch t := P.Car().(type) {
				case nil:
					switch t2 := P.CdrL().Car().(type) {
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
	for S.Cdr() != nil {
		E := S.Car()
		C := S.CdrLA("WTF! Bad stack")
		Ret := func(v interface{}) {
			if C == S.Cdr() {
				if C.Cdr() != nil {
					NCarL(C, 1).Last().SetCdr(&List{v, nil})
				}
				S.SetCdr(C.Cdr())
			}
		}
		f, ok := C.Car().(Lister)
		Assert(ok, "WTF! Bad stack frame")
		switch e := f.Car().(type) {
		case *Sym:
			fmt.Print(string(*e) + " ")
			v, _, ok := Lookup(E, *e)
			Assert(ok, "WTF! \""+string(*e)+"\" not defined")
			Ret(v)
		case Lister:
			if f.Cdr() == nil {
				S.SetCdr(&List{&List{e.Car(), nil}, C})
			} else if op, ok := f.CdrL().Car().(*Op); ok {
				f = f.CdrL()
				switch *op {
				case OpSv, OpPr: // op, arg, ret
					if *op == OpPr && NCdr(f, 2) != nil {
						fmt.Print(L2Str(NCar(f, 2), "WTF! "+op.String()+" takes a string"))
					}
					if e.Cdr() == nil {
						Ret(nil)
					} else if f.Cdr() == nil {
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.CdrL().Car() != nil {
						S.SetCdr(&List{&List{NCarL(f, 1).Car(), nil}, C})
						f.SetCdr(&List{NCarL(f, 1).Cdr(), nil})
					} else {
						Ret(NCar(f, 2))
					}
				case OpQ:
					Ret(e.CdrLA("WTF! Missing argument to quote").Car())
				case OpIf:
					// op, if part, then part, ret
					// op, then part, nil, ret
					if e.Cdr() == nil {
						Ret(nil)
					} else if NCdr(f, 1) == nil {
						f.SetCdr(&List{e.Cdr(), &List{e.CdrL().Cdr(), nil}})
					} else if NCdr(f, 3) == nil {
						S.SetCdr(&List{&List{NCarL(f, 1).Car(), nil}, C})
					} else if NCar(f, 2) == nil {
						Ret(NCar(f, 3))
					} else if NCar(f, 3) != nil {
						f.SetCdr(&List{NCarL(f, 1).Cdr(), &List{}})
					} else if NCarL(f, 2).Cdr() == nil {
						Ret(nil)
					} else {
						f.SetCdr(&List{NCarL(f, 2).Cdr(), &List{NCarL(f, 2).CdrL().Cdr(), nil}})
					}
				case OpCar, OpCdr, OpLast, OpSetCar, OpSetCdr, OpSetPair, OpList, OpLen, OpLookup, OpSpawn: // op, arg, ret...
					if f.Cdr() == nil {
						f.SetCdr(&List{e.Cdr(), nil})
					} else if f.CdrL().Car() != nil {
						S.SetCdr(&List{&List{NCarL(f, 1).Car(), nil}, C})
						f.CdrL().SetCar(NCarL(f, 1).Cdr())
					} else {
						AssertArgs := func(n int64) {
							Assert(Len(f) == n+2, fmt.Sprintf("WTF! %s takes %d arguments but you gave it %d", op.String(), n, Len(f)-2))
						}
						switch *op {
						case OpCar, OpCdr, OpLast:
							AssertArgs(1)
							if NCar(f, 2) == nil {
								Ret(nil)
							} else {
								arg := NCarLA(f, 2, "WTF! "+op.String()+" takes a list")
								switch *op {
								case OpCar:
									Ret(arg.Car())
								case OpCdr:
									Ret(arg.Cdr())
								case OpLast:
									Ret(arg.Last())
								default:
									op.Panic()
								}
							}
						case OpSetCar, OpSetCdr, OpSetPair:
							AssertArgs(2)
							x := NCarLA(f, 2, "WTF! 1st argument to "+op.String()+" must be a list")
							switch *op {
							case OpSetCar:
								Ret(x.SetCar(NCar(f, 3)))
							case OpSetCdr:
								Ret(x.SetCdr(NCar(f, 3)))
							case OpSetPair:
								y := NCarLA(f, 3, "WTF! 2nd argument to "+op.String()+" must be a list")
								Ret(x.SetCar(y.Car()).SetCdr(y.Cdr()))
							default:
								op.Panic()
							}
						case OpList:
							NCdrL(f, -2).SetCdr(f.Last().Car())
							Ret(NCdr(f, 2))
						case OpLen:
							AssertArgs(1)
							switch arg := NCar(f, 2).(type) {
							case nil:
								Ret(big.NewInt(0))
							case Lister:
								Ret(big.NewInt(Len(arg)))
							case *Set:
								Ret(big.NewInt(int64(len(*arg))))
							default:
								Panic("WTF! " + op.String() + " takes a list or set")
							}
						case OpLookup:
							AssertArgs(2)
							_, s, ok := Lookup(NCar(f, 2), *NCarSymA(f, 3, "WTF! 2nd argument to "+op.String()+" must be a symbol"))
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
								argvl := NCarLA(f, 3, m)
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
					if f.Cdr() == nil {
						var cdr Lister
						switch *op {
						case OpSet:
							cdr = &List{&Set{}, nil}
						case OpAdd:
							cdr = &List{big.NewInt(0), nil}
						case OpMul:
							cdr = &List{big.NewInt(1), nil}
						}
						f.SetCdr(&List{e.Cdr(), cdr})
					} else if NCdr(f, 3) != nil {
						switch *op {
						case OpNCar:
							x := NCarLA(f, 2, "WTF! "+op.String()+" takes a list")
							y := NCarIA(f, 3, "WTF! "+op.String()+" index must be an int").Int64()
							NCdrL(f, 2).SetCar(NCar(x, y))
						case OpNCdr:
							x := NCarLA(f, 2, "WTF! "+op.String()+" takes a list")
							y := NCarIA(f, 3, "WTF! "+op.String()+" index must be an int").Int64()
							NCdrL(f, 2).SetCar(NCdr(x, y))
						case OpSet, OpSetAdd:
							x := NCarSA(f, 2, "WTF! 1st argument to "+op.String()+" must be a set")
							(*x)[NCar(f, 3)] = true
						case OpSetRm:
							x := NCarSA(f, 2, "WTF! 1st argument to "+op.String()+" must be a set")
							delete(*x, NCar(f, 3))
						default:
							x := NCarIA(f, 2, "WTF! "+op.String()+" takes numbers")
							y := NCarIA(f, 3, "WTF! "+op.String()+" takes numbers")
							switch *op {
							case OpAdd:
								x.Add(x, y)
							case OpSub:
								NCdrL(f, 2).SetCar(new(big.Int).Sub(x, y))
							case OpMul:
								x.Mul(x, y)
							case OpIntDiv:
								Assert(y.Sign() != 0, "WTF! Int division by 0")
								// this does Euclidean division (like Python and unlike C), and I like that
								NCdrL(f, 2).SetCar(new(big.Int).Div(x, y))
							default:
								op.Panic()
							}
						}
						f.CdrL().CdrL().SetCdr(nil)
					} else if f.CdrL().Car() != nil {
						S.SetCdr(&List{&List{NCarL(f, 1).Car(), nil}, C})
						f.CdrL().SetCar(NCarL(f, 1).Cdr())
					} else {
						retL, ok := f.CdrL().Cdr().(Lister)
						Assert(ok, "WTF! Missing argument to "+op.String())
						Ret(retL.Car())
					}
				case OpEq, OpNe, OpLt, OpGt, OpLte, OpGte: // op, arg, ret1, ret2
					if f.Cdr() == nil {
						f.SetCdr(&List{e.Cdr(), nil})
					} else if *op != OpEq && *op != OpNe && NCdr(f, 2) != nil && NCar(f, 2) == nil {
						Ret(nil)
					} else if NCdr(f, 3) != nil {
						c := NCarIA(f, 2, "WTF! "+op.String()+" takes numbers").Cmp(NCarIA(f, 3, "WTF! "+op.String()+" takes numbers"))
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
							f.CdrL().SetCdr(NCdr(f, 3))
						} else {
							Ret(nil)
						}
					} else if f.CdrL().Car() != nil {
						S.SetCdr(&List{&List{NCarL(f, 1).Car(), nil}, C})
						f.CdrL().SetCar(NCarL(f, 1).Cdr())
					} else if *op != OpEq && *op != OpNe && f.CdrL().Cdr() != nil {
						Ret(f.Last().Car())
					} else {
						Ret(true)
					}
				default:
					op.Panic()
				}
			} else if f.CdrL().Cdr() == nil { // e, op, ret
				S.SetCdr(&List{&List{f.CdrL().Car(), nil}, C})
			} else {
				Ret(NCar(f, 2))
			}
		case *Set:
			Panic("TODO: evaluate the set")
		default:
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
	case *Op:
		fmt.Print(t.String() + " ")
	case *Sym:
		fmt.Print(string(*t) + " ")
	default:
		Panic("Unrecognized object in tree")
	}
}

func Lookup(ns interface{}, k Sym) (interface{}, Lister, bool) {
	switch t := ns.(type) {
	case Lister:
		k2, ok := t.Car().(*Sym)
		if ok {
			if k == *k2 {
				return t.Cdr(), &List{ns, nil}, true
			}
		} else {
			v, s, ok := Lookup(t.Car(), k)
			if ok {
				return v, &List{ns, s}, true
			}
			if t.Cdr() != nil {
				v, s, ok = Lookup(t.Cdr(), k)
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

func L2Str(v interface{}, m string) string {
	ls, _ := v.(Lister)
	s := make([]uint8, Len(ls))
	for i := 0; ls != nil; i++ {
		c, ok := ls.Car().(Inter)
		Assert(ok && c.Sign() >= 0 && c.Cmp(big.NewInt(256)) == -1, m)
		s[i] = uint8(c.Int64())
		ls, _ = ls.Cdr().(Lister)
	}
	return string(s)
}

func Str2L(s string) Lister {
	var f, b Lister
	for i := 0; i < len(s); i++ {
		e := &List{big.NewInt(int64(s[i])), nil}
		if f == nil {
			f, b = e, e
		} else {
			b.SetCdr(e)
			b = e
		}
	}
	return f
}

func Len(ls Lister) int64 {
	if ls == nil {
		return 0
	}
	cdr, ok := ls.Cdr().(Lister)
	if !ok {
		return 1
	}
	return Len(cdr) + 1
}

func NCar(ls Lister, n int64) interface{} {
	return NCdrL(ls, n).Car()
}

func NCarL(ls Lister, n int64) Lister {
	return NCarLA(ls, n, "WTF! Requested list element isn't a list")
}

func NCarLA(ls Lister, n int64, m string) Lister {
	nCar, ok := NCar(ls, n).(Lister)
	Assert(ok, m)
	return nCar
}

func NCarSA(ls Lister, n int64, m string) *Set {
	nCar, ok := NCar(ls, n).(*Set)
	Assert(ok, m)
	return nCar
}

func NCarSymA(ls Lister, n int64, m string) *Sym {
	nCar, ok := NCar(ls, n).(*Sym)
	Assert(ok, m)
	return nCar
}

func NCarIA(ls Lister, n int64, m string) *big.Int {
	nCar, ok := NCar(ls, n).(*big.Int)
	Assert(ok, m)
	return nCar
}

func NCdr(v interface{}, n int64) interface{} {
	if v == nil || n == 0 {
		return v
	}
	ls, ok := v.(Lister)
	Assert(ok, "WTF! Bad rest pointer in list")
	if n < 0 {
		n2 := Len(ls) + n
		if n2 < 0 {
			return nil
		}
		return NCdr(ls, n2)
	}
	return NCdr(ls.Cdr(), n-1)
}

func NCdrL(v interface{}, n int64) Lister {
	nCdr, ok := NCdr(v, n).(Lister)
	Assert(ok, "WTF! Out of bounds")
	return nCdr
}

func (ls *List) Car() interface{} {
	return ls.car
}

func (ls *List) SetCar(v interface{}) Lister {
	ls.car = v
	return ls
}

func (ls *List) Cdr() interface{} {
	return ls.cdr
}

func (ls *List) CdrL() Lister {
	return ls.CdrLA("WTF! Rest pointer doesn't point to a list")
}

func (ls *List) CdrLA(m string) Lister {
	cdr, ok := ls.cdr.(Lister)
	Assert(ok, m)
	return cdr
}

func (ls *List) SetCdr(v interface{}) Lister {
	ls.cdr = v
	return ls
}

func (ls *List) Last() Lister {
	cdr, ok := ls.cdr.(Lister)
	if ok {
		return cdr.Last()
	}
	return ls
}

func (op *Op) String() string {
	return string(Names[*op])
}

func (op *Op) Panic() {
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
