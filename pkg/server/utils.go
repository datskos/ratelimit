package server

func min(a, b uint32) uint32 {
	if a > b {
		return b
	}
	return a
}
