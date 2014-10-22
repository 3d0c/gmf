package main

import (
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

func main() {
	var srcFileName, dstFileName string

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	if len(os.Args) != 3 {
		fmt.Println("usage:",os.Args[0] ," input output")
		fmt.Println("API example program to remux a media file with libavformat and libavcodec.")
		fmt.Println("The output format is guessed according to the file extension.")

		os.Exit(0)
	} else {
		srcFileName = os.Args[1]
		dstFileName = os.Args[2]
	}

	inputCtx := assert(NewInputCtx(srcFileName)).(*FmtCtx)
	defer inputCtx.CloseInputAndRelease()
	inputCtx.Dump()

	outputCtx := assert(NewOutputCtxWithFormatName(dstFileName,"mpegts")).(*FmtCtx)
	defer outputCtx.CloseOutputAndRelease()

	fmt.Println("===================================")

	for i:=0 ; i < inputCtx.StreamsCnt() ; i++ {
		srcStream,err := inputCtx.GetStream(i)
		if err != nil {
			fmt.Println("GetStream error")
		}

		outputCtx.AddStreamWithCodeCtx(srcStream.CodecCtx())
	}
	outputCtx.Dump()

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	first := false
	for packet := range inputCtx.GetNewPackets() {

		if first {  //if read from rtsp ,the first packets is wrong.
			if err := outputCtx.WritePacket(packet); err != nil {
				fatal(err)
			}
		}

		first = true
		Release(packet)
	}

}
