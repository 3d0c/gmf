package gmf

import (
	"errors"
)

type DataSource struct {
	Locator MediaLocator
	ctx     *FormatContext
	valid   bool
}

func (src *DataSource) Connect() error {
	src.valid = false
	src.ctx = avformat_alloc_context()
	result := av_open_input_file(src.ctx, src.Locator.Filename, nil, 0, nil)
	if result != 0 {
		return errors.New("file not opened 123" + string(result))
	}
	result = av_find_stream_info(src.ctx)
	if result < 0 {
		return errors.New("could not find stream info")
	}
	src.valid = true
	return nil
}

func (src *DataSource) Disconnect() error {
	if src.valid {
		av_close_input_file(src.ctx)
	}
	return nil
}

func NewDataSource(loc MediaLocator) *DataSource {
	return &DataSource{Locator: loc, ctx: nil, valid: false}
}
