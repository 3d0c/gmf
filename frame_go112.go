//go:build go1.12
// +build go1.12

package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include <stdlib.h>
#include "libavcodec/avcodec.h"
#include "libavutil/frame.h"
#include "libavutil/imgutils.h"
#include "libavutil/timestamp.h"
#include "libavutil/timecode.h"
#include "libavutil/common.h"

void gmf_set_frame_data(AVFrame *frame, int idx, int l_size, uint8_t data) {
    if(!frame) {
        fprintf(stderr, "frame is NULL\n");
    }

    frame->data[idx][l_size] = data;
}

int gmf_get_frame_line_size(AVFrame *frame, int idx) {
	return frame->linesize[idx];
}

void gmf_free_data(AVFrame *frame) {
	av_freep(&frame->data[0]);
}

int gmf_get_timecode(AVFrameSideData *sd, AVRational avgFrameRate, char *out) {
	uint32_t *tc = (uint32_t*)sd->data;
	char tcbuf[AV_TIMECODE_STR_SIZE];
	av_timecode_make_smpte_tc_string2(tcbuf, avgFrameRate, tc[1], 0, 0);
	int n;
	n = sprintf(out, "%s", tcbuf);
	return n;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	AV_PICTURE_TYPE_NONE = iota // Undefined
	AV_PICTURE_TYPE_I           // Intra
	AV_PICTURE_TYPE_P           // Predicted
	AV_PICTURE_TYPE_B           // Bi-dir predicted
	AV_PICTURE_TYPE_S           // S(GMC)-VOP MPEG-4
	AV_PICTURE_TYPE_SI          // Switching Intra
	AV_PICTURE_TYPE_SP          // Switching Predicted
	AV_PICTURE_TYPE_BI          // BI type
)

type Frame struct {
	avFrame   *C.struct_AVFrame
	samples   *C.uint8_t
	mediaType int32
	err       error
	freeData  bool
}

func NewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_alloc()}
}

func (f *Frame) Encode(enc *CodecCtx) (*Packet, error) {
	pkt := NewPacket()
	if pkt == nil {
		return nil, fmt.Errorf("unable to initialize new packet")
	}

	if ret := int(C.avcodec_send_frame(enc.avCodecCtx, f.avFrame)); ret < 0 {
		return nil, fmt.Errorf("error sending frame - %v", AvErrno(ret))
	}

	for {
		ret := int(C.avcodec_receive_packet(enc.avCodecCtx, &pkt.avPacket))
		if AvErrno(ret) == syscall.EAGAIN {
			return nil, nil
		}
		if ret < 0 {
			return nil, fmt.Errorf("%v", AvErrno(ret))
		}
		if ret >= 0 {
			break
		}
	}

	return pkt, nil
}

func (f *Frame) Pts() int64 {
	return int64(f.avFrame.pts)
}

func (f *Frame) Unref() {
	C.av_frame_unref(f.avFrame)
}

func (f *Frame) SetPts(val int64) {
	f.avFrame.pts = (C.int64_t)(val)
}

// AVPixelFormat for video frames, AVSampleFormat for audio
func (f *Frame) Format() int {
	return int(f.avFrame.format)
}

func (f *Frame) Width() int {
	return int(f.avFrame.width)
}

func (f *Frame) Height() int {
	return int(f.avFrame.height)
}

func (f *Frame) PktPts() int64 {
	return int64(f.avFrame.pts)
}

func (f *Frame) PktPos() int64 {
	return int64(f.avFrame.pkt_pos)
}

func (f *Frame) SetPktPts(val int64) {
	f.avFrame.pts = (C.int64_t)(val)
}

func (f *Frame) PktDts() int {
	return int(f.avFrame.pkt_dts)
}

func (f *Frame) SetPktDts(val int) {
	f.avFrame.pkt_dts = (C.int64_t)(val)
}

func (f *Frame) KeyFrame() int {
	return int(f.avFrame.key_frame)
}

func (f *Frame) NbSamples() int {
	return int(f.avFrame.nb_samples)
}

func (f *Frame) Channels() int {
	return int(f.avFrame.channels)
}

func (f *Frame) SetFormat(val int32) *Frame {
	f.avFrame.format = C.int(val)
	return f
}

func (f *Frame) SetWidth(val int) *Frame {
	f.avFrame.width = C.int(val)
	return f
}

func (f *Frame) SetHeight(val int) *Frame {
	f.avFrame.height = C.int(val)
	return f
}

func (f *Frame) ImgAlloc() error {
	if ret := int(C.av_image_alloc(
		(**C.uint8_t)(unsafe.Pointer(&f.avFrame.data)),
		(*C.int)(unsafe.Pointer(&f.avFrame.linesize)),
		C.int(f.Width()), C.int(f.Height()), int32(f.Format()), 32)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate raw image buffer: %v", AvError(ret)))
	}

	f.freeData = true

	return nil
}

func NewAudioFrame(sampleFormat int32, channels, nb_samples int) (*Frame, error) {
	f := NewFrame()
	f.mediaType = AVMEDIA_TYPE_AUDIO
	f.SetNbSamples(nb_samples)
	f.SetFormat(sampleFormat)
	f.SetChannels(channels)

	//the codec gives us the frame size, in samples,
	//we calculate the size of the samples buffer in bytes
	size := C.av_samples_get_buffer_size(nil, C.int(channels), C.int(nb_samples),
		sampleFormat, 0)
	if size < 0 {
		return nil, errors.New("could not get sample buffer size")
	}

	f.samples = (*C.uint8_t)(C.av_malloc(C.size_t(size)))
	if f.samples == nil {
		return nil, errors.New(fmt.Sprintf("could not allocate %d bytes for samples buffer", size))
	}

	//setup the data pointers in the AVFrame
	ret := int(C.avcodec_fill_audio_frame(f.avFrame, C.int(channels), sampleFormat,
		f.samples, C.int(size), 0))
	if ret < 0 {
		return nil, errors.New("could not setup audio frame")
	}

	f.freeData = true

	return f, nil
}

func (f *Frame) SetData(idx int, lineSize int, data int) *Frame {
	C.gmf_set_frame_data(f.avFrame, C.int(idx), C.int(lineSize), (C.uint8_t)(data))

	return f
}

func (f *Frame) LineSize(idx int) int {
	return int(C.gmf_get_frame_line_size(f.avFrame, C.int(idx)))
}

func (f *Frame) Dump() {
	fmt.Printf("%v\n", f.avFrame)
}

func (f *Frame) CloneNewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_clone(f.avFrame)}
}

func (f *Frame) Free() {
	if f.freeData && f.avFrame != nil {
		C.gmf_free_data(f.avFrame)
	}
	if f.avFrame != nil {
		C.av_frame_free(&f.avFrame)
	}
}

func (f *Frame) SetNbSamples(val int) *Frame {
	f.avFrame.nb_samples = C.int(val)
	return f
}

func (f *Frame) SetChannelLayout(val int) *Frame {
	f.avFrame.channel_layout = (C.uint64_t)(val)
	return f
}

func (f *Frame) GetChannelLayout() int {
	return int(f.avFrame.channel_layout)
}

func (f *Frame) SetChannels(val int) *Frame {
	f.avFrame.channels = C.int(val)
	return f
}

func (f *Frame) SetQuality(val int) *Frame {
	f.avFrame.quality = C.int(val)
	return f
}

func (f *Frame) SetPictType(val uint32) {
	f.avFrame.pict_type = val
}

func (f *Frame) IsNil() bool {
	if f.avFrame == nil {
		return true
	}

	return false
}

func (f *Frame) GetRawFrame() *C.struct_AVFrame {
	return f.avFrame
}

func (f *Frame) GetRawAudioData(plane int) []byte {
	return C.GoBytes(unsafe.Pointer(f.avFrame.data[plane]), C.int(f.LineSize(plane)))
}

func (f *Frame) Time(timebase AVRational) int {
	return int(float64(timebase.AVR().Num) / float64(timebase.AVR().Den) * float64(f.Pts()))
}

var frameSideDataTypes []uint32 = []uint32{
	C.AV_FRAME_DATA_PANSCAN,
	C.AV_FRAME_DATA_A53_CC,
	C.AV_FRAME_DATA_AFD,
	C.AV_FRAME_DATA_AUDIO_SERVICE_TYPE,
	C.AV_FRAME_DATA_CONTENT_LIGHT_LEVEL,
	C.AV_FRAME_DATA_DISPLAYMATRIX,
	C.AV_FRAME_DATA_DOWNMIX_INFO,
	C.AV_FRAME_DATA_DYNAMIC_HDR_PLUS,
	C.AV_FRAME_DATA_FILM_GRAIN_PARAMS,
	C.AV_FRAME_DATA_GOP_TIMECODE,
	C.AV_FRAME_DATA_ICC_PROFILE,
	C.AV_FRAME_DATA_MASTERING_DISPLAY_METADATA,
	C.AV_FRAME_DATA_MATRIXENCODING,
	C.AV_FRAME_DATA_MOTION_VECTORS,
	C.AV_FRAME_DATA_REGIONS_OF_INTEREST,
	C.AV_FRAME_DATA_REPLAYGAIN,
	C.AV_FRAME_DATA_S12M_TIMECODE,
	C.AV_FRAME_DATA_SEI_UNREGISTERED,
	C.AV_FRAME_DATA_SKIP_SAMPLES,
	C.AV_FRAME_DATA_SPHERICAL,
	C.AV_FRAME_DATA_STEREO3D,
	C.AV_FRAME_DATA_VIDEO_ENC_PARAMS,
}

func (f *Frame) GetSideDataTypes() (map[uint32]string, error) {
	result := make(map[uint32]string)
	// get a pointer to the side data of the frame
	for _, sideDataType := range frameSideDataTypes {
		sideDataPtr := C.av_frame_get_side_data(f.avFrame, sideDataType)
		if sideDataPtr != nil {
			// cast the pointer to the side data to a AVFrameSideData struct
			sideData := (*C.struct_AVFrameSideData)(unsafe.Pointer(sideDataPtr))

			sideDataType := C.GoString(C.av_frame_side_data_name(sideData._type))
			result[uint32(sideData._type)] = sideDataType
		}
	}

	return result, nil
}

func (f *Frame) GetUserData() ([]byte, error) {
	// get a pointer to the side data of the frame
	sideDataPtr := C.av_frame_get_side_data(f.avFrame, C.AV_FRAME_DATA_SEI_UNREGISTERED)

	if sideDataPtr == nil {
		return nil, errors.New("no user data found")
	}

	// cast the pointer to the side data to a AVFrameSideData struct
	sideData := (*C.struct_AVFrameSideData)(unsafe.Pointer(sideDataPtr))

	// gets the bytes from the pointer
	data := C.GoBytes(unsafe.Pointer(sideData.data), C.int(sideData.size))

	return data, nil
}

func (f *Frame) GetTimeCode(avgFrameRate AVRational) (string, error) {
	// get a pointer to the side data of the frame
	sideDataPtr := C.av_frame_get_side_data(f.avFrame, C.AV_FRAME_DATA_S12M_TIMECODE)

	if sideDataPtr == nil {
		return "", errors.New("no timecode data found")
	}

	// cast the pointer to the side data to a AVFrameSideData struct
	sideData := (*C.struct_AVFrameSideData)(unsafe.Pointer(sideDataPtr))

	// check if the side data is valid
	if sideData._type != C.AV_FRAME_DATA_S12M_TIMECODE || sideData.size != 16 {
		return "", errors.New("invalid timecode side data")
	}

	// prepare a byte slice to hold the timecode
	ptr := C.malloc(C.sizeof_char * 1024)
	defer C.free(unsafe.Pointer(ptr))

	// cast the go avgFrameRate to a C AVRational
	afr := (C.struct_AVRational)(avgFrameRate)

	// get the timecode by calling the C function and passing the sideData, the AVRational and the pointer to the byte slice to hold the timecode
	// returns the length of the timecode in bytes
	tcSize := C.gmf_get_timecode(sideData, afr, (*C.char)(ptr))

	// gets the bytes from the pointer
	tc := C.GoBytes(ptr, tcSize)

	tcString := string(tc)
	return tcString, nil
}
