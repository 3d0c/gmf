package gmf_test

import (
	// "log"
	"github.com/3d0c/gmf"
	"testing"
)

func TestImg(t *testing.T) {
	img, err := gmf.NewImage(100, 100, gmf.AV_PIX_FMT_YUV420P, 1)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if img.Size() != 15000 {
		t.Fatalf("Expected bufsize = 15000, %d got\n", img.Size())
	}

	img.Free()
}
