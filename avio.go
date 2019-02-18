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
	"fmt"
	"unsafe"
)

const (
	AVIO_FLAG_READ       = 1
	AVIO_FLAG_WRITE      = 2
	AVIO_FLAG_READ_WRITE = (AVIO_FLAG_READ | AVIO_FLAG_WRITE)
)

var (
	IO_BUFFER_SIZE int = 32768
)

// Functions prototypes for custom IO. Implement necessary prototypes and pass instance pointer to NewAVIOContext.
//
// E.g.:
// 	func gridFsReader() ([]byte, int) {
// 		... implementation ...
//		return data, length
//	}
//
//	avoictx := NewAVIOContext(ctx, &AVIOHandlers{ReadPacket: gridFsReader})
type AVIOHandlers struct {
	ReadPacket  func() ([]byte, int)
	WritePacket func([]byte) int
	Seek        func(int64, int) int64
}

// Global map of AVIOHandlers
// one handlers struct per format context. Using ctx.avCtx pointer address as a key.
var handlersMap map[uintptr]*AVIOHandlers

type AVIOContext struct {
	avAVIOContext *_Ctype_AVIOContext
	// avAVIOContext *C.struct_AVIOContext
	handlerKey uintptr
	CgoMemoryManage
}

// NewAVIOContext constructor. Use it only if You need custom IO behaviour!
func NewAVIOContext(ctx *FmtCtx, handlers *AVIOHandlers, size ...int) (*AVIOContext, error) {
	this := &AVIOContext{}

	bufferSize := IO_BUFFER_SIZE

	if len(size) == 1 {
		bufferSize = size[0]
	}

	buffer := (*C.uchar)(C.av_malloc(C.size_t(bufferSize)))

	if buffer == nil {
		return nil, errors.New("unable to allocate buffer")
	}

	// we have to explicitly set it to nil, to force library using default handlers
	var ptrRead, ptrWrite, ptrSeek *[0]byte = nil, nil, nil

	if handlers != nil {
		if handlersMap == nil {
			handlersMap = make(map[uintptr]*AVIOHandlers)
		}

		handlersMap[uintptr(unsafe.Pointer(ctx.avCtx))] = handlers
		this.handlerKey = uintptr(unsafe.Pointer(ctx.avCtx))
	}

	var flag int = 0

	if handlers.ReadPacket != nil {
		ptrRead = (*[0]byte)(C.readCallBack)
		flag = 0
	}

	if handlers.WritePacket != nil {
		ptrWrite = (*[0]byte)(C.writeCallBack)
		flag = AVIO_FLAG_WRITE
	}

	if handlers.Seek != nil {
		ptrSeek = (*[0]byte)(C.seekCallBack)
	}

	if handlers.ReadPacket != nil && handlers.WritePacket != nil {
		flag = AVIO_FLAG_READ_WRITE
	}

	if this.avAVIOContext = C.avio_alloc_context(buffer, C.int(bufferSize), C.int(flag), unsafe.Pointer(ctx.avCtx), ptrRead, ptrWrite, ptrSeek); this.avAVIOContext == nil {
		C.av_free(unsafe.Pointer(this.avAVIOContext.buffer))
		return nil, errors.New("unable to initialize avio context")
	}

	this.avAVIOContext.min_packet_size = C.int(bufferSize)

	return this, nil
}

func (this *AVIOContext) Free() {
	delete(handlersMap, this.handlerKey)
	C.av_free(unsafe.Pointer(this.avAVIOContext.buffer))
	C.av_free(unsafe.Pointer(this.avAVIOContext))
}

func (this *AVIOContext) Flush() {
	C.avio_flush(this.avAVIOContext)
}

//export readCallBack
func readCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.ReadPacket == nil {
		panic("No reader handler initialized")
	}

	b, n := handlers.ReadPacket()
	if n > 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&b[0]), C.size_t(n))
	}

	return C.int(n)
}

//export writeCallBack
func writeCallBack(opaque unsafe.Pointer, buf *C.uint8_t, buf_size C.int) C.int {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.WritePacket == nil {
		panic("No writer handler initialized.")
	}

	return C.int(handlers.WritePacket(C.GoBytes(unsafe.Pointer(buf), buf_size)))
}

//export seekCallBack
func seekCallBack(opaque unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	handlers, found := handlersMap[uintptr(opaque)]
	if !found {
		panic(fmt.Sprintf("No handlers instance found, according pointer: %v", opaque))
	}

	if handlers.Seek == nil {
		panic("No seek handler initialized.")
	}

	return C.int64_t(handlers.Seek(int64(offset), int(whence)))
}
