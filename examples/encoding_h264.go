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

package main

import (
	"errors"
	"fmt"
	"log"

	. "github.com/3d0c/gmf"
)

func fatal(err error) {
	log.Fatal(err)
}

func audio(outputCtx *FmtCtx, output chan *Packet) *Stream {
	mic, _ := NewInputCtxWithFormatName("sine=b=4", "lavfi")
	mic.Dump()

	ast, err := mic.GetBestStream(AVMEDIA_TYPE_AUDIO)
	if err != nil {
		log.Fatal("failed to find audio stream")
	}
	cc := ast.CodecCtx()

	/// fifo
	fifo := NewAVAudioFifo(cc.SampleFmt(), cc.Channels(), 256)
	if fifo == nil {
		log.Fatal("failed to create audio fifo")
	}

	/// output
	codec, err := FindEncoder("aac")
	if err != nil {
		log.Fatal("find encoder error:", err.Error())
	}

	occ := NewCodecCtx(codec)
	if occ == nil {
		log.Fatal("new output codec context error:", err.Error())
	}

	outSampleFmt := AV_SAMPLE_FMT_FLTP
	occ.SetSampleFmt(outSampleFmt).
		SetSampleRate(cc.SampleRate()).
		SetBitRate(128e3)
	channelLayout := occ.SelectChannelLayout()
	occ.SetChannelLayout(channelLayout)

	if err := occ.Open(nil); err != nil {
		log.Fatal("can't open output codec context", err.Error())
		return nil
	}

	/// resample
	options := []*Option{
		{"in_channel_layout", cc.ChannelLayout()},
		{"out_channel_layout", occ.ChannelLayout()},
		{"in_sample_rate", cc.SampleRate()},
		{"out_sample_rate", occ.SampleRate()},
		{"in_sample_fmt", SampleFmt(cc.SampleFmt())},
		{"out_sample_fmt", SampleFmt(outSampleFmt)},
	}

	swrCtx := NewSwrCtx(options, occ)
	if swrCtx == nil {
		log.Fatal("unable to create Swr Context")
	}

	if outputCtx.IsGlobalHeader() {
		occ.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	ost := outputCtx.NewStream(codec)
	if ost == nil {
		log.Fatalf("Unable to create stream for [%s]\n", codec.LongName())
	}

	ost.SetCodecCtx(occ)

	frameSize := occ.FrameSize()
	log.Println("Codec frame size: ", frameSize)

	go func() {
		count := int64(0)
		for packet := range mic.GetNewPackets() {
			srcFrame, got, err := packet.Decode(ast.CodecCtx())
			Release(packet)
			if !got || err != nil {
				log.Println("input audio error:", err)
				continue
			}

			fifo.Write(srcFrame)

			for fifo.SamplesToRead() >= frameSize {
				winFrame := fifo.Read(frameSize)
				dstFrame := swrCtx.Convert(winFrame)
				Release(winFrame)

				if dstFrame == nil {
					continue
				}
				count += int64(frameSize)

				dstFrame.SetPts(count)

				writePacket, err := dstFrame.Encode(ost.CodecCtx())
				if err != nil {
					log.Fatal(err)
				}
				if writePacket != nil {
					writePacket.SetStreamIndex(ost.Index())
					output <- writePacket
				}
				Release(dstFrame)
			}
			Release(srcFrame)
		}
	}()
	return ost
}
func video(outputCtx *FmtCtx, output chan *Packet) *Stream {
	in, err := NewInputCtxWithFormatName("testsrc=decimals=3", "lavfi")
	ist, err := in.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Fatal("Can't open test video source")
	}

	ist.CodecCtx().PixFmt()

	codec, err := FindEncoder(AV_CODEC_ID_H264)
	if err != nil {
		fatal(err)
	}
	videoEncCtx := NewCodecCtx(codec)
	if videoEncCtx == nil {
		fatal(err)
	}

	videoEncCtx.
		SetBitRate(1e6).
		SetWidth(ist.CodecCtx().Width()).
		SetHeight(ist.CodecCtx().Height()).
		SetTimeBase(ist.TimeBase().AVR()).
		SetPixFmt(AV_PIX_FMT_YUV420P).
		SetMbDecision(FF_MB_DECISION_RD)

	if outputCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		fatal(errors.New(fmt.Sprintf("Unable to create stream for videoEnc [%s]\n", codec.LongName())))
	}
	if err := videoEncCtx.Open(nil); err != nil {
		fatal(err)
	}

	videoStream.SetCodecCtx(videoEncCtx)

	swsCtx := NewSwsCtx(ist.CodecCtx(), videoEncCtx, SWS_BICUBIC)

	dstFrame := NewFrame().
		SetWidth(ist.CodecCtx().Width()).
		SetHeight(ist.CodecCtx().Height()).
		SetFormat(AV_PIX_FMT_YUV420P)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	go func() {
		i := int64(0)

		for packet := range in.GetNewPackets() {
			if packet.StreamIndex() != ist.Index() {
				// skip non video streams
				continue
			}

			for frame := range packet.Frames(ist.CodecCtx()) {
				swsCtx.Scale(frame, dstFrame)
				dstFrame.SetPts(i)

				if p, err := dstFrame.Encode(videoStream.CodecCtx()); p != nil {
					p.SetStreamIndex(videoStream.Index())
					if p.Pts() != AV_NOPTS_VALUE {
						p.SetPts(RescaleQ(p.Pts(), videoEncCtx.TimeBase(), videoStream.TimeBase()))
					}

					if p.Dts() != AV_NOPTS_VALUE {
						p.SetDts(RescaleQ(p.Dts(), videoEncCtx.TimeBase(), videoStream.TimeBase()))
					}
					output <- p
				} else if err != nil {
					fatal(err)
				} else {
					log.Printf("encode frame=%d frame=%d is not ready", i, frame.Pts())
				}

				i++
			}
		}
		close(output)
	}()
	return videoStream
}
func main() {
	outputfilename := "sample-encoding.mp4"

	outputCtx, err := NewOutputCtx(outputfilename)
	if err != nil {
		fatal(err)
	}

	achan := make(chan *Packet)
	ast := audio(outputCtx, achan)

	vchan := make(chan *Packet)
	vst := video(outputCtx, vchan)

	outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	var vp *Packet
	i := 0
	for ap := range achan {
		for {
			if vp == nil {
				vp = <-vchan
			}

			if CompareTimeStamp(int(vp.Pts()), vst.TimeBase(), int(ap.Pts()), ast.TimeBase()) <= 0 {
				if err := outputCtx.WritePacket(vp); err != nil {
					fatal(err)
				}
				Release(vp)
				vp = nil
				i++
				continue
			} else {
				if err := outputCtx.WritePacket(ap); err != nil {
					fatal(err)
				}
				Release(ap)
			}
			break
		}
		if i > 200 {
			break
		}
	}

	outputCtx.CloseOutputAndRelease()
}
