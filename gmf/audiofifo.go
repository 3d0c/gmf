/*
Copyright (c) 2015, EMSYM Corporation

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

    * Redistributions of source code must retain the above copyright notice,
      this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright notice,
      this list of conditions and the following disclaimer in the documentation
      and/or other materials provided with the distribution.
    * Neither the name of EMSYM Corporation nor the names of its contributors
      may be used to endorse or promote products derived from this software
      without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF
THE POSSIBILITY OF SUCH DAMAGE.

Sleepy Programmer <hunan@emsym.com>

*/

package gmf

/*
#cgo pkg-config: libavutil
#include <libavutil/audio_fifo.h>
#include <libavutil/frame.h>

int write_fifo(AVAudioFifo* fifo, AVFrame* frame, int nb_samples){
	return av_audio_fifo_write(fifo,(void**)frame->data,nb_samples);
}

int read_fifo(AVAudioFifo* fifo, AVFrame* frame, int nb_samples){
	return av_audio_fifo_read(fifo,(void**)frame->data,nb_samples);
}
*/
import "C"

type AVAudioFifo struct {
	avAudioFifo  *C.struct_AVAudioFifo
	sampleFormat int32
	channels     int
}

func NewAVAudioFifo(sampleFormat int32, channels int, nb_samples int) *AVAudioFifo {
	fifo := C.av_audio_fifo_alloc(sampleFormat, C.int(channels), C.int(nb_samples))
	if fifo == nil {
		panic("unable to allocate fifo context\n")
		return nil
	}

	return &AVAudioFifo{
		avAudioFifo:  fifo,
		sampleFormat: sampleFormat,
		channels:     channels,
	}
}

func (this *AVAudioFifo) SamplesToRead() int {
	return int(C.av_audio_fifo_size(this.avAudioFifo))
}

func (this *AVAudioFifo) SamplesCanWrite() int {
	return int(C.av_audio_fifo_space(this.avAudioFifo))
}

func (this *AVAudioFifo) Write(frame *Frame) int {
	return int(C.write_fifo(this.avAudioFifo, frame.avFrame, C.int(frame.NbSamples())))
}

func (this *AVAudioFifo) Read(sampleCount int) *Frame {
	rsize := this.SamplesToRead()
	size := rsize

	if sampleCount <= rsize {
		size = sampleCount
	}

	frame, err := NewAudioFrame(this.sampleFormat, this.channels, size)
	if frame == nil || err != nil {
		return nil
	}

	if int(C.read_fifo(this.avAudioFifo, frame.avFrame, C.int(size))) == size {
		return frame
	}

	frame.Free()
	return nil
}

func (this *AVAudioFifo) Free() {
	C.av_audio_fifo_free(this.avAudioFifo)
}
