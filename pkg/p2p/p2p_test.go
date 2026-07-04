package p2p

import (
	"context"
	"testing"
	"crypto/rand"

	"github.com/1amKhush/CIPHER/pkg/chunker"
	"github.com/1amKhush/CIPHER/pkg/crypto"
	"github.com/1amKhush/CIPHER/pkg/engine"
	"github.com/libp2p/go-libp2p/core/peer"
)

func TestP2PLoopback(t *testing.T) {
	// 1. Setup provider
	providerOpts := HostOptions{ListenPort: 0, PrivKeyPath: "", EnableMDNS: false}
	providerHost, err := NewHost(context.Background(), providerOpts)
	if err != nil {
		t.Fatalf("failed to create provider host: %v", err)
	}
	defer providerHost.Close()

	// Prepare data
	dummyData := make([]byte, 100*1024)
	rand.Read(dummyData)
	chunks := chunker.ChunkBytes(dummyData)
	var leaves [][32]byte
	for _, c := range chunks {
		var fileID [32]byte
		leaf := crypto.MerkleLeaf(fileID, c.Index, uint32(len(c.Data)), c.Data)
		leaves = append(leaves, leaf)
	}
	tree := chunker.NewMerkleTree(leaves)
	
	var fileID [32]byte
	store := &engine.ChunkStore{FileID: fileID, Chunks: chunks, MerkleTree: tree}
	providerHost.SetStreamHandler(ProtocolID, ProviderStreamHandler(store))

	// 2. Setup client
	clientOpts := HostOptions{ListenPort: 0, PrivKeyPath: "", EnableMDNS: false}
	clientHost, err := NewHost(context.Background(), clientOpts)
	if err != nil {
		t.Fatalf("failed to create client host: %v", err)
	}
	defer clientHost.Close()

	providerInfo := peer.AddrInfo{
		ID:    providerHost.ID(),
		Addrs: providerHost.Addrs(),
	}

	if err := clientHost.Connect(context.Background(), providerInfo); err != nil {
		t.Fatalf("client failed to connect: %v", err)
	}

	// 3. Test loopback for all chunks
	privKey := GetHostPrivateKey(clientHost)
	var downloadedData []byte
	
	for i := uint64(0); i < uint64(len(chunks)); i++ {
		plaintext, err := RequestChunk(context.Background(), clientHost, providerHost.ID(), fileID, tree.Root, i, privKey)
		if err != nil {
			t.Fatalf("RequestChunk %d failed: %v", i, err)
		}
		
		if len(plaintext) != len(chunks[i].Data) {
			t.Fatalf("Length mismatch for chunk %d: got %d, expected %d", i, len(plaintext), len(chunks[i].Data))
		}
		downloadedData = append(downloadedData, plaintext...)
	}

	if len(downloadedData) != len(dummyData) {
		t.Fatalf("Final length mismatch: got %d, expected %d", len(downloadedData), len(dummyData))
	}
}
