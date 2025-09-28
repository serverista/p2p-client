package p2pclient

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/multiformats/go-multibase"
)

func TestEd25519PubKeyToDID(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	_ = priv // only need pub for this test

	did, err := Ed25519PubKeyToDID(pub)
	if err != nil {
		t.Fatalf("Ed25519PubKeyToDID error: %v", err)
	}
	if !strings.HasPrefix(did, "did:key:") {
		t.Fatalf("did must start with did:key:; got %q", did)
	}
	mbStr := strings.TrimPrefix(did, "did:key:")
	enc, data, err := multibase.Decode(mbStr)
	if err != nil {
		t.Fatalf("multibase decode: %v", err)
	}
	if enc != multibase.Base58BTC {
		t.Fatalf("expected base58btc multibase, got %v", enc)
	}
	if len(data) < 2 {
		t.Fatalf("decoded data too short")
	}
	if data[0] != 0xed || data[1] != 0x01 {
		t.Fatalf("multicodec prefix mismatch: %x %x", data[0], data[1])
	}
	if !bytes.Equal(data[2:], pub) {
		t.Fatalf("decoded public key not equal to original")
	}
}

func TestCreateCanonicalHeaderAndSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	did, err := Ed25519PubKeyToDID(pub)
	if err != nil {
		t.Fatalf("Ed25519PubKeyToDID error: %v", err)
	}

	c := &Client{privKey: priv, did: did}
	method := "POST"
	path := "/test/endpoint"
	body := []byte("payload")
	nonce := "n-1"
	ts := time.Now().Unix()

	header, sig, err := c.createCanonicalHeader(method, path, body, nonce, ts)
	if err != nil {
		t.Fatalf("createCanonicalHeader error: %v", err)
	}
	if !strings.Contains(header, did) {
		t.Fatalf("header missing did")
	}
	if !strings.Contains(header, nonce) {
		t.Fatalf("header missing nonce")
	}
	if !strings.Contains(header, fmt.Sprintf("ts=%d", ts)) {
		t.Fatalf("header missing ts")
	}

	// rebuild payload and verify signature
	payload := buildPayload(method, path, body, nonce, ts)
	if !ed25519.Verify(pub, []byte(payload), sig) {
		t.Fatalf("signature did not verify")
	}
}
