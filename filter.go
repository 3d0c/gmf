package gmf

/*

#cgo pkg-config: libavfilter

#include <stdio.h>
#include <libavfilter/buffersink.h>
#include <libavfilter/buffersrc.h>
#include <libavutil/pixdesc.h>

int gmf_create_filter(AVFilterContext **filt_ctx, const AVFilter *filt, const char *name, const char *args, void *opaque, AVFilterGraph *graph_ctx, int i) {
	return avfilter_graph_create_filter(&filt_ctx[i], filt, name, args, opaque, graph_ctx);
}

int gmf_av_buffersrc_add_frame_flags(AVFilterContext **filt_ctx, AVFrame *frame, int flags, int i) {
	return av_buffersrc_add_frame_flags(filt_ctx[i], frame, flags);
}

AVFilterContext *gmf_get_current(AVFilterContext **filt_ctx, int i) {
	return filt_ctx[i];
}

*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	AV_BUFFERSINK_FLAG_NO_REQUEST = 2
	AV_BUFFERSRC_FLAG_PUSH        = 4
)

type Filter struct {
	bufferCtx   **_Ctype_AVFilterContext
	sinkCtx     *_Ctype_AVFilterContext
	filterGraph *_Ctype_AVFilterGraph
	bufferCtxNb int
}

func NewFilter(desc string, srcStreams []*Stream, ost *Stream, options []*Option) (*Filter, error) {
	f := &Filter{}

	var (
		ret, i  int
		args    string
		inputs  *_Ctype_AVFilterInOut
		outputs *_Ctype_AVFilterInOut
		curr    *_Ctype_AVFilterInOut
		last    *_Ctype_AVFilterContext
		format  *_Ctype_AVFilterContext
	)

	cnameOut := C.CString("out")
	defer C.free(unsafe.Pointer(cnameOut))

	f.filterGraph = C.avfilter_graph_alloc()

	f.bufferCtxNb = len(srcStreams)
	csz := C.ulong(f.bufferCtxNb)

	f.bufferCtx = (**_Ctype_struct_AVFilterContext)(C.av_calloc(csz, C.sizeof_AVFilterContext))

	cdesc := C.CString(desc)
	defer C.free(unsafe.Pointer(cdesc))

	if ret = int(C.avfilter_graph_parse2(
		f.filterGraph,
		cdesc,
		&inputs,
		&outputs,
	)); ret < 0 {
		return nil, fmt.Errorf("error parsing filter graph - %s", AvError(ret))
	}

	for curr = inputs; curr != nil; curr = curr.next {
		if len(srcStreams) < i {
			return nil, fmt.Errorf("not enough of source streams")
		}
		srcStream := srcStreams[i]

		args = fmt.Sprintf("video_size=%s:pix_fmt=%d:time_base=%s:pixel_aspect=%s:sws_param=flags=%d", srcStream.CodecCtx().GetVideoSize(), srcStream.CodecCtx().PixFmt(), srcStream.TimeBase().AVR(), srcStream.CodecCtx().GetAspectRation().AVR(), SWS_BILINEAR)

		cargs := C.CString(args)

		ci := C.int(i)

		name := fmt.Sprintf("in_%d", i)
		cname := C.CString(name)

		fmt.Printf("args: %s\n", args)

		if ret = int(C.gmf_create_filter(
			f.bufferCtx,
			C.avfilter_get_by_name(C.CString("buffer")),
			cname,
			cargs,
			nil,
			f.filterGraph,
			ci)); ret < 0 {
			return nil, fmt.Errorf("error creating filter 'buffer' - %s", AvError(ret))
		}

		last = C.gmf_get_current(f.bufferCtx, ci)

		if ret = int(C.avfilter_link(last, 0, curr.filter_ctx, C.uint(i))); ret < 0 {
			return nil, fmt.Errorf("error linking filters - %s", AvError(ret))
		}

		i++
	}

	if ret = int(C.avfilter_graph_create_filter(
		&f.sinkCtx,
		C.avfilter_get_by_name(C.CString("buffersink")),
		cnameOut,
		nil,
		nil,
		f.filterGraph)); ret < 0 {
		return nil, fmt.Errorf("error creating filter 'buffersink' - %s", AvError(ret))
	}

	// XXX PIXFMT!
	if ret = int(C.avfilter_graph_create_filter(
		&format,
		C.avfilter_get_by_name(C.CString("format")),
		C.CString("format"),
		C.CString("yuv420p"),
		nil,
		f.filterGraph)); ret < 0 {
		return nil, fmt.Errorf("error creating filter 'buffer' - %s", AvError(ret))
	}

	if ret = int(C.avfilter_link(outputs.filter_ctx, 0, format, 0)); ret < 0 {
		return nil, fmt.Errorf("error linking output filters - %s", AvError(ret))
	}

	if ret = int(C.avfilter_link(format, 0, f.sinkCtx, 0)); ret < 0 {
		return nil, fmt.Errorf("error linking output filters - %s", AvError(ret))
	}

	if ret = int(C.avfilter_graph_config(f.filterGraph, nil)); ret < 0 {
		return nil, fmt.Errorf("graph config error - %s", AvError(ret))
	}

	fmt.Printf("%s\n", C.GoString(C.avfilter_graph_dump(f.filterGraph, nil)))

	C.avfilter_inout_free(&inputs)
	C.avfilter_inout_free(&outputs)

	return f, nil
}

func (f *Filter) AddFrame(frame *Frame, istIdx int) error {
	var ret int

	fmt.Printf("AddFrame: i=%d, width=%d, height=%d\n", istIdx, frame.Width(), frame.Height())

	if ret = int(C.gmf_av_buffersrc_add_frame_flags(
		f.bufferCtx,
		frame.avFrame,
		AV_BUFFERSRC_FLAG_PUSH,
		C.int(istIdx)),
	); ret < 0 {
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
		fmt.Printf("ret=%d\n", ret)
		if AvErrno(ret) == syscall.EAGAIN || ret == AVERROR_EOF {
			frame.Free()
			break
		} else if ret < 0 {
			frame.Free()
			return nil, AvError(ret)
		}

		result = append(result, frame)
	}

	return result, nil
}

func (f *Filter) GetFrame2() ([]*Frame, int) {
	var (
		ret    int
		result []*Frame = make([]*Frame, 0)
	)

	for {
		frame := NewFrame()

		if ret = int(C.av_buffersink_get_frame_flags(f.sinkCtx, frame.avFrame, AV_BUFFERSINK_FLAG_NO_REQUEST)); ret < 0 {
			return nil, ret
		}

		result = append(result, frame)
	}

	return result, ret
}

func (f *Filter) RequestOldest() error {
	var ret int

	if ret = int(C.avfilter_graph_request_oldest(f.filterGraph)); ret < 0 {
		return AvError(ret)
	}

	return nil
}
