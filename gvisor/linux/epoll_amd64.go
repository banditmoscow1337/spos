package linux

// EpollEvent is equivalent to struct epoll_event from epoll(2).
//
// +marshal slice:EpollEventSlice
type EpollEvent struct {
	Events uint32
	// Linux makes struct epoll_event::data a __u64. We represent it as
	// [2]int32 because, on amd64, Linux also makes struct epoll_event
	// __attribute__((packed)), such that there is no padding between Events
	// and Data.
	Data [2]int32
}
