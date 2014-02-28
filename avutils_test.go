package gmf

import (
	"testing"
)

func TestAvError(t *testing.T) {
	if err := AvError(-2); err.Error() != "No such file or directory" {
		t.Fatalf("Expected error is 'No such file or directory', '%s' got\n", err.Error())
	}
}
