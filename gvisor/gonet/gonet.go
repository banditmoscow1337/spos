package gonet

import (
	"sync"
	"time"

	"github.com/icexin/eggos/gvisor/stack"
	"github.com/icexin/eggos/gvisor/tcpip"
	"github.com/icexin/eggos/gvisor/waiter"
)

type deadlineTimer struct {
	// mu protects the fields below.
	mu sync.Mutex

	readTimer     *time.Timer
	readCancelCh  chan struct{}
	writeTimer    *time.Timer
	writeCancelCh chan struct{}
}

func (d *deadlineTimer) init() {
	d.readCancelCh = make(chan struct{})
	d.writeCancelCh = make(chan struct{})
}

// A UDPConn is a wrapper around a UDP tcpip.Endpoint that implements
// net.Conn and net.PacketConn.
type UDPConn struct {
	deadlineTimer

	stack *stack.Stack
	ep    tcpip.Endpoint
	wq    *waiter.Queue
}

// NewUDPConn creates a new UDPConn.
func NewUDPConn(s *stack.Stack, wq *waiter.Queue, ep tcpip.Endpoint) *UDPConn {
	c := &UDPConn{
		stack: s,
		ep:    ep,
		wq:    wq,
	}
	c.deadlineTimer.init()
	return c
}
