package util

import "math"

func BytesToGigabyte(b uint64) uint64 {
	return b / uint64(math.Pow(2, 30))
}
