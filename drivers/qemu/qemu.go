// Package qemu provides utility functions for interacting with the Qemu machine
// emulator and virtualizer.
package qemu

import "github.com/banditmoscow1337/spos/kernel/sys"

const (
	// port used for Qemu exit commands.
	qemuExitPort = 0x501
)

// Exit sends an exit command to Qemu with the given exit code.
func Exit(code int) {
	sys.Outb(qemuExitPort, byte(code))
}
