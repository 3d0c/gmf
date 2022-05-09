package gmf_test

import (
	"github.com/3d0c/gmf"
	"log"
	"testing"
)

func TestCodecEq(t *testing.T) {
	dec, enc, notfound := 0, 0, 0

	for _, codec := range gmf.Codecs {
		if codec.IsEncoder {
			encById, err := gmf.FindEncoder(codec.Id())
			if err != nil {
				log.Println("E", err, codec.Id())
				notfound++
				continue
			}

			encByName, err := gmf.FindEncoder(codec.Name())
			if err != nil {
				log.Println("E", err, codec.Id())
				notfound++
				continue
			}

			if encById.Id() != encByName.Id() {
				t.Fatal("different id. should be equal:", encById.Id(), encByName.Id())
			}

			enc++
		} else {
			decById, err := gmf.FindDecoder(codec.Id())
			if err != nil {
				log.Println("D", err, codec.Id())
				notfound++
				continue
			}

			decByName, err := gmf.FindDecoder(codec.Name())
			if err != nil {
				log.Println("D", err, codec.Id())
				notfound++
				continue
			}

			if decById.Id() != decByName.Id() {
				t.Fatal("different id. should be equal:", decById.Id(), decByName.Id())
			}

			dec++
		}
	}

	log.Printf("%d encoders, %d decoders checked. %d not found", enc, dec, notfound)
}

func TestFindByName(t *testing.T) {
	c, err := gmf.FindEncoder("libx264")
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("Found %s, %s\n", c.Name(), c.LongName())
}
