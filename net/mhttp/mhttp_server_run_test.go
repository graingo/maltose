package mhttp_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/require"
)

func TestRunReturnsAfterProgrammaticStop(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	server := mhttp.New()
	server.SetAddress(address)
	server.GET("/run-stop", func(request *mhttp.Request) {
		request.String(http.StatusOK, "ok")
	})

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		server.Run()
	}()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(2 * time.Second)
	for {
		response, requestErr := client.Get("http://" + address + "/run-stop")
		if requestErr == nil {
			response.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start: %v", requestErr)
		}
		time.Sleep(10 * time.Millisecond)
	}

	require.NoError(t, server.Stop(context.Background()))
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run did not return after Stop")
	}
}
