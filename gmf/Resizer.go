package gmf

//import "log"
//import "unsafe"
type Resizer struct {
	ctx    *SwsContext
	width  int
	height int
	fmt    int
	//    frame Frame
	//    buffer []byte
}


func (self *Resizer) Init(dec *Decoder, enc *Encoder) {
	self.ctx = new(SwsContext)
	sws_scale_getcontext(self.ctx, int(dec.Ctx.ctx.width), int(dec.Ctx.ctx.height), int(dec.Ctx.ctx.pix_fmt), int(enc.Ctx.ctx.width), int(enc.Ctx.ctx.height), int(enc.Ctx.ctx.pix_fmt), 1)
	self.width = int(enc.Ctx.ctx.width)
	self.height = int(enc.Ctx.ctx.height)
	self.fmt = int(enc.Ctx.ctx.pix_fmt)
}

func (self *Resizer) Resize(in *Frame) *Frame {
	frame := NewFrame(self.fmt, self.width, self.height)

	if result := sws_scale(self.ctx, in, frame); result <= 0 {
		//log.Printf("failed to resize the image")
	} else {
		//log.Printf("frame result=%d", result)
	}
	frame.Pts = in.Pts
	frame.Duration = in.Duration
	return frame
}

func (self *Resizer) Close() {
	sws_free_context(self.ctx)
}
