include $(GOROOT)/src/Make.inc

TARG = pdfdraw
CGOFILES = \
	pdf.go

include $(GOROOT)/src/Make.pkg
