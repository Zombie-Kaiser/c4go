int addup(int *p, int n) {
  int i;
  int s;
  s = 0;
  i = 0;
  while (i < n) {
    s = s + p[i];
    i++;
  }
  return s;
}

int slen(char *s) {
  int n;
  n = 0;
  while (*s) {
    n++;
    s = s + 1;
  }
  return n;
}

int main() {
  int *arr;
  int i;
  int *j;
  int *p;
  int *q;
  int fd;
  int n;
  char *s;
  char *buf;

  // c4 has no array syntax; a malloc'd block plus a pointer is the idiom
  arr = malloc(80);

  i = 0;
  while (i < 10) {
    arr[i] = i * i;
    i++;
  }
  printf("arr    %d %d %d %d\n", arr[0], arr[1], arr[5], arr[9]);
  printf("sum    %d\n", addup(arr, 10));

  // address-of and pointer deref
  j = &arr[3];
  *j = 999;
  printf("deref  %d\n", arr[3]);
  printf("ptr    %d\n", *j);

  // pointer difference
  p = &arr[9];
  q = &arr[4];
  printf("pdiff  %d\n", p - q);

  // char pointer over a NUL-terminated string
  s = "hello";
  printf("slen   %d\n", slen(s));
  printf("s      %s\n", s);

  // adjacent string literals concatenate
  printf("concat %s\n", "abc" "def");

  // memset / memcmp
  buf = malloc(16);
  memset(buf, 'x', 5);
  buf[5] = 0;
  printf("mem    %s\n", buf);
  printf("mcmp   %d %d\n", memcmp("abc", "abc", 3), memcmp("abc", "abd", 3));

  // open / read / close
  fd = open("test/small.txt", 0);
  if (fd >= 0) {
    n = read(fd, buf, 15);
    buf[n] = 0;
    close(fd);
    printf("file   [%s] %d\n", buf, n);
  }

  free(buf);
  free(arr);
  printf("done\n");
  return 0;
}
