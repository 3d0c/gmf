package gmf

/*

#cgo pkg-config: libavformat

#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
#include <string.h>

#include "libavformat/avio.h"

extern int readCallBack(void*, uint8_t*, int);

static int wrap(void *o, uint8_t *buf, int buf_size) {
	fprintf(stderr, "buf_size:%d\n", buf_size);
	buf = av_malloc(buf_size);

	int ret = readCallBack(o, buf, buf_size);

	fprintf(stderr, "len:%d\n", ret);

	return ret;
}

static AVIOContext *gmf_avio_alloc_context(unsigned char *buffer, int buf_size, void *opaque) {
    return avio_alloc_context(buffer, buf_size, 0, opaque, readCallBack, NULL, NULL);
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
	IO_BUFFER_SIZE int = 1024
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
	var file *os.File
	var err error

	if section == nil {
		file, err = os.Open("tmp/ref.mp4")
		// file, err := os.Open("./avio.go")
		if err != nil {
			panic(err)
		}

		section = io.NewSectionReader(file, 0, int64(buf_size))
	}

	b := make([]byte, int(buf_size))

	n, err := section.Read(b)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	fmt.Println(n, "bytes, [0-10]:", b[0:10])

	buf = (*C.uint8_t)(unsafe.Pointer(C.av_malloc(C.size_t(n))))
	C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))

	return C.int(n)

}

func reader() {

}
