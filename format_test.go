package gmf

import (
	"errors"
	"fmt"
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
	ctx := NewCtx()

	if ctx.avCtx == nil {
		t.Fatal("AVContext is not initialized")
	}

	Release(ctx)
}

func TestCtxInput(t *testing.T) {
	inputCtx, err := NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	inputCtx.CloseInputAndRelease()
}

func TestCtxOutput(t *testing.T) {
	cases := map[interface{}]error{
		outputSampleFilename:                        nil,
		FindOutputFmt("mp4", "", ""):                nil,
		FindOutputFmt("", outputSampleFilename, ""): nil,
		FindOutputFmt("", "", "application/mp4"):    nil,
		FindOutputFmt("", "", "wrong/mime"):         errors.New(fmt.Sprintf("output format is not initialized. Unable to allocate context")),
	}

	for arg, expected := range cases {
		if outuptCtx, err := NewOutputCtx(arg); err != nil {
			if err.Error() != expected.Error() {
				t.Error("Unexpected error:", err)
			}
		} else {
			outuptCtx.CloseOutputAndRelease()
		}
	}

	log.Println("OutputContext is OK.")
}

func TestCtxCloseEmpty(t *testing.T) {
	ctx := NewCtx()

	ctx.CloseInputAndRelease()
	ctx.CloseOutputAndRelease()
	Release(ctx)
}

func TestNewStream(t *testing.T) {
	ctx := NewCtx()
	if ctx.avCtx == nil {
		t.Fatal("AVContext is not initialized")
	}
	defer Release(ctx)

	c := assert(FindEncoder(AV_CODEC_ID_MPEG1VIDEO)).(*Codec)

	cc := NewCodecCtx(c)
	defer Release(cc)

	cc.SetTimeBase(AVR{1, 25})
	cc.SetDimension(320, 200)

	if ctx.IsGlobalHeader() {
		cc.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	log.Println("Dummy stream is created")
}

func TestWriteHeader(t *testing.T) {
	outputCtx, err := NewOutputCtx(outputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer Release(outputCtx)

	// write_header needs a valid stream with code context initialized
	c := assert(FindEncoder(AV_CODEC_ID_MPEG1VIDEO)).(*Codec)
	stream := outputCtx.NewStream(c)
	defer Release(stream)
	cc := NewCodecCtx(c).SetTimeBase(AVR{1, 25}).SetDimension(10, 10).SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	defer Release(cc)
	stream.SetCodecCtx(cc)

	if err := outputCtx.WriteHeader(); err != nil {
		t.Fatal(err)
	}

	log.Println("Header has been written to", outputSampleFilename)

	if err := os.Remove(outputSampleFilename); err != nil {
		log.Fatal(err)
	}
}

func TestPacketsIterator(t *testing.T) {
	inputCtx, err := NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer inputCtx.CloseInputAndRelease()

	for packet := range inputCtx.GetNewPackets() {
		if packet.Size() <= 0 {
			t.Fatal("Expected size > 0")
		} else {
			log.Printf("One packet has been read. size: %v, pts: %v\n", packet.Size(), packet.Pts())
		}
		Release(packet)

		break
	}
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

	b := make([]byte, IO_BUFFER_SIZE)

	n, err := section.Read(b)
	if err != nil {
		fmt.Println("section.Read():", err)
		file.Close()
	}

	return b, n
}

func TestAVIOContext(t *testing.T) {
	ictx := NewCtx()

	if err := ictx.SetInputFormat("mov"); err != nil {
		t.Fatal(err)
	}

	avioCtx, err := NewAVIOContext(ictx, &AVIOHandlers{ReadPacket: customReader})
	defer Release(avioCtx)
	if err != nil {
		t.Fatal(err)
	}

	ictx.SetPb(avioCtx).SetFlag(AV_NOPTS_VALUE).OpenInput("")

	for p := range ictx.GetNewPackets() {
		_ = p
		Release(p)
	}

	ictx.CloseInputAndRelease()

}

func ExampleNewAVIOContext(t *testing.T) {
	ctx := NewCtx()
	defer Release(ctx)

	// In this example, we're using custom reader implementation,
	// so we should specify format manually.
	if err := ctx.SetInputFormat("mov"); err != nil {
		t.Fatal(err)
	}

	avioCtx, err := NewAVIOContext(ctx, &AVIOHandlers{ReadPacket: customReader})
	defer Release(avioCtx)
	if err != nil {
		t.Fatal(err)
	}

	// Setting up AVFormatContext.pb
	ctx.SetPb(avioCtx).SetFlag(AV_NOPTS_VALUE)

	// Calling OpenInput with empty arg, because all files stuff we're doing in custom reader.
	// But the library have to initialize some stuff, so we call it anyway.
	ctx.OpenInput("")

	for p := range ctx.GetNewPackets() {
		_ = p
		Release(p)
	}
}
