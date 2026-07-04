package main

import (
	"encoding/hex"
	"fmt"
	"github.com/1amKhush/CIPHER/pkg/chunker"
	"github.com/1amKhush/CIPHER/pkg/crypto"
	"github.com/1amKhush/CIPHER/pkg/p2p"
	"context"
)

func main() {
	opts := p2p.HostOptions{ListenPort: 9020, PrivKeyPath: "provider_key.key", EnableMDNS: false}
	h, _ := p2p.NewHost(context.Background(), opts)
	defer h.Close()
	fmt.Printf("PROVIDER_ID=%s\n", h.ID().String())
	
	chunks, _ := chunker.ChunkFile("test_file.txt")
	var leaves [][32]byte
	for _, c := range chunks {
		var fileID [32]byte
		leaves = append(leaves, crypto.MerkleLeaf(fileID, c.Index, uint32(len(c.Data)), c.Data))
	}
	tree := chunker.NewMerkleTree(leaves)
	fmt.Printf("ROOT=%s\n", hex.EncodeToString(tree.Root[:]))
}
