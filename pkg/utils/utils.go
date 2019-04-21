package utils

// Byte2Sring convert slice uint8 to string
func Byte2Sring(bs []uint8) string {
	ba := make([]byte, 0, len(bs))
	for _, b := range bs {
		ba = append(ba, byte(b))
	}
	return string(ba)
}
