package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include <stdint.h>
#include <stdlib.h>
#include "libavutil/imgutils.h"
#include "libavutil/mem.h"
#include "libavutil/frame.h"

static uint8_t **alloc_uint4() {
    uint8_t **ptr;

    ptr = (uint8_t **)calloc(4, sizeof(uint8_t));

    return ptr;
}

static int *alloc_int4() {
    int *ptr;

    ptr = (int *)calloc(4, sizeof(int));

    return ptr;
}

static void free_ptr(uint8_t **ptrs, int *lines_ptr) {
    av_freep(&ptrs[0]);
    free(lines_ptr);
}

// static void write_data(uint8_t **ptrs, int bufsize) {
//    fwrite(ptrs[0], 1, bufsize, fp);
// }

static void copy_helper(uint8_t **ptrs, int *linesize, AVFrame *src, int w, int h, int pixfmt) {
    av_image_copy(ptrs, linesize, (const uint8_t **)(src->data), src->linesize,
        pixfmt, w, h);
}

*/
import "C"

//
// UNFINISHED!
//

import (
	"errors"
	"fmt"
	// "unsafe"
)

type Image struct {
	avPointers **C.uint8_t
	avLineSize *C.int
	bufsize    int
	pixFmt     int32
	width      int
	height     int
	CgoMemoryManage
}

// @todo find better way to do allocation
func NewImage(w, h int, pixFmt int32, align int) (*Image, error) {
	this := &Image{
		avPointers: C.alloc_uint4(), // allocate uint8_t *pointers[4]
		avLineSize: C.alloc_int4(),  // allocate int[4]
	}

	ret := C.av_image_alloc(this.avPointers, this.avLineSize, C.int(w), C.int(h), pixFmt, C.int(align))
	if ret < 0 {
		return nil, errors.New(fmt.Sprintf("Unable to allocate image:%s", AvError(int(ret))))
	}

	this.bufsize = int(ret)
	this.pixFmt = pixFmt
	this.width = w
	this.height = h

	return this, nil
}

func (this *Image) Copy(frame *Frame) {
	C.copy_helper(this.avPointers, this.avLineSize, frame.avFrame, C.int(this.width), C.int(this.height), C.int(this.pixFmt))
}

func (this *Image) Free() {
	C.free_ptr(this.avPointers, this.avLineSize)
}
