package gmf

/*

#cgo pkg-config: libavfilter

#include <stdio.h>
#include <libavfilter/buffersink.h>
#include <libavfilter/buffersrc.h>
#include <libavutil/pixdesc.h>

*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	AV_BUFFERSINK_FLAG_PEEK       = 1
	AV_BUFFERSINK_FLAG_NO_REQUEST = 2
	AV_BUFFERSRC_FLAG_PUSH        = 4
)

type Filter struct {
	bufferCtx   []*_Ctype_AVFilterContext
	sinkCtx     *_Ctype_AVFilterContext
	filterGraph *_Ctype_AVFilterGraph
}

func NewFilter(desc string, srcStreams []*Stream, ost *Stream, options []*Option) (*Filter, error) {
	f := &Filter{
		filterGraph: C.avfilter_graph_alloc(),
		bufferCtx:   make([]*_Ctype_AVFilterContext, 0),
	}

	var (
		ret, i  int
		args    string
		inputs  *_Ctype_AVFilterInOut
		outputs *_Ctype_AVFilterInOut
		curr    *_Ctype_AVFilterInOut
		last    *_Ctype_AVFilterContext
	)

	cdesc := C.CString(desc)
	defer C.free(unsafe.Pointer(cdesc))

	if ret = int(C.avfilter_graph_parse2(
		f.filterGraph,
		cdesc,
		&inputs,
		&outputs,
	)); ret < 0 {
		return f, fmt.Errorf("error parsing filter graph - %s", AvError(ret))
	}
	defer C.avfilter_inout_free(&inputs)
	defer C.avfilter_inout_free(&outputs)

	for curr = inputs; curr != nil; curr = curr.next {
		if len(srcStreams) < i {
			return nil, fmt.Errorf("not enough of source streams")
		}

		src := srcStreams[i]

		args = fmt.Sprintf("video_size=%s:pix_fmt=%d:time_base=%s:pixel_aspect=%s:sws_param=flags=%d:frame_rate=%s", src.CodecCtx().GetVideoSize(), src.CodecCtx().PixFmt(), src.TimeBase().AVR(), src.CodecCtx().GetAspectRation().AVR(), SWS_BILINEAR, src.GetRFrameRate().AVR().String())

		if last, ret = f.create("buffer", fmt.Sprintf("in_%d", i), args); ret < 0 {
			return f, fmt.Errorf("error creating input buffer - %s", AvError(ret))
		}

		f.bufferCtx = append(f.bufferCtx, last)

		if ret = int(C.avfilter_link(last, 0, curr.filter_ctx, C.uint(i))); ret < 0 {
			return f, fmt.Errorf("error linking filters - %s", AvError(ret))
		}

		i++
	}

	if f.sinkCtx, ret = f.create("buffersink", "out", ""); ret < 0 {
		return f, fmt.Errorf("error creating filter 'buffersink' - %s", AvError(ret))
	}

	// XXX hardcoded PIXFMT!
	if last, ret = f.create("format", "format", "yuv420p"); ret < 0 {
		return f, fmt.Errorf("error creating format filter - %s", AvError(ret))
	}

	if ret = int(C.avfilter_link(outputs.filter_ctx, 0, last, 0)); ret < 0 {
		return f, fmt.Errorf("error linking output filters - %s", AvError(ret))
	}

	if ret = int(C.avfilter_link(last, 0, f.sinkCtx, 0)); ret < 0 {
		return f, fmt.Errorf("error linking output filters - %s", AvError(ret))
	}

	if ret = int(C.avfilter_graph_config(f.filterGraph, nil)); ret < 0 {
		return f, fmt.Errorf("graph config error - %s", AvError(ret))
	}

	return f, nil
}

func (f *Filter) create(filter, name, args string) (*_Ctype_AVFilterContext, int) {
	var (
		ctx *_Ctype_AVFilterContext
		ret int
	)

	cfilter := C.CString(filter)
	cname := C.CString(name)
	cargs := C.CString(args)

	ret = int(C.avfilter_graph_create_filter(
		&ctx,
		C.avfilter_get_by_name(cfilter),
		cname,
		cargs,
		nil,
		f.filterGraph))

	C.free(unsafe.Pointer(cfilter))
	C.free(unsafe.Pointer(cname))
	C.free(unsafe.Pointer(cargs))

	return ctx, ret
}

func (f *Filter) AddFrame(frame *Frame, istIdx int, flag int) error {
	var ret int

	if istIdx >= len(f.bufferCtx) {
		return fmt.Errorf("unexpected stream index #%d", istIdx)
	}

	if ret = int(C.av_buffersrc_add_frame_flags(
		f.bufferCtx[istIdx],
		frame.avFrame,
		C.int(flag)),
	); ret < 0 {
		return AvError(ret)
	}

	return nil
}

func (f *Filter) Close(istIdx int) error {
	var ret int

	if ret = int(C.av_buffersrc_close(f.bufferCtx[istIdx], 0, AV_BUFFERSRC_FLAG_PUSH)); ret < 0 {
		return AvError(ret)
	}

	return nil
}

func (f *Filter) GetFrame() ([]*Frame, error) {
	var (
		ret    int
		result []*Frame = make([]*Frame, 0)
	)

	for {
		frame := NewFrame()

		ret = int(C.av_buffersink_get_frame_flags(f.sinkCtx, frame.avFrame, AV_BUFFERSINK_FLAG_NO_REQUEST))
		if AvErrno(ret) == syscall.EAGAIN || ret == AVERROR_EOF {
			frame.Free()
			break
		} else if ret < 0 {
			frame.Free()
			return nil, AvError(ret)
		}

		result = append(result, frame)
	}

	f.RequestOldest()

	return result, AvError(ret)
}

func (f *Filter) RequestOldest() error {
	var ret int

	if ret = int(C.avfilter_graph_request_oldest(f.filterGraph)); ret < 0 {
		return AvError(ret)
	}

	return nil
}

func (f *Filter) Dump() {
	fmt.Println(C.GoString(C.avfilter_graph_dump(f.filterGraph, nil)))
}

func (f *Filter) Release() {
	if f.sinkCtx != nil {
		C.avfilter_free(f.sinkCtx)
	}

	for i, _ := range f.bufferCtx {
		C.avfilter_free(f.bufferCtx[i])
	}

	C.avfilter_graph_free(&f.filterGraph)
}
