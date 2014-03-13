package gmf

import (
	// "log"
	"testing"
)

func TestImg(t *testing.T) {
	img, err := NewImage(100, 100, AV_PIX_FMT_YUV420P, 1)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if img.bufsize != 15000 {
		t.Fatalf("Expected bufsize = 15000, %d got\n", img.bufsize)
	}

	img.Free()
}
