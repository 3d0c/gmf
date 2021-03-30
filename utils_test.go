package gmf_test

import (
	"github.com/3d0c/gmf"
	"testing"
)

func TestAvError(t *testing.T) {
	if err := gmf.AvError(-2); err.Error() != "No such file or directory" {
		t.Fatalf("Expected error is 'No such file or directory', '%s' got\n", err.Error())
	}
}
