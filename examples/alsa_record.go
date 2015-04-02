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
	"log"

	. "github.com/3d0c/gmf"
)

func main() {
	/// input
	mic, _ := NewInputCtxWithFormatName("default", "alsa")
	mic.Dump()

	ast, err := mic.GetBestStream(AVMEDIA_TYPE_AUDIO)
	if err != nil {
		log.Fatal("failed to find audio stream")
	}
	cc := ast.CodecCtx()

	/// fifo
	fifo := NewAVAudioFifo(cc.SampleFmt(), cc.Channels(), 1024)
	if fifo == nil {
		log.Fatal("failed to create audio fifo")
	}

	/// output
	codecName := ""
	switch cc.SampleFmt() {
	case AV_SAMPLE_FMT_S16:
		codecName = "pcm_s16le"
	default:
		log.Fatal("sample format not support")
	}
	codec, err := FindEncoder(codecName)
	if err != nil {
		log.Fatal("find encoder error:", err.Error())
	}

	occ := NewCodecCtx(codec)
	if occ == nil {
		log.Fatal("new output codec context error:", err.Error())
	}
	defer Release(occ)

	occ.SetSampleFmt(cc.SampleFmt()).
		SetSampleRate(cc.SampleRate()).
		SetChannels(cc.Channels())

	if err := occ.Open(nil); err != nil {
		log.Fatal("can't open output codec context", err.Error())
		return
	}
	outputCtx, err := NewOutputCtx("test.wav")
	if err != nil {
		log.Fatal("new output fail", err.Error())
		return
	}

	ost := outputCtx.NewStream(codec)
	if ost == nil {
		log.Fatal("Unable to create stream for [%s]\n", codec.LongName())
	}
	defer Release(ost)

	ost.SetCodecCtx(occ)

	if err := outputCtx.WriteHeader(); err != nil {
		log.Fatal(err.Error())
	}

	count := 0
	for packet := range mic.GetNewPackets() {
		srcFrame, got, ret, err := packet.DecodeToNewFrame(ast.CodecCtx())
		Release(packet)
		if !got || ret < 0 || err != nil {
			log.Println("capture audio error:", err)
			continue
		}

		wrote := fifo.Write(srcFrame)
		count += wrote

		for fifo.SamplesToRead() >= 1152 {
			dstFrame := fifo.Read(1152)
			if dstFrame == nil {
				continue
			}

			writePacket, ready, _ := dstFrame.EncodeNewPacket(occ)
			if ready {
				if err := outputCtx.WritePacket(writePacket); err != nil {
					log.Println("write packet err", err.Error())
				}

				Release(writePacket)
			}
		}
		if count > int(cc.SampleRate())*10 {
			break
		}
		Release(srcFrame)
	}
}
