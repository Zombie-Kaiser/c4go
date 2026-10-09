# c4go — c4 in Go

A port of [c4](https://github.com/rswier/c4) (Robert Swierczek's "C in four
functions") to Go. Same algorithm, same bytecode VM, same language surface —
a C compiler that fits in one file and compiles itself.

```
go build -o c4go .
```

## Usage

```
c4go [-s] [-d] file ...
```

| flag | meaning |
|------|---------|
| `-s` | dump each source line interleaved with the code it generated, then exit |
| `-d` | trace every instruction as the program runs |

```sh
./c4go hello.c              # compile and run
./c4go -s hello.c           # show the generated code
./c4go -d hello.c           # single-step the VM
./c4go c4.c hello.c         # use the compiled c4 to compile hello.c
./c4go c4.c c4.c hello.c    # two levels down: still compiles hello
```

The last two are the real test. c4 compiles itself, and so does this.

## Language

Everything c4 supports, and nothing more:

- `char`, `int`, pointers
- `if` / `else`, `while`, `return`
- globals, locals, recursion, `enum`
- full operator set including `? :`, `&&`, `||`, `++`, `--`, `&=` etc.
- `sizeof`, hex/octal/char literals, adjacent string literal concatenation
- `open read close printf malloc free memset memcmp exit`

Deliberate omissions, same as c4: no `for`, `switch`, `do`, `struct`, `union`,
floating point, arrays (`int a[10]` is not syntax — use `malloc` and pointers),
function prototypes, or local declarations that don't sit at the top of a block.

## How it maps to c4.c

| c4.c | c4go |
|------|------|
| `#define int long long` | Go's `int` is already 64-bit on amd64 |
| flat `int *sym` array, offsets `Tk/Hash/Name/...` | `map[string]*Sym` keyed by name |
| hashed identifiers into `tk` | real strings compared by Go |
| `(int)pp` casts to fake a pointer type | `int(uintptr(unsafe.Pointer(p)))` |
| `--argc; ++argv` | `os.Args[1:]` |
| `printf` with 6 always-passed args | `fmt.Printf`, arity matched to the verbs |

The `unsafe` usage is unavoidable: this is a VM that hands raw addresses to
guest programs, so it needs real pointers. `unsafe` appears in about six
places, all in the address helpers at the bottom of the state block.

## Layout

```
main.go          the whole compiler: lexer, expression parser, stmt, VM
test/features.c  every operator and literal form
test/pointers.c  malloc, pointer arithmetic, memcmp, file IO
test/concat.c    adjacent string literals
test/esc.c       escape sequences
test/fileio.c    open/read/close
test/small.txt   input for the file IO tests
```

## Caveats

- `%s` works by inspecting each argument's verb in the format string, since
  every value arrives in the VM as an `int` and only `%s` ones are really
  pointers. Widths like `%8.4s` parse; more exotic specifiers do not.
- `free` is a no-op. Go's GC owns the block.
- File descriptors are indices into a table, not OS fds.
- Fixed 1 MB data area and 1 MB stack. Bigger inputs need `pool` raised.
