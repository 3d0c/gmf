package gmf

import (
	"errors"
)

type Resampler struct {
	ctx         *ResampleContext
	isize       int
	osize       int
	outbuffer   []byte
	channels    int
	sample_rate int
	dec         *Decoder
	enc         *Encoder
}

func (self *Resampler) Init(dec *Decoder, enc *Encoder) error {
	self.ctx = av_audio_resample_init(
		int(enc.Ctx.ctx.channels),
		int(dec.Ctx.ctx.request_channels),
		int(enc.Ctx.ctx.sample_rate),
		int(dec.Ctx.ctx.sample_rate),
		int(enc.Ctx.ctx.sample_fmt),
		int(dec.Ctx.ctx.sample_fmt))
	if self.ctx.ctx == nil {
		return errors.New("Could not create resample context!")
	}
	self.isize = av_get_bits_per_sample_fmt(dec.Ctx.ctx.sample_fmt) / 8
	self.osize = av_get_bits_per_sample_fmt(enc.Ctx.ctx.sample_fmt) / 8
	self.outbuffer = make([]byte, (2*128*1024)+8)
	self.channels = int(enc.Ctx.ctx.channels)
	self.sample_rate = int(enc.Ctx.ctx.sample_rate)
	self.dec = dec
	self.enc = enc
	return nil
}

func (self *Resampler) Resample(f *Frame) *Frame {
	frame := new(Frame)
	if self.ctx.ctx == nil {
		return f
	}

	out_size := audio_resample(self.ctx, self.outbuffer, f.buffer, (f.size / (self.channels * self.isize)))
	frame.buffer = self.outbuffer
	frame.size = out_size * self.channels * self.osize
	frame.Duration = Timestamp{int64(out_size), Rational{1, self.sample_rate}}

	last_insamples := av_rescale_q(int64(f.size/self.isize), Rational{1, int(self.dec.Ctx.ctx.sample_rate)}, Rational{1, int(self.enc.Ctx.ctx.sample_rate)})
	last_outsamples := int64(frame.size / self.osize)
	delta := av_clip(int(last_insamples-last_outsamples), -2, 2)
	av_resample_compensate(self.ctx, delta, int(self.enc.Ctx.ctx.sample_rate/2))
	return frame
}

func (self *Resampler) Close() {
	audio_resample_close(self.ctx)
}
