package gmf

//#include "libavformat/avformat.h"
import "C"

import (
	"fmt"
)

//A Track abstracts the information specific to an individual track in a media stream.
//A media stream might contain multiple media tracks, such as separate tracks for audio, video, and midi data.
//A Track is the output of a Demultiplexer.
type Track struct {
	Stream   *C.AVStream
	stream   chan Packet
	next_pts int64
}

func (self *Track) String() string {
	return fmt.Sprintf("Idx:%d;CTB:%d/%d;STB:%d/%d", self.Stream.index, self.Stream.codec.time_base.num, self.Stream.codec.time_base.den, self.Stream.time_base.num, self.Stream.time_base.den)
}

func (self *Track) GetFormat() Format {
	if self.Stream.codec.codec_type == CODEC_TYPE_VIDEO {
		return VideoFormat{
			Rational{int(self.Stream.r_frame_rate.num), int(self.Stream.r_frame_rate.den)},
			Rational{int(self.Stream.codec.time_base.num), int(self.Stream.codec.time_base.den)},
			int(self.Stream.codec.width),
			int(self.Stream.codec.height)}
	}
	return AudioFormat{
		int(self.Stream.codec.channels),
		int(self.Stream.codec.frame_size),
		int(self.Stream.codec.sample_rate),
		av_get_bits_per_sample_fmt(self.Stream.codec.sample_fmt) / 8}
}

func (self *Track) GetStreamIndex() int {
	return int(self.Stream.index)
}

func (self *Track) GetStartTime() Timestamp {
	return Timestamp{}
}

func (self *Track) ReadPacket(p *Packet) bool {
	result := false
	*p, result = <-self.stream
	return result
}

func (self *Track) WritePacket(p *Packet) bool {
	if self.stream == nil {
		return false
	}
	p.Stream = int(self.Stream.index)

	if self.next_pts > 0 && p.Pts.Time != self.next_pts {
		//log.Printf("Fail: next_pts=%d incoming pts=%d", self.next_pts, p.Pts.Time)
	} else {
		self.next_pts = p.Pts.Time
	}
	self.next_pts += p.Duration.Time
	self.stream <- *p
	return true
}

func (self *Track) GetDecoder() *Decoder {
	coder := Decoder{}                //NewCoder()
	coder.Ctx.ctx = self.Stream.codec //dpx.Ds.Ctx.streams[streamid].codec
	coder.frame_rate = Rational{int(self.Stream.r_frame_rate.num), int(self.Stream.r_frame_rate.den)}
	coder.time_base = Rational{int(self.Stream.codec.time_base.num), int(self.Stream.codec.time_base.den)}
	return &coder
}
