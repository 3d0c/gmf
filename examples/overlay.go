package main

/*

#cgo pkg-config: libavcodec libavutil
#cgo CFLAGS: -I/usr/local/include/opencv
#cgo CFLAGS: -I/usr/local/include/opencv2
#cgo LDFLAGS: -lopencv_imgproc -lopencv_core -lopencv_highgui


#include <stdio.h>
#include <strings.h>
#include <stdlib.h>
#include <stdint.h>

#include "libavcodec/avcodec.h"
#include "libavutil/frame.h"
#include "opencv2/highgui/highgui_c.h"
#include "cv.h"

typedef struct {
    unsigned char *data;
    unsigned int length;
} Blob;

void blender(AVFrame *frame, Blob *imgdata, CvRect *roi) {
    // init backgroud
    int flen = frame->width*frame->height*3;
	CvMat *bgMat = cvCreateMat(1, flen, CV_8UC1);
    bgMat->data.ptr = frame->data[0];

	// init foreground
    CvMat *fgMat = cvCreateMat(1, imgdata->length, CV_8UC1);
    fgMat->data.ptr = imgdata->data;

    // CvMat *fgMat = cvDecodeImageM(fgBuf, CV_LOAD_IMAGE_UNCHANGED);
    // cvReleaseMat(&fgBuf);

	int loc_x = (roi->x > 0) ? roi->x : 0;
    int loc_y = (roi->y > 0) ? roi->y : 0;

    int y, x, c;

    for(y = loc_y; y < bgMat->rows; y++) {
        int fY = y - roi->y;

        if(fY >= fgMat->rows) {
            break;
        }

        for(x = loc_x; x < bgMat->cols; x++) {
            int fX = x - roi->x;

            if(fX >= fgMat->cols) {
                break;
            }

            double opacity;

			opacity = ((double)fgMat->data.ptr[fY * fgMat->step + fX * 3 + 3]) / 255.;


            for(c = 0; opacity > 0 && c < bgMat->nChannels; c++) {
                unsigned char foregroundPx = fgMat->data.ptr[fY * fgMat->step + fX * fgMat->nChannels + c];
                unsigned char backgroundPx = bgMat->data.ptr[y * bgMat->step + x * bgMat->nChannels + c];
                resultMat->data.ptr[y*resultMat->step + bgMat->nChannels*x + c] = backgroundPx * (1.-opacity) + foregroundPx * opacity;
            }
        }
    }
}

*/
import "C"

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/3d0c/gmf"
)

type CvRect _Ctype_CvRect

func main() {
	var (
		img, dst       string
		octx           *gmf.FmtCtx
		swsCtx, swsYUV *gmf.SwsCtx
		err            error
		width, height  int = 640, 480
	)

	flag.StringVar(&img, "img", "", "overlay with image, e.g. -img=logo.png")
	flag.StringVar(&dst, "dst", "", "destination file, e.g. -dst=result.mp4")
	flag.Parse()

	if img == "" || dst == "" {
		log.Fatalf("-img or -dst not set\n")
	}

	f, err := os.OpenFile(img, 0644, os.O_RDONLY)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	defer f.Close()

	imgdata := ioutil.ReadAll(f)

	if octx, err = gmf.NewOutputCtx(dst); err != nil {
		log.Fatalf("%s\n", err)
	}

	codec, err := gmf.FindEncoder("libx264")
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	encCtx := NewCodecCtx(codec)
	if encCtx == nil {
		log.Fatalf("unable to create codec context\n")
	}
	defer encCtx.Free()

	encCtx.
		SetBitRate(100000).
		SetWidth(width).
		SetHeight(height).
		SetTimeBase(gmf.AVR{Num: 1, Den: 25}).
		SetPixFmt(gmf.AV_PIX_FMT_YUV420P).
		SetProfile(gmf.FF_PROFILE_MPEG4_SIMPLE).
		SetMbDecision(gmf.FF_MB_DECISION_RD)

	ost := octx.NewStream(codec)
	if ost == nil {
		log.Fatalf("Unable to create output stream for %s\n", codec.Name())
	}
	defer ost.Free()

	ost.SetCodecCtx(encCtx)

	if swsCtx = NewSwsCtxExplicit(width, height, encCtx.PixFmt(), width, height, gmf.AV_PIX_FMT_RGB24, 0); swsCtx == nil {
		log.Fatalf("Unable to create sws context\n")
	}

	if swsYUV = NewSwsCtxExplicit(width, height, gmf.AV_PIX_FMT_RGB24, width, height, encCtx.PixFmt(), 0); swsCtx == nil {
		log.Fatalf("Unable to create sws context\n")
	}

	if err = octx.WriteHeader(); err != nil {
		log.Fatalf("%s\n", err)
	}

	var (
		srcFrame *gmf.Frame
		rgbFrame *gmf.Frame = gmf.NewFrame().SetWidth(width).SetHeight(heigh).SetFormat(gmf.AV_PIX_FMT_BGR24)
	)

	if err = rgbFrame.ImgAlloc(); err != nil {
		log.Fatalf("%s\n", err)
	}

	for frame = range GenSyntVideoNewFrame(videoEncCtx.Width(), videoEncCtx.Height(), videoEncCtx.PixFmt()) {
		frame.SetPts(i)

		swsCtx.Scale(srcFrame, rgbFrame)
		blender(rgbFrame, (*C.uchar)(unsafe.Pointer(&imgdata[0])), &CvRect{0, 0, 0, 0})

		if p, err := frame.Encode(ost.CodecCtx()); p != nil {
			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			}

			n++

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", frame.Pts(), p.Size(), p.Pts(), p.Dts())

			Release(p)
		} else if err != nil {
			fatal(err)
		}

		Release(frame)
		i++

	}
}
