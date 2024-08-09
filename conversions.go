package working

// Taken from https://0x0f.me/blog/golang-compiler-optimization/
func Bool2byte(b bool) byte {
	// The compiler currently only optimizes this form.
	// See issue 6011.
	var i byte
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// Adapted from https://0x0f.me/blog/golang-compiler-optimization/
func Byte2bool(b byte) bool {
	// The compiler currently only optimizes this form.
	// See issue 6011.
	var i bool
	if b == 0 {
		i = false
	} else {
		i = true
	}
	return i
}
