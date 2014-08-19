package main

import (
	"errors"
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
	"runtime/debug"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
	os.Exit(0)
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		fatal(err)
	}

	return i
}

func addStream(codecName string, oc *FmtCtx, ist *Stream) (int, int) {
	var cc *CodecCtx
	var ost *Stream

	codec := assert(NewEncoder(codecName)).(*Codec)

	// Create Video stream in output context
	if ost = oc.NewStream(codec); ost == nil {
		fatal(errors.New("unable to create stream in output context"))
	}

	if cc = NewCodecCtx(codec); cc == nil {
		fatal(errors.New("unable to create codec context"))
	}

	if oc.IsGlobalHeader() {
		cc.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	if codec.IsExperimental() {
		cc.SetStrictCompliance(-2)
	}

	if cc.Type() == AVMEDIA_TYPE_AUDIO {
		cc.SetSampleFmt(ist.CodecCtx().SampleFmt())
		cc.SetSampleRate(ist.CodecCtx().SampleRate())
		cc.SetChannels(ist.CodecCtx().Channels())
		cc.SelectChannelLayout()
		cc.SelectSampleRate()

	}

	if cc.Type() == AVMEDIA_TYPE_VIDEO {
		cc.SetTimeBase(AVR{1, 25})
		cc.SetProfile(FF_PROFILE_MPEG4_SIMPLE)
		cc.SetDimension(ist.CodecCtx().Width(), ist.CodecCtx().Height())
		cc.SetPixFmt(ist.CodecCtx().PixFmt())
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	ost.SetCodecCtx(cc)

	return ist.Index(), ost.Index()
}

func main() {
	var srcFileName, dstFileName string
	var stMap map[int]int = make(map[int]int, 0)
	var lastDelta int

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	if len(os.Args) != 3 {
		fmt.Println("Simple transcoder, it guesses source format and codecs and tries to convert it to v:mpeg4/a:mp2.")
		fmt.Println("usage: [input filename] [output.mp4]")
		os.Exit(0)
	} else {
		srcFileName = os.Args[1]
		dstFileName = os.Args[2]
	}

	inputCtx := assert(NewInputCtx(srcFileName)).(*FmtCtx)
	defer inputCtx.CloseInputAndRelease()

	outputCtx := assert(NewOutputCtx(dstFileName)).(*FmtCtx)
	defer outputCtx.CloseOutputAndRelease()

	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Println("No video stream found in", srcFileName)
	} else {
		i, o := addStream("mpeg4", outputCtx, srcVideoStream)
		stMap[i] = o
	}

	srcAudioStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_AUDIO)
	if err != nil {
		log.Println("No audio stream found in", srcFileName)
	} else {
		i, o := addStream("aac", outputCtx, srcAudioStream)
		stMap[i] = o
	}

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	for packet := range inputCtx.GetNewPackets() {
		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)
		ost := assert(outputCtx.GetStream(stMap[ist.Index()])).(*Stream)

		for frame := range packet.Frames(ist.CodecCtx()) {
			if ost.IsAudio() {
				fsTb := AVR{1, ist.CodecCtx().SampleRate()}
				outTb := AVR{1, ist.CodecCtx().SampleRate()}

				frame.SetPts(packet.Pts())

				pts := RescaleDelta(ist.TimeBase(), frame.Pts(), fsTb.AVRational(), frame.NbSamples(), &lastDelta, outTb.AVRational())

				frame.
					SetNbSamples(ost.CodecCtx().FrameSize()).
					SetFormat(ost.CodecCtx().SampleFmt()).
					SetChannelLayout(ost.CodecCtx().ChannelLayout()).
					SetPts(pts)
			} else {
				frame.SetPts(ost.Pts)
			}

			if p, ready, _ := frame.EncodeNewPacket(ost.CodecCtx()); ready {
				if p.Pts() != AV_NOPTS_VALUE {
					p.SetPts(RescaleQ(p.Pts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
				}

				if p.Dts() != AV_NOPTS_VALUE {
					p.SetDts(RescaleQ(p.Dts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
				}

				p.SetStreamIndex(ost.Index())

				if err := outputCtx.WritePacket(p); err != nil {
					fatal(err)
				}
				Release(p)
			}

			ost.Pts++
		}
		Release(packet)
	}

	// Flush encoders
	// @todo refactor it (should be a better way)
	for i := 0; i < outputCtx.StreamsCnt(); i++ {
		ist := assert(inputCtx.GetStream(0)).(*Stream)
		ost := assert(outputCtx.GetStream(stMap[ist.Index()])).(*Stream)

		frame := NewFrame()

		for {
			if p, ready, _ := frame.FlushNewPacket(ost.CodecCtx()); ready {
				if p.Pts() != AV_NOPTS_VALUE {
					p.SetPts(RescaleQ(p.Pts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
				}

				if p.Dts() != AV_NOPTS_VALUE {
					p.SetDts(RescaleQ(p.Dts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
				}

				p.SetStreamIndex(ost.Index())

				if err := outputCtx.WritePacket(p); err != nil {
					fatal(err)
				}
				Release(p)
			} else {
				Release(p)
				break
			}


			ost.Pts++
		}

		Release(frame)
	}
}
