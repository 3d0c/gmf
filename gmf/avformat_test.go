package gmf

import "testing"
import "unsafe"

//import "time"
var filename = "../video.mp4"

//var filename = "/media/video/big_buck_bunny_480p_surround-fix.avi"

func TestCopyStream(t *testing.T) {
	/*openning input file*/
	ctx := avformat_alloc_context()
	result := av_open_input_file(ctx, filename, nil, 0, nil)
	if result != 0 {
		t.Fatalf("error while openning input file")
	}

	result = av_find_stream_info(ctx)
	if result < 0 {
		t.Fatalf("error while find stream info")
	}

	/*openning the outputfile*/
	outformat := "flv"
	outfile := "test." + outformat
	outctx := avformat_alloc_context()
	format := av_guess_format(outformat, outfile)
	outctx.ctx.oformat = (*_Ctype_struct_AVOutputFormat)(unsafe.Pointer(format.format))

	result = url_fopen(outctx, outfile)
	if result < 0 {
		t.Fatalf("error while open outputfile")
	}
	/*add new output stream*/
	stream := av_new_stream(outctx, 0)
	//decoder:=ctx.ctx.streams[0]
	coder := Decoder{coder{Ctx: _CodecContext{ctx: ctx.ctx.streams[0].codec}}, 0, Rational{}, Rational{}}
	coder.Open()
	//coder.Ctx.ctx=self.codec//dpx.Ds.Ctx.streams[streamid].codec

	var encoder Encoder
	encoder.SetParameter("codecid", "22")
	encoder.SetParameter("time_base", "1/15")
	encoder.SetParameter("width", "320")
	encoder.SetParameter("height", "240")
	encoder.SetParameter("bf", "0")
	encoder.SetParameter("b", "200000")
	encoder.SetParameter("g", "250")
	encoder.SetParameter("qmin", "2")
	encoder.SetParameter("qmax", "51")
	encoder.SetParameter("qdiff", "4")
	encoder.SetParameter("flags", "+global_header")
	encoder.Open()

	stream.codec = encoder.Ctx.ctx
	//stream.codec=decoder.codec
	//println(decoder.codec)

	result = av_write_header(outctx)
	if result < 0 {
		t.Fatalf("error while write output header")
	}

	var packet *avPacket = new(avPacket)
	//av_init_packet(packet)
	var sumout = 0
	var sumin = 0
	resizer := new(Resizer)
	resizer.Init(&coder, &encoder)

	for av_read_frame(ctx, packet) >= 0 {
		//time.Sleep(500000)
		//packet.destroy()
		//continue
		/*
				if(packet.stream_index==0){
				    frame:=coder.Decode(packet)
				    if(frame!=nil&&frame.isFinished&&coder.Ctx.ctx.codec_type==CODEC_TYPE_VIDEO){
					of:=resizer.Resize(frame)
					//ppsWriter(of)
					//continue
					//encoder.Encode(*of)

					if true {
					op:=encoder.Encode(of)
					//println(op)
					if(op!=nil){
			    		    sumout+=int(op.avpacket.size)
					    sumin+=int(packet.avpacket.size)
					    av_interleaved_write_frame(outctx,op)
					    //op.destroy()
					}

					}
					//of.destroy()
					//frame.destroy()

				    }
				}*/
		//packet.Free()

	}
	println("size")
	println(sumin)
	println(sumout)
	av_close_input_file(ctx)

	url_fclose(outctx)
	av_free_format_context(outctx)
	//av_free(ctx)
}
