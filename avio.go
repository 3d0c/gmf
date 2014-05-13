package gmf

/*

#cgo pkg-config: libavformat

#include <stdio.h>
#include <stdlib.h>

#include "libavformat/avio.h"

extern int readCallBack(void*, uint8_t*, int);

static int wrap(void *o, uint8_t *buf, int buf_size) {
	int ret = readCallBack(o, buf, buf_size);

	fprintf(stderr, "ret: %d\ndata:\n%s\n", ret, buf);
	return ret;
}

static AVIOContext *gmf_avio_alloc_context(unsigned char *buffer, int buf_size, void *opaque) {
    return avio_alloc_context(buffer, buf_size, 0, opaque, wrap, NULL, NULL);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"
)

var (
	IO_BUFFER_SIZE int = 32768
)

var ReaderHandler func()

type AVIOContext struct {
	avAVIOContext *_Ctype_AVIOContext
	buffer        *C.uchar
}

func NewAVIOContext(ctx *FmtCtx) (*AVIOContext, error) {
	this := &AVIOContext{}

	this.buffer = (*C.uchar)(C.av_malloc(C.size_t(IO_BUFFER_SIZE)))
	if this.buffer == nil {
		return nil, errors.New("unable to allocate buffer")
	}

	if this.avAVIOContext = C.gmf_avio_alloc_context(this.buffer, C.int(IO_BUFFER_SIZE), unsafe.Pointer(ctx.avCtx)); this.avAVIOContext == nil {
		return nil, errors.New("unable to initialize avio context")
	}

	this.avAVIOContext.opaque = unsafe.Pointer(ctx.avCtx)

	return this, nil
}

var section *io.SectionReader

//export readCallBack
func readCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	// b, n := reader()
	// fmt.Println((*_Ctype_AVIOContext)(opaque).buffer)
	// return C.int(n)
	file, err := os.Open("tmp/ref.mp4")
	if err != nil {
		panic(err)
	}

	if section == nil {
		section = io.NewSectionReader(file, 0, int64(buf_size))
	}

	p := make([]byte, int(buf_size))

	n, err := section.Read(p)
	if err != nil {
		panic(err)
	}

	fmt.Println(n, "bytes read")

	buf = (*C.uint8_t)(unsafe.Pointer(&p[0]))

	return C.int(n)
}

func reader() {

}
