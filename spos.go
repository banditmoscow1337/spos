package spos

import (
	"runtime"

	"github.com/banditmoscow1337/spos/console"
	"github.com/banditmoscow1337/spos/drivers/cga/fbcga"

	//_ "github.com/banditmoscow1337/spos/drivers/e1000"
	"github.com/banditmoscow1337/spos/drivers/kbd"
	"github.com/banditmoscow1337/spos/drivers/pci"
	"github.com/banditmoscow1337/spos/drivers/ps2/mouse"
	"github.com/banditmoscow1337/spos/drivers/uart"
	"github.com/banditmoscow1337/spos/drivers/vbe"
	"github.com/banditmoscow1337/spos/fs"

	//"github.com/banditmoscow1337/spos/inet"
	"github.com/banditmoscow1337/spos/kernel"
)

func kernelInit() {
	// trap and syscall threads use two Ps,
	// and the remainings are for other goroutines
	runtime.GOMAXPROCS(6)

	kernel.Init()
	uart.Init()
	kbd.Init()
	mouse.Init()
	console.Init()

	fs.Init()
	vbe.Init()
	fbcga.Init()
	pci.Init()
	//inet.Init()
}

func init() {
	kernelInit()
}
