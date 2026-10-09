int main() {
  char *buf;
  int fd;
  int n;

  buf = malloc(256);
  fd = open("test/small.txt", 0);
  printf("fd %d\n", fd);
  n = read(fd, buf, 100);
  printf("n %d\n", n);
  buf[n] = 0;
  printf("got [%s]\n", buf);
  close(fd);
  return 0;
}
