package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/1amKhush/CIPHER/pkg/logger"
	"github.com/1amKhush/CIPHER/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	providerAddr := flag.String("provider", "", "Multiaddr of the provider to connect to")
	port := flag.Int("port", 9001, "Port to listen on")
	flag.Parse()

	if *providerAddr == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -provider <multiaddr>\n", os.Args[0])
		os.Exit(1)
	}

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
		PrivKeyPath: "client_key.key",
		EnableMDNS:  true,
	}

	h, err := p2p.NewHost(ctx, opts)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start client host")
	}
	defer h.Close()

	logger.Info().Str("peer_id", h.ID().String()).Msg("Client started")

	// Parse provider multiaddr
	maddr, err := ma.NewMultiaddr(*providerAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid provider multiaddr")
	}

	// Extract peer ID from multiaddr
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to extract peer info from multiaddr")
	}

	// Connect to provider
	logger.Info().Str("provider", info.ID.String()).Msg("Connecting to provider...")
	if err := h.Connect(ctx, *info); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to provider")
	}
	logger.Info().Msg("Connected!")

	// Open stream
	s, err := h.NewStream(ctx, info.ID, p2p.ProtocolID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open stream to provider")
	}

	msg := fmt.Sprintf("Hello from client %s!", h.ID())
	if _, err := s.Write([]byte(msg)); err != nil {
		logger.Fatal().Err(err).Msg("Failed to write to stream")
	}

	// Read reply
	buf := make([]byte, 1024)
	n, err := s.Read(buf)
	if err != nil && err != io.EOF {
		logger.Fatal().Err(err).Msg("Failed to read reply")
	}

	logger.Info().Str("reply", string(buf[:n])).Msg("Received reply from provider")
}
