// c4go - C in four functions, ported to Go
//
// char, int, and pointer types
// if, while, return, and expression statements
// just enough features to allow self-compilation and a bit more
//
// Written by Robert Swierczek
//
//	go build -o c4go .
//	./c4go hello.c
//	./c4go -s hello.c
//	./c4go -d hello.c
//	./c4go c4.c hello.c
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

const tEOF = 0

const (
	Num = 128 + iota
	Fun
	Sys
	Glo
	Loc
	Id
	Char
	Else
	Enum
	If
	Int
	Return
	Sizeof
	While
	Assign
	Cond
	Lor
	Lan
	Or
	Xor
	And
	Eq
	Ne
	Lt
	Gt
	Le
	Ge
	Shl
	Shr
	Add
	Sub
	Mul
	Div
	Mod
	Inc
	Dec
	Brak
)

const (
	LEA = iota
	IMM
	JMP
	JSR
	BZ
	BNZ
	ENT
	ADJ
	LEV
	LI
	LC
	SI
	SC
	PSH
	OR
	XOR
	AND
	EQ
	NE
	LT
	GT
	LE
	GE
	SHL
	SHR
	ADD
	SUB
	MUL
	DIV
	MOD
	OPEN
	READ
	CLOS
	PRTF
	MALC
	FREE
	MSET
	MCMP
	EXIT
)

const opNames = "LEA ,IMM ,JMP ,JSR ,BZ  ,BNZ ,ENT ,ADJ ,LEV ,LI  ,LC  ,SI  ,SC  ,PSH ," +
	"OR  ,XOR ,AND ,EQ  ,NE  ,LT  ,GT  ,LE  ,GE  ,SHL ,SHR ,ADD ,SUB ,MUL ,DIV ,MOD ," +
	"OPEN,READ,CLOS,PRTF,MALC,FREE,MSET,MCMP,EXIT,"

func opName(i int) string { return opNames[i*5 : i*5+4] }

const (
	CHAR = iota
	INT
	PTR
)

const (
	word = 8
	pool = 1 << 20
)

type Sym struct {
	Name                string
	Tk                  int
	Class, Type, Val    int
	HClass, HType, HVal int
}

var (
	src     string
	pos     int
	lp      int
	line    int
	tk      int
	ival    int
	ty      int
	loc     int
	verbose bool
	trace   bool

	code  []int
	le    int
	data  = make([]byte, pool)
	dp    int
	stack = make([]int, pool/word)
	syms  = map[string]*Sym{}
	id    *Sym
	argv  []int
	files []*os.File
	heaps [][]byte
	keep  [][]byte
)

func die(msg string) {
	fmt.Printf("%d: %s\n", line, msg)
	os.Exit(1)
}

func peek() byte {
	if pos < len(src) {
		return src[pos]
	}
	return 0
}

func expect(t int) {
	if tk != t {
		die(fmt.Sprintf("expected '%c'", t))
	}
	next()
}

func emit(v ...int) { code = append(code, v...) }

func jump(op int) int {
	emit(op, 0)
	return len(code) - 1
}

func fix(at int) { code[at] = len(code) }

func top() int     { return code[len(code)-1] }
func setTop(v int) { code[len(code)-1] = v }

func dataAddr(off int) int { return int(uintptr(unsafe.Pointer(&data[0]))) + off }

func addr(i int) int {
	if i < 0 || i >= len(stack) {
		die("stack overflow")
	}
	return int(uintptr(unsafe.Pointer(&stack[0]))) + i*word
}

func raw(p int) unsafe.Pointer { return unsafe.Pointer(uintptr(p)) }

func mem(p, n int) []byte { return unsafe.Slice((*byte)(raw(p)), n) }

func load(p int) int { return *(*int)(raw(p)) }

func store(p, v int) { *(*int)(raw(p)) = v }

func cstr(p int) string {
	b := mem(p, pool)
	n := 0
	for b[n] != 0 {
		n++
	}
	return unsafe.String(&b[0], n)
}

var twoChar = map[string]int{
	"==": Eq, "++": Inc, "--": Dec, "!=": Ne, "<=": Le, "<<": Shl,
	">=": Ge, ">>": Shr, "||": Lor, "&&": Lan,
}

var oneChar = map[byte]int{
	'=': Assign, '+': Add, '-': Sub, '<': Lt, '>': Gt, '|': Or,
	'&': And, '^': Xor, '%': Mod, '*': Mul, '[': Brak, '?': Cond,
}

const punct = "!~;{}()],:"

func isAlpha(c byte) bool { return c == '_' || (c|32) >= 'a' && (c|32) <= 'z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isAlnum(c byte) bool { return isAlpha(c) || isDigit(c) }
func isHex(c byte) bool   { return isDigit(c) || (c|32) >= 'a' && (c|32) <= 'f' }

func hexVal(c byte) int {
	if isDigit(c) {
		return int(c - '0')
	}
	return int(c|32) - 'a' + 10
}

func next() {
	for {
		c := peek()
		if c == 0 {
			tk = tEOF
			return
		}
		pos++
		switch {
		case c == '\n':
			if verbose {
				dump()
			}
			line++
		case c == ' ' || c == '\t' || c == '\r':
		case c == '#':
			for peek() != 0 && peek() != '\n' {
				pos++
			}
		case isAlpha(c):
			st := pos - 1
			for isAlnum(peek()) {
				pos++
			}
			name := src[st:pos]
			s, ok := syms[name]
			if !ok {
				s = &Sym{Name: name, Tk: Id}
				syms[name] = s
			}
			id, tk = s, s.Tk
			return
		case isDigit(c):
			ival = int(c) - '0'
			if ival != 0 {
				for isDigit(peek()) {
					ival = ival*10 + int(peek()) - '0'
					pos++
				}
			} else if peek() == 'x' || peek() == 'X' {
				pos++
				for isHex(peek()) {
					ival = ival*16 + hexVal(peek())
					pos++
				}
			} else {
				for peek() >= '0' && peek() <= '7' {
					ival = ival*8 + int(peek()) - '0'
					pos++
				}
			}
			tk = Num
			return
		case c == '/':
			if peek() == '/' {
				for peek() != 0 && peek() != '\n' {
					pos++
				}
			} else {
				tk = Div
				return
			}
		case c == '"' || c == '\'':
			start := dp
			for peek() != 0 && peek() != c {
				ch := peek()
				pos++
				if ch == '\\' {
					ch = peek()
					pos++
					switch ch {
					case 'n':
						ch = '\n'
					case 't':
						ch = '\t'
					case 'r':
						ch = '\r'
					case '0':
						ch = 0
					}
				}
				ival = int(ch)
				if c == '"' {
					data[dp] = ch
					dp++
				}
			}
			pos++
			if c == '\'' {
				tk = Num
				return
			}

			ival = dataAddr(start)
			tk = '"'
			return
		default:
			if pos < len(src) {
				if t, ok := twoChar[src[pos-1:pos+1]]; ok {
					pos++
					tk = t
					return
				}
			}
			if t, ok := oneChar[c]; ok {
				tk = t
				return
			}
			if strings.IndexByte(punct, c) >= 0 {
				tk = int(c)
				return
			}
			die(fmt.Sprintf("unexpected character '%c'", c))
		}
	}
}

func dump() {
	fmt.Printf("%d: %s", line, src[lp:pos])
	lp = pos
	for le < len(code) {
		i := code[le]
		le++
		fmt.Printf("%8.4s", opName(i))
		if i <= ADJ {
			fmt.Printf(" %d\n", code[le])
			le++
		} else {
			fmt.Println()
		}
	}
}

func expr(lev int) {
	t := 0
	switch tk {
	case Num:
		emit(IMM, ival)
		next()
		ty = INT
	case '"':
		emit(IMM, ival)
		next()
		for tk == '"' {
			next()
		}
		dp = (dp + word - 1) &^ (word - 1)
		ty = PTR
	case Sizeof:
		next()
		expect('(')
		ty = INT
		if tk == Int {
			next()
		} else if tk == Char {
			next()
			ty = CHAR
		}
		for tk == Mul {
			next()
			ty += PTR
		}
		expect(')')
		if ty == CHAR {
			emit(IMM, 1)
		} else {
			emit(IMM, word)
		}
		ty = INT
	case Id:
		d := id
		next()
		switch {
		case tk == '(':
			next()
			n := 0
			for tk != ')' {
				expr(Assign)
				emit(PSH)
				n++
				if tk == ',' {
					next()
				}
			}
			next()
			switch d.Class {
			case Sys:
				emit(d.Val)
			case Fun:
				emit(JSR, d.Val)
			default:
				die("bad function call")
			}
			if n != 0 {
				emit(ADJ, n)
			}
			ty = d.Type
		case d.Class == Num:
			emit(IMM, d.Val)
			ty = INT
		default:
			switch d.Class {
			case Loc:
				emit(LEA, loc-d.Val)
			case Glo:
				emit(IMM, d.Val)
			default:
				die("undefined variable")
			}
			ty = d.Type
			if ty == CHAR {
				emit(LC)
			} else {
				emit(LI)
			}
		}
	case '(':
		next()
		if tk == Int || tk == Char {
			t = INT
			if tk == Char {
				t = CHAR
			}
			next()
			for tk == Mul {
				next()
				t += PTR
			}
			expect(')')
			expr(Inc)
			ty = t
		} else {
			expr(Assign)
			expect(')')
		}
	case Mul:
		next()
		expr(Inc)
		if ty <= INT {
			die("bad dereference")
		}
		ty -= PTR
		if ty == CHAR {
			emit(LC)
		} else {
			emit(LI)
		}
	case And:
		next()
		expr(Inc)
		if top() != LC && top() != LI {
			die("bad address-of")
		}
		code = code[:len(code)-1]
		ty += PTR
	case '!':
		next()
		expr(Inc)
		emit(PSH, IMM, 0, EQ)
		ty = INT
	case '~':
		next()
		expr(Inc)
		emit(PSH, IMM, -1, XOR)
		ty = INT
	case Add:
		next()
		expr(Inc)
		ty = INT
	case Sub:
		next()
		emit(IMM)
		if tk == Num {
			emit(-ival)
			next()
		} else {
			emit(-1, PSH)
			expr(Inc)
			emit(MUL)
		}
		ty = INT
	case Inc, Dec:
		pre := tk
		next()
		expr(Inc)
		switch top() {
		case LC:
			setTop(PSH)
			emit(LC)
		case LI:
			setTop(PSH)
			emit(LI)
		default:
			die("bad lvalue in pre-increment")
		}
		emit(PSH)
		if ty > PTR {
			emit(IMM, word)
		} else {
			emit(IMM, 1)
		}
		if pre == Inc {
			emit(ADD)
		} else {
			emit(SUB)
		}
		if ty == CHAR {
			emit(SC)
		} else {
			emit(SI)
		}
	default:
		die("bad expression")
	}

	for tk >= lev {
		t = ty
		switch tk {
		case Assign:
			next()
			if top() != LC && top() != LI {
				die("bad lvalue in assignment")
			}
			setTop(PSH)
			expr(Assign)
			ty = t
			if ty == CHAR {
				emit(SC)
			} else {
				emit(SI)
			}
		case Cond:
			next()
			b := jump(BZ)
			expr(Assign)
			expect(':')
			code[b] = len(code) + 2
			c := jump(JMP)
			expr(Cond)
			fix(c)
		case Lor:
			next()
			b := jump(BNZ)
			expr(Lan)
			fix(b)
			ty = INT
		case Lan:
			next()
			b := jump(BZ)
			expr(Or)
			fix(b)
			ty = INT
		case Or:
			next()
			emit(PSH)
			expr(Xor)
			emit(OR)
			ty = INT
		case Xor:
			next()
			emit(PSH)
			expr(And)
			emit(XOR)
			ty = INT
		case And:
			next()
			emit(PSH)
			expr(Eq)
			emit(AND)
			ty = INT
		case Eq:
			next()
			emit(PSH)
			expr(Lt)
			emit(EQ)
			ty = INT
		case Ne:
			next()
			emit(PSH)
			expr(Lt)
			emit(NE)
			ty = INT
		case Lt:
			next()
			emit(PSH)
			expr(Shl)
			emit(LT)
			ty = INT
		case Gt:
			next()
			emit(PSH)
			expr(Shl)
			emit(GT)
			ty = INT
		case Le:
			next()
			emit(PSH)
			expr(Shl)
			emit(LE)
			ty = INT
		case Ge:
			next()
			emit(PSH)
			expr(Shl)
			emit(GE)
			ty = INT
		case Shl:
			next()
			emit(PSH)
			expr(Add)
			emit(SHL)
			ty = INT
		case Shr:
			next()
			emit(PSH)
			expr(Add)
			emit(SHR)
			ty = INT
		case Add:
			next()
			emit(PSH)
			expr(Mul)
			if ty = t; ty > PTR {
				emit(PSH, IMM, word, MUL)
			}
			emit(ADD)
		case Sub:
			next()
			emit(PSH)
			expr(Mul)
			if t > PTR && t == ty {
				emit(SUB, PSH, IMM, word, DIV)
				ty = INT
			} else if ty = t; ty > PTR {
				emit(PSH, IMM, word, MUL, SUB)
			} else {
				emit(SUB)
			}
		case Mul:
			next()
			emit(PSH)
			expr(Inc)
			emit(MUL)
			ty = INT
		case Div:
			next()
			emit(PSH)
			expr(Inc)
			emit(DIV)
			ty = INT
		case Mod:
			next()
			emit(PSH)
			expr(Inc)
			emit(MOD)
			ty = INT
		case Inc, Dec:
			switch top() {
			case LC:
				setTop(PSH)
				emit(LC)
			case LI:
				setTop(PSH)
				emit(LI)
			default:
				die("bad lvalue in post-increment")
			}
			sz := 1
			if ty > PTR {
				sz = word
			}
			emit(PSH, IMM, sz)
			if tk == Inc {
				emit(ADD)
			} else {
				emit(SUB)
			}
			if ty == CHAR {
				emit(SC)
			} else {
				emit(SI)
			}
			emit(PSH, IMM, sz)
			if tk == Inc {
				emit(SUB)
			} else {
				emit(ADD)
			}
			next()
		case Brak:
			next()
			emit(PSH)
			expr(Assign)
			expect(']')
			if t > PTR {
				emit(PSH, IMM, word, MUL)
			} else if t < PTR {
				die("pointer type expected")
			}
			emit(ADD)
			ty = t - PTR
			if ty == CHAR {
				emit(LC)
			} else {
				emit(LI)
			}
		default:
			die(fmt.Sprintf("compiler error tk=%d", tk))
		}
	}
}

func stmt() {
	switch tk {
	case If:
		next()
		expect('(')
		expr(Assign)
		expect(')')
		b := jump(BZ)
		stmt()
		if tk == Else {
			code[b] = len(code) + 2
			b = jump(JMP)
			next()
			stmt()
		}
		fix(b)
	case While:
		next()
		topOfLoop := len(code)
		expect('(')
		expr(Assign)
		expect(')')
		b := jump(BZ)
		stmt()
		emit(JMP, topOfLoop)
		fix(b)
	case Return:
		next()
		if tk != ';' {
			expr(Assign)
		}
		emit(LEV)
		expect(';')
	case '{':
		next()
		for tk != '}' {
			stmt()
		}
		next()
	case ';':
		next()
	default:
		expr(Assign)
		expect(';')
	}
}

func run(main *Sym, args []string) {

	argv = make([]int, len(args))
	for i, s := range args {
		b := append([]byte(s), 0)
		keep = append(keep, b)
		argv[i] = int(uintptr(unsafe.Pointer(&b[0])))
	}
	argvPtr := int(uintptr(unsafe.Pointer(&argv[0])))

	emit(PSH, EXIT)
	tramp := len(code) - 2

	pc := main.Val
	bp := len(stack)
	sp := bp
	sp--
	stack[sp] = len(args)
	sp--
	stack[sp] = argvPtr
	sp--
	stack[sp] = tramp

	var a int
	cycle := 0
	for {
		if pc < 0 || pc >= len(code) {
			die("pc out of range")
		}
		i := code[pc]
		pc++
		cycle++
		if trace {
			fmt.Printf("%d> %8.4s", cycle, opName(i))
			if i <= ADJ {
				fmt.Printf(" %d\n", code[pc])
			} else {
				fmt.Println()
			}
		}
		switch i {
		case LEA:
			a = addr(bp + code[pc])
			pc++
		case IMM:
			a = code[pc]
			pc++
		case JMP:
			pc = code[pc]
		case JSR:
			sp--
			stack[sp] = pc + 1
			pc = code[pc]
		case BZ:
			if a == 0 {
				pc = code[pc]
			} else {
				pc++
			}
		case BNZ:
			if a != 0 {
				pc = code[pc]
			} else {
				pc++
			}
		case ENT:
			sp--
			stack[sp] = bp
			bp = sp
			sp -= code[pc]
			pc++
		case ADJ:
			sp += code[pc]
			pc++
		case LEV:
			sp = bp
			bp = stack[sp]
			sp++
			pc = stack[sp]
			sp++
		case LI:
			a = load(a)
		case LC:
			a = int(int8(mem(a, 1)[0]))
		case SI:
			store(stack[sp], a)
			sp++
		case SC:
			*(*int8)(raw(stack[sp])) = int8(byte(a))
			sp++
		case PSH:
			sp--
			stack[sp] = a

		case OR:
			a = stack[sp] | a
			sp++
		case XOR:
			a = stack[sp] ^ a
			sp++
		case AND:
			a = stack[sp] & a
			sp++
		case EQ:
			a = b2i(stack[sp] == a)
			sp++
		case NE:
			a = b2i(stack[sp] != a)
			sp++
		case LT:
			a = b2i(stack[sp] < a)
			sp++
		case GT:
			a = b2i(stack[sp] > a)
			sp++
		case LE:
			a = b2i(stack[sp] <= a)
			sp++
		case GE:
			a = b2i(stack[sp] >= a)
			sp++
		case SHL:
			a = stack[sp] << a
			sp++
		case SHR:
			a = stack[sp] >> a
			sp++
		case ADD:
			a = stack[sp] + a
			sp++
		case SUB:
			a = stack[sp] - a
			sp++
		case MUL:
			a = stack[sp] * a
			sp++
		case DIV:
			a = stack[sp] / a
			sp++
		case MOD:
			a = stack[sp] % a
			sp++

		case OPEN:
			f, err := os.Open(cstr(stack[sp+1]))
			if err != nil {
				a = -1
			} else {
				files = append(files, f)
				a = len(files) - 1
			}
		case READ:
			n, _ := files[stack[sp+2]].Read(mem(stack[sp+1], stack[sp]))
			a = n
		case CLOS:
			files[stack[sp]].Close()
			a = 0
		case PRTF:

			n := code[pc+1]
			f := cstr(stack[sp+n-1])
			use := n - 1
			if v := nverbs(f); v < use {
				use = v
			}
			var fa []any
			for k := 0; k < use; k++ {
				v := stack[sp+n-2-k]
				if verb(f, k) == 's' {
					fa = append(fa, cstr(v))
				} else {
					fa = append(fa, v)
				}
			}
			a, _ = fmt.Printf(f, fa...)
		case MALC:
			b := make([]byte, stack[sp])
			heaps = append(heaps, b)
			a = int(uintptr(unsafe.Pointer(&b[0])))
		case FREE:
			a = 0
		case MSET:
			dst := mem(stack[sp+2], stack[sp])
			for k := range dst {
				dst[k] = byte(stack[sp+1])
			}
			a = stack[sp+2]
		case MCMP:
			a = bytes.Compare(mem(stack[sp+2], stack[sp]), mem(stack[sp+1], stack[sp]))
		case EXIT:
			fmt.Printf("exit(%d) cycle = %d\n", stack[sp], cycle)
			os.Exit(stack[sp])
		default:
			die(fmt.Sprintf("unknown instruction = %d", i))
		}
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nverbs(f string) int {
	n := 0
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			continue
		}
		j := i + 1
		for j < len(f) && strings.IndexByte("-+ #0", f[j]) >= 0 {
			j++
		}
		for j < len(f) && (isDigit(f[j]) || f[j] == '.') {
			j++
		}
		if j >= len(f) {
			break
		}
		if f[j] != '%' {
			n++
		}
		i = j
	}
	return n
}

func verb(f string, k int) byte {
	n := 0
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			continue
		}
		j := i + 1
		for j < len(f) && strings.IndexByte("-+ #0", f[j]) >= 0 {
			j++
		}
		for j < len(f) && (isDigit(f[j]) || f[j] == '.') {
			j++
		}
		if j >= len(f) {
			return 0
		}
		if f[j] == '%' {
			i = j
			continue
		}
		if n == k {
			return f[j]
		}
		n++
		i = j
	}
	return 0
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "-s" {
		verbose = true
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "-d" {
		trace = true
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Println("usage: c4go [-s] [-d] file ...")
		os.Exit(1)
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		die("could not open " + args[0])
	}
	src = string(b)

	code = make([]int, 0, pool/word)
	line = 1

	for i, k := range strings.Fields("char else enum if int return sizeof while") {
		syms[k] = &Sym{Name: k, Tk: Char + i}
	}
	for i, k := range strings.Fields("open read close printf malloc free memset memcmp exit") {
		syms[k] = &Sym{Name: k, Tk: Id, Class: Sys, Type: INT, Val: OPEN + i}
	}

	syms["void"] = &Sym{Name: "void", Tk: Char, Type: CHAR}
	idmain := &Sym{Name: "main", Tk: Id}
	syms["main"] = idmain
	mainOK := false

	next()
	for tk != tEOF {
		bt := INT
		switch tk {
		case Int:
			next()
		case Char:
			next()
			bt = CHAR
		case Enum:
			next()
			if tk != '{' {
				next()
			}
			if tk == '{' {
				next()
				i := 0
				for tk != '}' {
					if tk != Id {
						die("bad enum identifier")
					}
					next()
					if tk == Assign {
						next()
						if tk != Num {
							die("bad enum initializer")
						}
						i = ival
						next()
					}
					id.Class, id.Type, id.Val = Num, INT, i
					i++
					if tk == ',' {
						next()
					}
				}
				next()
			}
		}

		for tk != ';' && tk != '}' {
			ty = bt
			for tk == Mul {
				next()
				ty += PTR
			}
			if tk != Id {
				die("bad global declaration")
			}
			if id.Class != 0 {
				die("duplicate global definition")
			}
			next()
			id.Type = ty

			if tk == '(' {
				id.Class = Fun
				id.Val = len(code)
				if id == idmain {
					mainOK = true
				}
				next()
				i := 0
				for tk != ')' {
					ty = INT
					if tk == Int {
						next()
					} else if tk == Char {
						next()
						ty = CHAR
					}
					for tk == Mul {
						next()
						ty += PTR
					}
					if tk != Id {
						die("bad parameter declaration")
					}
					if id.Class == Loc {
						die("duplicate parameter definition")
					}
					id.HClass, id.Class = id.Class, Loc
					id.HType, id.Type = id.Type, ty
					id.HVal, id.Val = id.Val, i
					i++
					next()
					if tk == ',' {
						next()
					}
				}
				next()
				if tk != '{' {
					die("bad function definition")
				}
				i++
				loc = i
				next()
				for tk == Int || tk == Char {
					if tk == Char {
						bt = CHAR
					} else {
						bt = INT
					}
					next()
					for tk != ';' {
						ty = bt
						for tk == Mul {
							next()
							ty += PTR
						}
						if tk != Id {
							die("bad local declaration")
						}
						if id.Class == Loc {
							die("duplicate local definition")
						}
						id.HClass, id.Class = id.Class, Loc
						id.HType, id.Type = id.Type, ty
						id.HVal, id.Val = id.Val, i+1
						i++
						next()
						if tk == ',' {
							next()
						}
					}
					next()
				}
				emit(ENT, i-loc)
				for tk != '}' {
					stmt()
				}
				emit(LEV)
				for _, s := range syms {
					if s.Class == Loc {
						s.Class, s.Type, s.Val = s.HClass, s.HType, s.HVal
					}
				}
			} else {
				id.Class = Glo
				id.Val = dataAddr(dp)
				dp += word
			}
			if tk == ',' {
				next()
			}
		}
		next()
	}

	if !mainOK {
		die("main() not defined")
	}
	if verbose {
		return
	}
	run(idmain, args)
}
