package linux

const (
	ARCH_SET_FS = 0x1002
)

// Timespec represents struct timespec in <time.h>.
//
// +marshal slice:TimespecSlice
type Timespec struct {
	Sec  int64
	Nsec int64
}
