include $(GOROOT)/src/Make.inc

TARG=pdfdraw

GOFILES=\
	pdfdraw.go

# Only enable Poppler backend if we can find it via pkg-config
ifeq ($(shell pkg-config --exists poppler-glib && echo ok),ok)
	CGOFILES += pdfdraw_poppler.go
endif

# Always enable Quartz backend when on Mac OS X
ifeq ($(GOOS),darwin)
	CGOFILES += pdfdraw_quartz.go
endif

include $(GOROOT)/src/Make.pkg
