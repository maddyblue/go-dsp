include $(GOROOT)/src/Make.inc

all: install

# Order matters!
DIRS=\
	dsputils\
	fft\
	window\

install clean nuke:
	for dir in $(DIRS); do \
		$(MAKE) -C $$dir $@ || exit 1; \
	done
