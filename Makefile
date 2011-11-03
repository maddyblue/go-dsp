include $(GOROOT)/src/Make.inc

.PHONY: all install clean nuke fmt

all:
	gomake -C fft

install: all
	gomake -C fft install

clean:
	gomake -C fft clean

nuke:
	gomake -C fft nuke

fmt:
	gomake -C fft fmt
