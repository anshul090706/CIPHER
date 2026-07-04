package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/1amKhush/CIPHER/pkg/logger"
	"github.com/1amKhush/CIPHER/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/network"
)

func main() {
	port := flag.Int("port", 9000, "Port to listen on")
	flag.Parse()

	// Initialize logger
	logCfg := logger.DefaultConfig()
	logCfg.Level = "debug"
	if err := logger.Init(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := p2p.HostOptions{
		ListenPort:  *port,
		PrivKeyPath: "provider_key.key",
		EnableMDNS:  true,
	}

	h, err := p2p.NewHost(ctx, opts)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start provider host")
	}
	defer h.Close()

	logger.Info().Str("peer_id", h.ID().String()).Msg("Provider started")
	
	// Print listen addresses
	fmt.Println("Listening on:")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID())
	}

	// Register a simple stream handler for local loopback verification
	h.SetStreamHandler(p2p.ProtocolID, func(s network.Stream) {
		logger.Info().Str("remote_peer", s.Conn().RemotePeer().String()).Msg("Incoming stream")
		
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			logger.Error().Err(err).Msg("Failed to read from stream")
			s.Reset()
			return
		}

		msg := string(buf[:n])
		logger.Info().Str("msg", msg).Msg("Received message")

		// Echo back
		reply := fmt.Sprintf("Hello from provider %s! I received: %s", h.ID(), msg)
		if _, err := s.Write([]byte(reply)); err != nil {
			logger.Error().Err(err).Msg("Failed to write to stream")
			s.Reset()
			return
		}

		s.Close()
	})

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info().Msg("Shutting down provider")
}
