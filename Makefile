CC=$(GOBIN)/8g
LD=$(GOBIN)/8l

all: main
	cp main main.exe

main: main.8 fft.8
	$(LD) -L . -o $@ $<

%.8: %.go
	$(CC) $<
