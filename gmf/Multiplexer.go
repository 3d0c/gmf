package gmf

//#include <stdint.h>
import "C"
import "unsafe"

import "log"

type Multiplexer struct {
	Ds           DataSink
	tracks       []Track
	ch           chan Packet
	stream_count int
}

func (self *Multiplexer) AddTrack(enc *Encoder) *Track {
	if self.ch == nil {
		self.ch = make(chan Packet)
		self.tracks = make([]Track, 20) //@TODO: ugly hack, max of 20 streams
	}

	result := Track{av_new_stream(self.Ds.ctx, self.stream_count), self.ch, 0}
	enc.stream_index = self.stream_count
	enc.Track = &result
	result.Stream.codec = enc.Ctx.ctx
	result.Stream.time_base = enc.Ctx.ctx.time_base
	result.Stream.sample_aspect_ratio = enc.Ctx.ctx.sample_aspect_ratio

	self.tracks[self.stream_count] = result
	self.stream_count++
	return &result
}

func (self *Multiplexer) Start() {
	av_write_header(self.Ds.ctx)
	//dump_format(self.Ds.ctx)
	/*for i:=0;i<self.stream_count;i++ {
	    log.Printf("Track %s",self.tracks[i].String())
	}*/
	go func() {
		for {
			p := <-self.ch
			//if err  {
			//	println("channel closed")
			//break
			//}
			if p.Size == 0 {
				log.Printf("0 Size packet in multiplexer received for stream%d:", p.Stream)
			}
			if p.avpacket == nil {
				println("nil packet in multiplexer received")
				continue
			}
			stream := self.tracks[p.Stream]
			p.avpacket.size = (_Ctype_int)(p.Size)
			p.avpacket.data = (*C.uint8_t)(unsafe.Pointer(&p.Data[0]))
			p.avpacket.pts = (C.int64_t)(p.Pts.RescaleTo(Rational{int(stream.Stream.time_base.num), int(stream.Stream.time_base.den)}).Time)
			p.avpacket.duration = (_Ctype_int)(p.Duration.RescaleTo(Rational{int(stream.Stream.time_base.num), int(stream.Stream.time_base.den)}).Time)
			p.avpacket.flags = C.int(p.Flags)
			p.avpacket.stream_index = _Ctype_int(p.Stream)
			p.avpacket.dts = C.int64_t(AV_NOPTS_VALUE)
			if p.avpacket.data == nil {
				println("nil packet.data in multiplexer received")
				continue
			}

			result := av_interleaved_write_frame(self.Ds.ctx, &p)
			if result != 0 {
				log.Printf("failed write packet to stream")
				log.Printf("Packet:" + p.String())
			}

			p.Free()
		}
		log.Printf("Multiplexer End")
	}()
}

func (self *Multiplexer) Stop() {
	close(self.ch)
	log.Printf("Writing Trailer")
	av_write_trailer(self.Ds.ctx)
}

func NewMultiplexer(sink *DataSink) *Multiplexer {
	return &Multiplexer{Ds: *sink}
}
