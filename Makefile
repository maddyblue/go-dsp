include $(GOROOT)/src/Make.inc

all: install

# Order matters!
DIRS=\
	fft\

install clean nuke:
	for dir in $(DIRS); do \
		$(MAKE) -C $$dir $@ || exit 1; \
	done
