package httpserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"gokick/app/core/testkit"
)

func TestServeAnswersUntilCancelled(t *testing.T) {
	listener := listen(t)
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	go func() { stopped <- serveListener(ctx, listener, handler, testkit.DiscardLogs()) }()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+listener.Addr().String(), http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()

	if string(body) != "ok" {
		t.Errorf("body %q", body)
	}

	cancel()

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("stopped with %v", err)
		}
	case <-time.After(shutdownTimeout):
		t.Fatal("the server did not stop")
	}
}

func TestServerLimitsSlowAndIdleConnections(t *testing.T) {
	server := newServer(http.NotFoundHandler(), testkit.DiscardLogs())

	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 10*time.Second || server.WriteTimeout != 10*time.Second {
		t.Errorf("read header %s, read %s, write %s", server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout)
	}

	if traefikIdleConnTimeout := 90 * time.Second; server.IdleTimeout <= traefikIdleConnTimeout {
		t.Errorf("idle %s must outlive the idle connections Traefik reuses", server.IdleTimeout)
	}
}

func listen(t *testing.T) net.Listener {
	t.Helper()
	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	return listener
}
