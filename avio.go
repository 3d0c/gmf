package gmf

/*

#cgo pkg-config: libavformat

#include <stdio.h>
#include <stdlib.h>
#include <inttypes.h>
#include <string.h>

#include "libavformat/avio.h"
#include "libavformat/avformat.h"

extern int readCallBack(void*, uint8_t*, int);

static AVIOContext *gmf_avio_alloc_context(unsigned char *buffer, int buf_size, void *opaque) {
    return avio_alloc_context(buffer, buf_size, 0, opaque, readCallBack, NULL, NULL);
}

*/
import "C"

import (
	"errors"
	// "fmt"
	"unsafe"
)

var (
	IO_BUFFER_SIZE int = 32768
)

type DataHandler struct {
	Reader func() ([]byte, int)
}

type AVIOContext struct {
	avAVIOContext *_Ctype_AVIOContext
	buffer        *C.uchar
}

var ReadHandler *DataHandler

// @todo memory management
func NewAVIOContext(ctx *FmtCtx) (*AVIOContext, error) {
	this := &AVIOContext{}

	this.buffer = (*C.uchar)(C.av_malloc(C.size_t(IO_BUFFER_SIZE)))

	if this.buffer == nil {
		return nil, errors.New("unable to allocate buffer")
	}

	if this.avAVIOContext = C.gmf_avio_alloc_context(this.buffer, C.int(IO_BUFFER_SIZE), unsafe.Pointer(ctx.avCtx)); this.avAVIOContext == nil {
		return nil, errors.New("unable to initialize avio context")
	}

	return this, nil
}

//export readCallBack
func readCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	if ReadHandler == nil {
		panic("No handler initialized")
	}

	b, n := ReadHandler.Reader()
	if n >= 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	}

	return C.int(n)
}

// Unfortunately, there is no way to use per instance reader, e.g.:
// >> NewAVIOContext(ctx, &DataHandler{Reader: customReader})
// because it looks, that there is no place to store user specific
// information. So we have to use some global variables, to store it.
//
// @todo search more
//
// Some way it could look like:
//
/*

// main.go
func customReader() ([]byte, int) {
	...
}

// main.go:
NewAVIOContext(ctx, &DataHandler{Reader: customReader})

// avio.go
func NewAVIOContext(ctx *FmtCtx) (*AVIOContext, error) {
	...
	ctx.avCtx.some_private_ptr = *(*unsafe.Pointer)(unsafe.Pointer(&handler))
	C.gmf_avio_alloc_context(this.buffer, C.int(IO_BUFFER_SIZE), unsafe.Pointer(ctx.avCtx))
}

// avio.go:
//export readCallBack
func readCallBack(ptr unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	b, n := (*DataHandler)(ptr).Reader()

	C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	return C.int(n)
}
*/
