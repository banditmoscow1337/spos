package kernel

import (
	"github.com/banditmoscow1337/spos/drivers/multiboot"
	"github.com/banditmoscow1337/spos/drivers/pic"
	"github.com/banditmoscow1337/spos/drivers/uart"
	"github.com/banditmoscow1337/spos/kernel/mm"
)

//go:nosplit
func rt0()

//go:nosplit
func go_entry()

//go:nosplit
func wrmsr(reg uint32, value uintptr)

//go:nosplit
func rdmsr(reg uint32) (value uintptr)

//go:nosplit
func preinit(magic, mbiptr uintptr) {
	simdInit()
	gdtInit()
	idtInit()
	multiboot.Init(magic, mbiptr)
	mm.Init()
	uart.PreInit()
	syscallInit()
	trapInit()
	threadInit()
	pic.Init()
	timerInit()
	schedule()
}
