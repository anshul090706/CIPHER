package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateChunkKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("hello world, this is a test chunk")
	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text %s does not match original %s", string(decrypted), string(plaintext))
	}
}

func TestKeccak256(t *testing.T) {
	// A simple test to ensure Keccak256 works
	hash := Keccak256([]byte("test"))
	if len(hash) != 32 {
		t.Errorf("expected hash length 32, got %d", len(hash))
	}
}
