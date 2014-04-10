package main

import (
	// "errors"
	"errors"
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
)

func fatal(err error) {
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

	cc.CopyRequired(ist)

	if cc.Type() == int(AVMEDIA_TYPE_VIDEO) && oc.IsGlobalHeader() {
		cc.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
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

	if len(os.Args) != 3 {
		fmt.Println("Simple transcoder, it guesses source format and codecs and tries to convert it to v:mpeg4/a:mp2.")
		fmt.Println("usage: [input filename] [output.mp4]")
		os.Exit(0)
	} else {
		srcFileName = os.Args[1]
		dstFileName = os.Args[2]
	}

	inputCtx := assert(NewInputCtx(srcFileName)).(*FmtCtx)
	defer inputCtx.CloseInput()

	outputCtx := assert(NewOutputCtx(dstFileName)).(*FmtCtx)
	defer outputCtx.CloseOutput()

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
		i, o := addStream("mp2", outputCtx, srcAudioStream)
		stMap[i] = o
	}

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	i := 0

	for packet := range inputCtx.Packets() {
		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		if ist.CodecCtx().Type() == int(AVMEDIA_TYPE_AUDIO) {
			log.Println("skipping audio packet")
			continue
		}

		frame, got, err := packet.Decode(ist.CodecCtx())
		if got != 0 {
			if ist.CodecCtx().Type() == int(AVMEDIA_TYPE_VIDEO) {
				frame.SetPts(i)
			}

			ost := assert(outputCtx.GetStream(stMap[ist.Index()])).(*Stream)

			if p, ready, _ := frame.Encode(ost.CodecCtx()); ready {
				if ost.CodecCtx().Type() == int(AVMEDIA_TYPE_VIDEO) {
					if p.Pts() != AV_NOPTS_VALUE {
						p.SetPts(RescaleQ(p.Pts(), ost.CodecCtx().TimeBase(), ist.TimeBase()))
					}

					if p.Dts() != AV_NOPTS_VALUE {
						p.SetDts(RescaleQ(p.Dts(), ost.CodecCtx().TimeBase(), ist.TimeBase()))
					}
				}
				if err := outputCtx.WritePacket(p); err != nil {
					fatal(err)
				}
			}
		}

		if got == 0 || err != nil {
			fatal(err)
		}

		frame.Unref()

		i++
	}
}
