package gmf_test

import (
	"testing"

	"github.com/3d0c/gmf"
	"github.com/stretchr/testify/require"
)

func TestAvError(t *testing.T) {
	if err := gmf.AvError(-2); err.Error() != "No such file or directory" {
		t.Fatalf("Expected error is 'No such file or directory', '%s' got\n", err.Error())
	}
}

func TestIsTimeCodeBetween(t *testing.T) {
	timeCodeStart := "20:31:00:62"
	timeCodeEnd := "04:56:12:00"
	tests := []struct {
		timeCodeToTest string
		expectedBool   bool
		errorExpected  bool
	}{
		{"20:15:31:00", false, false},
		{"20:30:59:99", false, false},
		{"20:30:59:99", false, false},
		{"20:31:00:00", false, false},
		{"20:31:00:61", false, false},
		{"20:31:31:62", true, false},
		{"04:31:31:00", true, false},
		{"04:56:12:00", true, false},
		{"04:56:12:01", false, false},
		{"05:31:31:00", false, false},
		{"05:31:31", false, true},
	}
	for _, test := range tests {
		isTimeCodeBetween, err := gmf.IsTimeCodeBetween(test.timeCodeToTest, timeCodeStart, timeCodeEnd)
		if test.errorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, test.expectedBool, isTimeCodeBetween, "hourToTest: %s", test.timeCodeToTest)
	}
}
