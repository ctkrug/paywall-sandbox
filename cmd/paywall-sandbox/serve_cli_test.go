package main

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

// freeAddr returns a loopback address on a port that was free at the time
// of the call, for a subprocess to bind. Inherently racy (nothing stops
// another process from grabbing it first) but good enough for a test.
func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freeAddr: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func TestCLIServeChallengesConfiguredRoute(t *testing.T) {
	addr := freeAddr(t)
	cmd := exec.Command(binPath, "serve", "--addr", addr, "--path", "/paid", "--amount", "100", "--asset", "USDC")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start serve: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	url := fmt.Sprintf("http://%s/paid", addr)
	var resp *http.Response
	var err error
	for i := 0; i < 50; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		t.Errorf("GET %s status = %d, want %d", url, resp.StatusCode, http.StatusPaymentRequired)
	}
	if resp.Header.Get("X-Payment-Required") == "" {
		t.Error("response missing X-Payment-Required header")
	}
}
