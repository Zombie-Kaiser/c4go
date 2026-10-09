// Feature coverage for the c4go port.
enum { ZERO, ONE, TWO, TEN = 10, ELEVEN };

int g;
char gc;

int fib(int n) {
  if (n < 2) return n;
  return fib(n - 1) + fib(n - 2);
}

int counter;

int main() {
  int i;
  int s;

  // integers, hex, octal
  printf("dec   %d\n", 42);
  printf("hex   %d\n", 0xff);
  printf("oct   %d\n", 010);
  printf("neg   %d\n", -7);

  // enum
  printf("enum  %d %d %d %d %d\n", ZERO, ONE, TWO, TEN, ELEVEN);

  // sizeof and arithmetic precedence
  printf("size  %d %d\n", sizeof(int), sizeof(char));
  printf("prec  %d\n", 2 + 3 * 4 - 6 / 2);
  printf("prec2 %d\n", (2 + 3) * (4 - 6) / 2);

  // bitwise and shifts
  printf("bits  %d %d %d %d %d %d\n", 6 & 3, 6 | 3, 6 ^ 3, 1 << 5, 64 >> 3, ~5);

  // comparisons and logicals
  printf("cmp   %d %d %d %d %d %d\n", 1 < 2, 1 > 2, 1 <= 1, 2 >= 3, 1 == 1, 1 != 1);
  printf("log   %d %d %d\n", 1 && 0, 1 || 0, !0);

  // ternary
  printf("tern  %d\n", 1 ? 10 : 20);
  printf("tern2 %d\n", 0 ? 10 : 20);

  // unary minus and plus
  printf("unary %d %d\n", -5, +5);

  // increment and decrement, prefix and postfix
  counter = 0;
  printf("inc   %d\n", counter++);
  printf("inc2  %d\n", counter);
  printf("dec   %d\n", counter--);
  printf("dec2  %d\n", --counter);
  printf("pinc  %d\n", ++counter);

  // globals
  g = 7;
  printf("glo   %d\n", g);

  // recursion
  printf("fib   %d\n", fib(10));

  // while loop and nested if/else
  i = 0;
  s = 0;
  // (locals must be declared at the top of a block, as in c4)
  while (i < 10) {
    if (i % 2 == 0) s = s + i;
    else s = s - i;
    i++;
  }
  printf("loop  %d\n", s);

  // if without else
  if (s > 0) printf("pos\n");

  // chars and char pointer arithmetic
  gc = 'Z';
  printf("char  %c %d\n", gc, gc);
  printf("esc   %c%c\n", '\n', '\t');

  return 0;
}
