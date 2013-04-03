include $(GOROOT)/src/Make.inc
PKGDIR=$(GOROOT)/pkg/$(GOOS)_$(GOARCH)


TARG=transcoder
GOFILES=transcoder.go

DEPS=\
	gmf

all:

clean: myclean

myclean :
	for d in $(DEPS); do (cd $$d; $(MAKE) clean ); done 

test :
	for d in $(DEPS); do (cd $$d; $(MAKE) test ); done 

include $(GOROOT)/src/Make.cmd


