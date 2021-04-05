package gmf_test

import (
	"errors"
	"fmt"
	"github.com/3d0c/gmf"
	"io"
	"log"
	"os"
	"testing"
)

var (
	inputSampleFilename  string = "examples/tests-sample.mp4"
	outputSampleFilename string = "examples/tests-output.mp4"
	inputSampleWidth     int    = 320
	inputSampleHeight    int    = 200
)

func assert(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}

func TestCtxCreation(t *testing.T) {
	ctx := gmf.NewCtx()

	if ctx == nil {
		t.Fatal("AVContext is not initialized")
	}

	ctx.Free()
}

func TestCtxInput(t *testing.T) {
	inputCtx, err := gmf.NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	inputCtx.Free()
}

func TestCtxOutput(t *testing.T) {
	cases := map[interface{}]error{
		outputSampleFilename:                            nil,
		gmf.FindOutputFmt("mp4", "", ""):                nil,
		gmf.FindOutputFmt("", outputSampleFilename, ""): nil,
		gmf.FindOutputFmt("", "", "application/mp4"):    nil,
		gmf.FindOutputFmt("", "", "wrong/mime"):         errors.New(fmt.Sprintf("output format is not initialized. Unable to allocate context")),
	}

	for arg, expected := range cases {
		if outuptCtx, err := gmf.NewOutputCtx(arg); err != nil {
			if err.Error() != expected.Error() {
				t.Error("Unexpected error:", err)
			}
		} else {
			outuptCtx.Free()
		}
	}

	log.Println("OutputContext is OK.")
}

func TestCtxCloseEmpty(t *testing.T) {
	ctx := gmf.NewCtx()

	ctx.Free()
}

func TestNewStream(t *testing.T) {
	ctx := gmf.NewCtx()
	if ctx == nil {
		t.Fatal("AVContext is not initialized")
	}
	ctx.Free()

	c := assert(gmf.FindEncoder(gmf.AV_CODEC_ID_MPEG1VIDEO)).(*gmf.Codec)

	cc := gmf.NewCodecCtx(c)
	defer cc.Free()

	cc.SetTimeBase(gmf.AVR{Num: 1, Den: 25})
	cc.SetDimension(320, 200)

	if ctx.IsGlobalHeader() {
		cc.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	log.Println("Dummy stream is created")
}

func TestWriteHeader(t *testing.T) {
	outputCtx, err := gmf.NewOutputCtx(outputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer outputCtx.Free()

	// write_header needs a valid stream with code context initialized
	c := assert(gmf.FindEncoder(gmf.AV_CODEC_ID_MPEG1VIDEO)).(*gmf.Codec)
	stream := outputCtx.NewStream(c)
	defer stream.Free()
	cc := gmf.NewCodecCtx(c).SetTimeBase(gmf.AVR{Num: 1, Den: 25}).SetDimension(10, 10).SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	defer cc.Free()
	stream.DumpContexCodec(cc)
	// stream.SetCodecCtx(cc)

	if err := outputCtx.WriteHeader(); err != nil {
		t.Fatal(err)
	}

	log.Println("Header has been written to", outputSampleFilename)

	if err := os.Remove(outputSampleFilename); err != nil {
		log.Fatal(err)
	}
}

func TestPacketsIterator(t *testing.T) {
	inputCtx, err := gmf.NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer inputCtx.Free()

	for packet := range inputCtx.GetNewPackets() {
		if packet.Size() <= 0 {
			t.Fatal("Expected size > 0")
		} else {
			log.Printf("One packet has been read. size: %v, pts: %v\n", packet.Size(), packet.Pts())
		}
		packet.Free()

		break
	}
}

func TestGetNextPacket(t *testing.T) {
	inputCtx, err := gmf.NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer inputCtx.Free()

	packet, _ := inputCtx.GetNextPacket()
	if packet.Size() <= 0 {
		t.Fatal("Expected size > 0")
	} else {
		log.Printf("One packet has been read. size: %v, pts: %v\n", packet.Size(), packet.Pts())
	}
	packet.Free()
}

var section *io.SectionReader

func customReader() ([]byte, int) {
	var file *os.File
	var err error

	if section == nil {
		file, err = os.Open(inputSampleFilename)
		if err != nil {
			panic(err)
		}

		fi, err := file.Stat()
		if err != nil {
			panic(err)
		}

		section = io.NewSectionReader(file, 0, fi.Size())
	}

	b := make([]byte, gmf.IO_BUFFER_SIZE)

	n, err := section.Read(b)
	if err != nil {
		fmt.Println("section.Read():", err)
		file.Close()
	}

	return b, n
}

func TestAVIOContext(t *testing.T) {
	ictx := gmf.NewCtx()

	if err := ictx.SetInputFormat("mov"); err != nil {
		t.Fatal(err)
	}

	avioCtx, err := gmf.NewAVIOContext(ictx, &gmf.AVIOHandlers{ReadPacket: customReader})
	defer avioCtx.Free()
	if err != nil {
		t.Fatal(err)
	}

	ictx.SetPb(avioCtx).OpenInput("")

	for p := range ictx.GetNewPackets() {
		_ = p
		p.Free()
	}

	ictx.Free()

}

func ExampleNewAVIOContext() {
	ctx := gmf.NewCtx()
	defer ctx.Free()

	// In this example, we're using custom reader implementation,
	// so we should specify format manually.
	if err := ctx.SetInputFormat("mov"); err != nil {
		log.Fatal(err)
	}

	avioCtx, err := gmf.NewAVIOContext(ctx, &gmf.AVIOHandlers{ReadPacket: customReader})
	defer avioCtx.Free()
	if err != nil {
		log.Fatal(err)
	}

	// Setting up AVFormatContext.pb
	ctx.SetPb(avioCtx)

	// Calling OpenInput with empty arg, because all files stuff we're doing in custom reader.
	// But the library have to initialize some stuff, so we call it anyway.
	ctx.OpenInput("")

	for p := range ctx.GetNewPackets() {
		_ = p
		p.Free()
	}
}
