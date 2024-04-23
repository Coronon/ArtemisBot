package util

func FindLeftmostSetBit(n int) int {
	if n == 0 {
		return -1
	}

	var pos int
	for n > 0 {
		n >>= 1
		pos++
	}

	return pos
}
