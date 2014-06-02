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
extern int writeCallBack(void*, uint8_t*, int);
extern int64_t seekCallBack(void*, int64_t, int);

*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	IO_BUFFER_SIZE int = 32768
)

// Function prototypes for custom IO. Handlers should implement these prototypes, then instance should be assigned to CustomHandlers global pointer.
//
// E.g.:
// 	func gridFsReader() ([]byte, int) {
// 		... implementation ...
//		return data, length
//	}
//
//	CustomHandlers = &AVIOHandlers{ReadPacket: gridFsReader}
type AVIOHandlers struct {
	ReadPacket  func() ([]byte, int)
	WritePacket func([]byte)
	Seek        func(int64, int) int64
}

// This is a global pointer to AVIOHandler.
var CustomHandlers *AVIOHandlers

type AVIOContext struct {
	avAVIOContext *_Ctype_AVIOContext
	buffer        *C.uchar
}

// AVIOContext constructor. Use it only if You need custom IO behaviour!
func NewAVIOContext(ctx *FmtCtx) (*AVIOContext, error) {
	this := &AVIOContext{}

	this.buffer = (*C.uchar)(C.av_malloc(C.size_t(IO_BUFFER_SIZE)))

	if this.buffer == nil {
		return nil, errors.New("unable to allocate buffer")
	}

	var ptrRead, ptrWrite, ptrSeek *[0]byte = nil, nil, nil

	if CustomHandlers.ReadPacket != nil {
		ptrRead = (*[0]byte)(C.readCallBack)
	}

	if CustomHandlers.WritePacket != nil {
		ptrWrite = (*[0]byte)(C.writeCallBack)
	}

	if CustomHandlers.Seek != nil {
		ptrSeek = (*[0]byte)(C.seekCallBack)
	}

	if this.avAVIOContext = C.avio_alloc_context(this.buffer, C.int(IO_BUFFER_SIZE), 0, unsafe.Pointer(ctx.avCtx), ptrRead, ptrWrite, ptrSeek); this.avAVIOContext == nil {
		return nil, errors.New("unable to initialize avio context")
	}

	return this, nil
}

func (this *AVIOContext) Free() {
	C.av_free(unsafe.Pointer(this.avAVIOContext))
	C.free(unsafe.Pointer(this.buffer))
}

//export readCallBack
func readCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	if CustomHandlers.ReadPacket == nil {
		panic("No reader handler initialized")
	}

	b, n := CustomHandlers.ReadPacket()
	if n >= 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	}

	return C.int(n)
}

//export writeCallBack
func writeCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	if CustomHandlers.WritePacket == nil {
		panic("No writer handler initialized.")
	}
	println("we're here")
	CustomHandlers.WritePacket(C.GoBytes(unsafe.Pointer(buf), buf_size))
	return 0
}

//export seekCallBack
func seekCallBack(opaque unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	if CustomHandlers.Seek == nil {
		panic("No seek handler initialized.")
	}

	return C.int64_t(CustomHandlers.Seek(int64(offset), int(whence)))
}
