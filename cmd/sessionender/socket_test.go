package main

import (
	"testing"
	"net"
	"path/filepath"
	"os"
)

func TestSetupkeepalive(t *testing.T) {
	dir := t.TempDir()
	socketpath := filepath.Join(dir, "sessionender")

	if err := os.WriteFile(socketpath, []byte{}, 0500); err != nil {
		t.Fatalf("can't write file: %v", err)
	}
	
	c, err := setupkeepalive(dir)
	if err != nil {
		t.Fatalf("can't make a listener: %v", err)
	}

	fi, err := os.Stat(socketpath)
	if err != nil {
		t.Fatalf("can't stat the socket? %v", err)
	}

	if got, want := fi.Mode(), os.FileMode(os.ModeSocket | 0666); got != want {
		t.Errorf("socket has wrong mode got %v, want %v", got, want)
	}

	conn, err := net.Dial("unix", socketpath)
	if err != nil {
		t.Fatalf("can't dial: %v", err)
	}

	if _, err := conn.Write([]byte("foob")); err != nil {
		t.Fatalf("can't write: %v", err)
	}

	// A rea test would worry about timeouts and such.
	if got, want := <-c, 1;  got != want {
		t.Fatalf("channel wasn't notified")
	}
}

