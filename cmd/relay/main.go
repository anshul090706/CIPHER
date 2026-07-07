package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", port),
		),
		libp2p.Transport(ws.New),
	)

	if err != nil {
		panic(err)
	}

	relayService, err := relay.New(h)
	if err != nil {
		panic(err)
	}
	defer relayService.Close()

	fmt.Println("========================================")
	fmt.Println("CIPHER RELAY STARTED")
	fmt.Println("Relay Peer ID:", h.ID())
	fmt.Println("Port:", port)
	fmt.Println("Circuit Relay v2 service: ACTIVE")
	fmt.Println("========================================")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	h.Close()
}