package util

import (
	"math"
	"testing"
)

func TestByteToGigabyteIsValid(t *testing.T) {
	got := BytesToGigabyte(uint64(math.Pow(2, 30) * 5))
	if got != uint64(5) {
		t.Errorf("BytesToGigabyte(%d) = %d; wants 5", uint64(math.Pow(2, 30)*5), got)
	}
}
