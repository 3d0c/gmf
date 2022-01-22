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
	"fmt"
	"log"

	"github.com/3d0c/gmf"
)

func main() {
	/// input
	mic, _ := gmf.NewInputCtxWithFormatName("default", "alsa")
	mic.Dump()

	ast, err := mic.GetBestStream(gmf.AVMEDIA_TYPE_AUDIO)
	if err != nil {
		log.Fatal("failed to find audio stream")
	}
	cc := ast.CodecCtx()

	/// fifo
	fifo := gmf.NewAVAudioFifo(cc.SampleFmt(), cc.Channels(), 1024)
	if fifo == nil {
		log.Fatal("failed to create audio fifo")
	}

	/// output
	codecName := ""
	switch cc.SampleFmt() {
	case gmf.AV_SAMPLE_FMT_S16:
		codecName = "pcm_s16le"
	default:
		log.Fatal("sample format not support")
	}
	codec, err := gmf.FindEncoder(codecName)
	if err != nil {
		log.Fatal("find encoder error:", err.Error())
	}

	audioEncCtx := gmf.NewCodecCtx(codec)
	if audioEncCtx == nil {
		log.Fatal("new output codec context error:", err.Error())
	}
	defer audioEncCtx.Free()

	outputCtx, err := gmf.NewOutputCtx("test.wav")
	if err != nil {
		log.Fatal("new output fail", err.Error())
		return
	}
	defer outputCtx.Free()

	audioEncCtx.SetSampleFmt(cc.SampleFmt()).
		SetSampleRate(cc.SampleRate()).
		SetChannels(cc.Channels())

	if outputCtx.IsGlobalHeader() {
		audioEncCtx.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	audioStream := outputCtx.NewStream(codec)
	if audioStream == nil {
		log.Fatal(fmt.Errorf("unable to create stream for audioEnc [%s]", codec.LongName()))
	}
	defer audioStream.Free()

	if err := audioEncCtx.Open(nil); err != nil {
		log.Fatal("can't open output codec context", err.Error())
		return
	}
	audioStream.DumpContexCodec(audioEncCtx)

	outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		log.Fatal(err.Error())
	}

	outputCtx.Dump()

	count := 0
	for packet := range mic.GetNewPackets() {
		srcFrames, err := ast.CodecCtx().Decode(packet)
		packet.Free()
		if err != nil {
			log.Println("capture audio error:", err)
			continue
		}

		for _, srcFrame := range srcFrames {
			wrote := fifo.Write(srcFrame)
			count += wrote

			for fifo.SamplesToRead() >= 1152 {
				dstFrame := fifo.Read(1152)
				if dstFrame == nil {
					continue
				}

				writePacket, err := dstFrame.Encode(audioEncCtx)
				if err == nil {
					if err := outputCtx.WritePacket(writePacket); err != nil {
						log.Println("write packet err", err.Error())
					}

					writePacket.Free()
				} else {
					log.Fatal(err)
				}
				dstFrame.Free()
			}
			if count > int(cc.SampleRate())*10 {
				break
			}
		}
	}
}
