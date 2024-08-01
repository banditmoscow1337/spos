package main

import (
	"io"
	_ "net/http/pprof"
	"os"
	"runtime"

	_ "github.com/banditmoscow1337/spos"
	"github.com/banditmoscow1337/spos/app/sh"
	"github.com/banditmoscow1337/spos/console"
	"github.com/banditmoscow1337/spos/log"
)

func main() {
	log.Infof("[runtime] go version:%s", runtime.Version())
	log.Infof("[runtime] args:%v", os.Args)
	w := console.Console()
	io.WriteString(w, "\nwelcome to spos\n")
	sh.Bootstrap()
}
