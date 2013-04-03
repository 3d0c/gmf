package gmf


type Deinterlacer struct {
	dec *Decoder
}


func (self *Deinterlacer) Init(dec *Decoder) {
	self.dec = dec
}

func (self *Deinterlacer) Deinterlace(frame *Frame) *Frame {
	if frame.avframe.interlaced_frame == 0 {
		return frame
	}
	var result *Frame = NewFrame(int(self.dec.Ctx.ctx.pix_fmt), int(self.dec.Ctx.ctx.width), int(self.dec.Ctx.ctx.height))
	avpicture_deinterlace(result, frame, int(self.dec.Ctx.ctx.pix_fmt), int(self.dec.Ctx.ctx.width), int(self.dec.Ctx.ctx.height))
	return result
}
