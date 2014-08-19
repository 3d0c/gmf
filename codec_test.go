package gmf

import (
	"log"
	"testing"
)

func TestCodec(t *testing.T) {
	dec, enc, notfound := 0, 0, 0

	for _, codec := range Codecs {
		if codec.IsEncoder {
			encById, err := FindEncoder(codec.Id())
			if err != nil {
				// log.Println(err)
				notfound++
				continue
			}

			encByName, err := FindEncoder(codec.Name())
			if err != nil {
				// log.Println(err)
				notfound++
				continue
			}

			if encById.Id() != encByName.Id() {
				t.Fatal("different id. should be equal:", encById.Id(), encByName.Id())
			}

			enc++
		} else {
			decById, err := FindDecoder(codec.Id())
			if err != nil {
				// log.Println(err)
				notfound++
				continue
			}

			decByName, err := FindDecoder(codec.Name())
			if err != nil {
				// log.Println(err)
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
