package gmf

/*

#cgo pkg-config: libavformat

#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
#include <string.h>

#include "libavformat/avio.h"

extern int readCallBack(void*, uint8_t*, int);

static AVIOContext *gmf_avio_alloc_context(unsigned char *buffer, int buf_size, void *opaque) {
	fprintf(stderr, "buffer: %p\n", buffer);
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
	IO_BUFFER_SIZE int = 32768
)

var ReaderHandler func()

type AVIOContext struct {
	avAVIOContext *_Ctype_AVIOContext
	buffer        *C.uchar
}

// @todo memory management
func NewAVIOContext(ctx *FmtCtx) (*AVIOContext, error) {
	this := &AVIOContext{}

	this.buffer = (*C.uchar)(C.av_malloc(C.size_t(IO_BUFFER_SIZE)))
	fmt.Println(this.buffer)
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
	var file *os.File
	var err error
	// fmt.Println("readCallBack,buf:", buf, "data:", C.GoBytes(unsafe.Pointer(buf), C.int(buf_size))[:10])

	if section == nil {
		file, err = os.Open("tmp/ref.mp4")
		if err != nil {
			panic(err)
		}

		fi, err := file.Stat()
		if err != nil {
			panic(err)
		}

		section = io.NewSectionReader(file, 0, fi.Size())
	}

	b := make([]byte, int(buf_size))

	n, err := section.Read(b)
	if err != nil {
		fmt.Println(err)
		file.Close()
	}

	C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))

	return C.int(n)
}

func reader() {

}
