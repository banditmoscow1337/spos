package tests

import (
	"net"
	"net/http"
	"testing"

	_ "github.com/banditmoscow1337/spos"
)

func TestHTTP(t *testing.T) {
	server := http.Server{}

	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go server.Serve(listener)

	resp, err := http.Get("http://10.0.2.15")
	if err != nil {
		t.Error(err)
		return
	}
	resp.Body.Close()
}
