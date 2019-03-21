package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"
#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>

int gmf_sw_resample(SwrContext* ctx, AVFrame* dstFrame, AVFrame* srcFrame){
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		(const uint8_t **)srcFrame->data, srcFrame->nb_samples);
}

int gmf_swr_flush(SwrContext* ctx, AVFrame* dstFrame) {
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		NULL, 0);
}

*/
import "C"

import (
	"fmt"
)

type SwrCtx struct {
	swrCtx   *C.struct_SwrContext
	channels int
	format   int32
}

func NewSwrCtx(options []*Option, channels int, format int32) (*SwrCtx, error) {
	ctx := &SwrCtx{
		swrCtx:   C.swr_alloc(),
		channels: channels,
		format:   format,
	}

	for _, option := range options {
		option.Set(ctx.swrCtx)
	}

	if ret := int(C.swr_init(ctx.swrCtx)); ret < 0 {
		return nil, fmt.Errorf("error initializing swr context - %s", AvError(ret))
	}

	return ctx, nil
}

func (ctx *SwrCtx) Free() {
	C.swr_free(&ctx.swrCtx)
}

func (ctx *SwrCtx) Convert(input *Frame) (*Frame, error) {
	var (
		dst *Frame
		err error
	)

	if dst, err = NewAudioFrame(ctx.format, ctx.channels, input.NbSamples()); err != nil {
		return nil, fmt.Errorf("error creating new audio frame - %s\n", err)
	}

	C.gmf_sw_resample(ctx.swrCtx, dst.avFrame, input.avFrame)

	return dst, nil
}

func (ctx *SwrCtx) Flush(nbSamples int) (*Frame, error) {
	var (
		dst *Frame
		err error
	)

	if dst, err = NewAudioFrame(ctx.format, ctx.channels, nbSamples); err != nil {
		return nil, fmt.Errorf("error creating new audio frame - %s\n", err)
	}

	C.gmf_swr_flush(ctx.swrCtx, dst.avFrame)

	return dst, nil
}

func DefaultResampler(ost *Stream, frames []*Frame, flush bool) []*Frame {
	var (
		result             []*Frame = make([]*Frame, 0)
		winFrame, tmpFrame *Frame
	)

	if ost.SwrCtx == nil || ost.AvFifo == nil {
		return frames
	}

	frameSize := ost.CodecCtx().FrameSize()

	for i, _ := range frames {
		ost.AvFifo.Write(frames[i])

		for ost.AvFifo.SamplesToRead() >= frameSize {
			winFrame = ost.AvFifo.Read(frameSize)
			winFrame.SetChannelLayout(ost.CodecCtx().GetDefaultChannelLayout(ost.CodecCtx().Channels()))

			tmpFrame, _ = ost.SwrCtx.Convert(winFrame)
			if tmpFrame == nil || tmpFrame.IsNil() {
				break
			}

			tmpFrame.SetPts(ost.Pts)
			tmpFrame.SetPktDts(int(ost.Pts))

			ost.Pts += int64(frameSize)

			result = append(result, tmpFrame)
		}
	}

	if flush {
		if tmpFrame, _ = ost.SwrCtx.Flush(frameSize); tmpFrame != nil && !tmpFrame.IsNil() {
			tmpFrame.SetPts(ost.Pts)
			tmpFrame.SetPktDts(int(ost.Pts))

			ost.Pts += int64(frameSize)

			result = append(result, tmpFrame)
		}
	}

	for i := 0; i < len(frames); i++ {
		if frames[i] != nil {
			frames[i].Free()
		}
	}

	return result
}
