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

	codec := assert(FindEncoder(codecName)).(*Codec)

	// Create Video stream in output context
	if ost = oc.NewStream(codec); ost == nil {
		fatal(errors.New("unable to create stream in output context"))
	}
	defer Release(ost)

	if cc = NewCodecCtx(codec); cc == nil {
		fatal(errors.New("unable to create codec context"))
	}
	defer Release(cc)

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



func addStreamCode(oc *FmtCtx, ist *Stream) (int, int) {
	var ost *Stream

	// Create Video stream in output context
	if ost = oc.NewStream(ist.CodecCtx().Codec()); ost == nil {
		fatal(errors.New("unable to create stream in output context"))
	}
	defer Release(ost)

	ost.DumpContexCodec(ist)

	if oc.CheckFlagsAVFMTGLOBALHEADER() > 0  {
		ost.SetCodecFlags()
	}

	return ist.Index(), ost.Index()
}

func main() {
	var srcFileName, dstFileName string

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

	inputCtx.Dump()

	outputCtx := assert(NewOutputCtx2(dstFileName,"mpegts")).(*FmtCtx)

	outputCtx.DumpOutput()
	fmt.Println("===================================")
	defer outputCtx.CloseOutputAndRelease()

	for i:=0 ; i < inputCtx.StreamsCnt() ; i++ {
		srcStream,err := inputCtx.GetStream(i)
		if err != nil {
			fmt.Println("GetStream error")
		}

		addStreamCode(outputCtx,srcStream)
	}
//	outputCtx.Dump()
	outputCtx.DumpOutput()

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	first := false
	for packet := range inputCtx.GetNewPackets() {

		if first {
			if err := outputCtx.WritePacket(packet); err != nil {
				fatal(err)
			}
		}

		first = true
		Release(packet)
	}

	outputCtx.WriteTrailer()

}
