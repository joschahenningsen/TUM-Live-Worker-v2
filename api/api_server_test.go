package api

import (
	"context"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"testing"
)

var mockServer = server{}

func setup() {
	cfg.WorkerID = "123"
}

func TestServer_RequestStream(t *testing.T) {
	setup()
	_, err := mockServer.RequestStream(context.Background(), &pb.StreamRequest{
		WorkerId: "234"})
	if err == nil {
		t.Errorf("Request with wrong WorkerID should be rejected")
		return
	}
}
