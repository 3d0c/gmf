package gmf

import (
	"errors"
	"unsafe"
)

type DataSink struct {
	Locator MediaLocator
	ctx     *FormatContext
	Valid   bool
}

func (src *DataSink) Connect() error {
	src.Valid = false
	src.ctx = avformat_alloc_context()
	format := av_guess_format(src.Locator.Format, src.Locator.Filename)
	src.ctx.ctx.oformat = (*_Ctype_struct_AVOutputFormat)(unsafe.Pointer(format.format))

	result := url_fopen(src.ctx, src.Locator.Filename)

	if result != 0 {
		return errors.New("file not opened")
	}
	src.ctx.ctx.preload = 500000
	src.ctx.ctx.max_delay = 700000
	src.ctx.ctx.loop_output = -1
	src.ctx.ctx.flags |= 0x0004
	src.ctx.ctx.mux_rate = 0
	src.ctx.ctx.packet_size = 0
	src.Valid = true
	return nil
}

func (src *DataSink) Disconnect() error {
	if src.Valid {
		url_fclose(src.ctx)
	}
	av_free_format_context(src.ctx)
	return nil
}

func (src *DataSink) GetLocator() MediaLocator {
	return src.Locator
}

func NewDatasink(loc MediaLocator) *DataSink {
	return &DataSink{Locator: loc, ctx: nil, Valid: false}
}
