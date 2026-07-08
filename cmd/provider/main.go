package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/1amKhush/CIPHER/pkg/chunker"
	"github.com/1amKhush/CIPHER/pkg/crypto"
	"github.com/1amKhush/CIPHER/pkg/engine"
	"github.com/1amKhush/CIPHER/pkg/logger"
	"github.com/1amKhush/CIPHER/pkg/p2p"
)

func main() {
	port := flag.Int("port", 9000, "Port to listen on")
	relayAddr := flag.String("relay", "", "Relay multiaddr to connect to (optional)")
	verbose := flag.Bool("verbose", false, "Enable verbose debug logging")
	enableQUIC := flag.Bool("quic", false, "Enable QUIC transport")
	flag.Parse()

	cfg := logger.DefaultConfig()
	if *verbose {
		cfg.Level = "debug"
	}
	logger.Init(cfg)

	// 1. Create a dummy file for MVP if it doesn't exist
	fileName := "test_file.txt"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		dummyData := make([]byte, 100*1024) // 100 KB
		if _, err := rand.Read(dummyData); err != nil {
			logger.Fatal().Err(err).Msg("Failed to generate dummy file contents")
		}
		if err := os.WriteFile(fileName, dummyData, 0644); err != nil {
			logger.Fatal().Err(err).Msg("Failed to write dummy file")
		}
		logger.Info().Msgf("Created dummy file %s (100KB)", fileName)
	} else {
		logger.Info().Msgf("Using existing file %s", fileName)
	}

	// 2. Chunk the file
	chunks, err := chunker.ChunkFile(fileName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to chunk file")
	}

	// 3. Create Merkle Tree
	var leaves [][32]byte
	for _, c := range chunks {
		var fileID [32]byte // zeroed
		length := uint32(len(c.Data))
		leaf := crypto.MerkleLeaf(fileID, c.Index, length, c.Data)
		leaves = append(leaves, leaf)
	}
	tree := chunker.NewMerkleTree(leaves)

	var fileID [32]byte // zeroed
	store := &engine.ChunkStore{
		FileID:     fileID,
		Chunks:     chunks,
		MerkleTree: tree,
	}

	rootHex := hex.EncodeToString(tree.Root[:])
	logger.Info().Msgf("File processed. Chunks: %d, Root: %s", len(chunks), rootHex)

	// 4. Start libp2p host
	opts := p2p.HostOptions{
		ListenPort:  *port,
		PrivKeyPath: "provider_key.key",
		EnableMDNS:  true,
		RelayAddr:   *relayAddr,
		EnableQUIC:  *enableQUIC,
	}
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), p2p.OperationTimeout)
	defer cancelStartup()
	h, err := p2p.NewHost(startupCtx, opts)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start host")
	}
	defer h.Close()

	logger.Info().Msgf("Provider Peer ID: %s", h.ID())
	for _, addr := range h.Addrs() {
		logger.Info().Msgf("Provider Address: %s/p2p/%s", addr, h.ID())
	}

	// 5. Register Handler
	h.SetStreamHandler(p2p.ProtocolID, p2p.ProviderStreamHandler(store))

	// 6. Wait for sigint
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Shutting down...")
}
