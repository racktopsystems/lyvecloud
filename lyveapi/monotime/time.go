package monotime

import (
	"time"
	_ "unsafe"
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

func Monotonic() time.Duration {
	return time.Duration(nanotime())
}
