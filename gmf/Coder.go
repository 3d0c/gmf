package gmf

//#include <stdint.h>
import "C"
import "unsafe"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
)

type coder struct {
	Parameter     map[string]string
	Ctx           _CodecContext
	Codec         _Codec
	Valid         bool
	initialized   bool
	codec_type    int
	Pixelformat   int
	Sampleformat  int
	pre_allocated bool
	ExtraData     []byte
}

func (self *coder) init() {
	if self.Parameter == nil {
		println("creating Parameter map")
		self.Parameter = make(map[string]string)
		self.pre_allocated = true
	}
	self.initialized = true
}

func (self *coder) SetParameter(key, val string) {
	self.init()
	self.Parameter[key] = val
}

func (self *coder) GetParameter(key string) string {
	self.init()
	return self.Parameter[key]
}

func (self *coder) GetParameters() map[string]string {
	self.init()
	var op = new(avOption)
	for true {
		op = av_next_option(&self.Ctx, op)
		if op.opt == nil {
			break
		}
		self.Parameter[op.Name] = av_get_string(&self.Ctx, op.Name)
	}
	self.Parameter["codecid"] = fmt.Sprintf("%d", self.Ctx.ctx.codec_id)
	self.Parameter["width"] = fmt.Sprintf("%d", self.Ctx.ctx.width)
	self.Parameter["height"] = fmt.Sprintf("%d", self.Ctx.ctx.height)
	self.Parameter["channels"] = fmt.Sprintf("%d", self.Ctx.ctx.channels)
	// self.Parameter["sample_rate"] = fmt.Sprintf("%d", self.Ctx.ctx.sample_rate)
	return self.Parameter
}

func (self *coder) GetExtraData() []byte {
	if self.Ctx.ctx.extradata_size > 0 {
		self.ExtraData = make([]byte, self.Ctx.ctx.extradata_size+8)
		data := (*(*[1 << 30]byte)(unsafe.Pointer(self.Ctx.ctx.extradata)))[0:self.Ctx.ctx.extradata_size]
		for i := 0; i < int(self.Ctx.ctx.extradata_size); i++ {
			self.ExtraData[i] = data[i]
		}
	}
	return self.ExtraData
}

func (self *coder) SetExtraData(extra []byte) {
	self.ExtraData = extra

	self.Ctx.ctx.extradata = (*C.uint8_t)(unsafe.Pointer(&self.ExtraData[0]))
}

func (c *coder) Close() {
	log.Printf("Closing Coder")
	if c.Valid {
		avcodec_close(c.Ctx)
		if !c.pre_allocated {
			//av_free_codec_context(c)
		}
	}
}

func (c *coder) Free() {
	log.Printf("Freeing Coder")
	if c.Valid {
		//avcodec_close(c.Ctx)
		if !c.pre_allocated {
			av_free_codec_context(c)
		}
	}
}

func (c *coder) open(t int) {
	c.codec_type = t

	c.init()
	c.prepare()
	if c.codec_type == CODEC_TYPE_DECODER {
		c.Codec = avcodec_find_decoder(int32(c.Ctx.ctx.codec_id))
		//    c.pre_allocated=false
	}

	if c.codec_type == CODEC_TYPE_ENCODER {
		//log.Printf("first prepare")
		//c.prepare()
		c.Codec = avcodec_find_encoder(int32(c.Ctx.ctx.codec_id))
		//c.Ctx.ctx.sample_fmt=0
		c.prepare()
		//    c.Ctx.ctx.time_base.num=1//int32(c.Pixelformat)
		//    c.Ctx.ctx.time_base.den=44100//int32(c.Pixelformat)

		//log.Printf("SampleFormat %d", c.Ctx.ctx.pix_fmt)
	}
	//c.Ctx.ctx.request_channels = 2;
	//c.Ctx.ctx.request_channel_layout = 2;

	if c.Codec.codec == nil {
		log.Printf("could not find Codec for id %d", c.Ctx.ctx.codec_id)
		return
	}
	//c.Ctx.ctx.codec_id=c.Codec.codec.codec_id
	//if(c.Ctx.ctx.codec_id<0x10000){
	c.Ctx.ctx.codec_type = c.Codec.codec._type

	//}
	//c.Ctx.ctx.width=320
	//c.Ctx.ctx.height=240
	//avcodec_get_context_defaults2(&c.Ctx, c.Codec);

	c.setCodecParams()
	//c.Ctx.ctx.time_base.num=1
	//c.Ctx.ctx.time_base.den=25
	//c.GetParameters()
	c.Ctx.ctx.pix_fmt = 0    //int32(c.Pixelformat)
	c.Ctx.ctx.sample_fmt = 1 //int32(c.Pixelformat)
	res := avcodec_open(c.Ctx, c.Codec)

	if res < 0 {
		log.Printf("error openning codec for id %d\n", c.Ctx.ctx.codec_id)
		return
	}
	log.Printf("codec openned for id %d\n", c.Ctx.ctx.codec_id)
	c.Valid = true
}

func (self *coder) prepare() {
	if self.Ctx.ctx == nil {
		println("Alloc Encoder Context")
		cid, _ := strconv.Atoi(self.Parameter["codecid"])
		if cid == 0 {
			if self.codec_type == CODEC_TYPE_ENCODER {
				self.Codec = avcodec_find_encoder_by_name(self.Parameter["codecid"])
			}
			if self.codec_type == CODEC_TYPE_DECODER {
				self.Codec = avcodec_find_decoder_by_name(self.Parameter["codecid"])
			}
			if self.Codec.codec != nil {
				cid = int(self.Codec.codec.id) //,_:=strconv.Atoui64(self.Parameter["codecid"])
			}
		}
		log.Printf("searching for codec id %d", cid)
		self.Ctx = *avcodec_alloc_context()
		if len(self.ExtraData) > 0 {
			// self.Ctx.ctx.extradata = (*_Ctypedef_uint8_t)(unsafe.Pointer(&self.ExtraData[0]))
			self.Ctx.ctx.extradata = (*C.uint8_t)(unsafe.Pointer(&self.ExtraData[0]))
		}

		self.pre_allocated = false
		self.Ctx.ctx.codec_id = uint32(cid)
		runtime.SetFinalizer(self, close_coder)
	} else {
		//c.pre_allocated=true
	}
	width, err := strconv.Atoi(self.Parameter["width"])
	if err == nil {
		self.Ctx.ctx.width = _Ctype_int(width)
		log.Printf("setting width to %d", self.Ctx.ctx.width)
	} else {
		//    println(err.String())
	}
	height, err := strconv.Atoi(self.Parameter["height"])
	if err == nil {
		self.Ctx.ctx.height = _Ctype_int(height)
		log.Printf("setting height to %d", self.Ctx.ctx.height)
	} else {
		//    println(err.String())
	}
	channels, err := strconv.Atoi(self.Parameter["ac"])
	if err == nil {
		self.Ctx.ctx.channels = _Ctype_int(channels)
		log.Printf("setting channels to %d", self.Ctx.ctx.channels)
	} else {
		//    println(err.String())
	}
	self.Ctx.ctx.debug = 0
	self.Ctx.ctx.debug_mv = 0
	/**
	 * @TODO: settng the fixed params like width, height, channels...
	 */
}

func (self *coder) setCodecParams() {
	for mkey, mval := range self.Parameter {
		if mval != "" && mkey != "width" && mkey != "height" && mkey != "codecid" && mkey != "channels" {
			//fmt.Printf("Setting Context String '%s' = '%s' len(val)=%d\n",mkey,mval,len(mval))
			av_set_string(&self.Ctx, mkey, mval)
		}
	}
	if self.Ctx.ctx.codec_type == CODEC_TYPE_AUDIO {
		self.Ctx.ctx.time_base.num = 1
		self.Ctx.ctx.time_base.den = self.Ctx.ctx.sample_rate
	}
}

func close_coder(c *coder) {
	c.Free()
}
