package p2pclient

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multibase"
)

// ProtocolID to communicate with p2p-gateway.
const ProtocolID = "/serverista-proxy/1.0.0"

const sendRequestTimeout = 10 * time.Second

// ProxyRequest represents the payload to send to the p2p-gateway.
type ProxyRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
}

// ProxyResponse is the response from p2p-gateway.
type ProxyResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    []byte            `json:"body,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// Client hold the structures to sign messages and communicate with the p2p-gateway.
type Client struct {
	host           host.Host
	privKey        ed25519.PrivateKey
	did            string
	p2pGatewayAddr string
	addrInfo       *peer.AddrInfo
}

// New creates a new client given a libp2p host which will be used to connect and send a message to the remote protocol.
// in the params you can use any ed25519 private key to sign the messages. This should be the private key that the DID
// was derived from and entered in serverista IAM DID Key.
// proxyAddr
func New(h host.Host, privKey ed25519.PrivateKey, p2pGatewayAddr string) (*Client, error) {
	pubKey := privKey.Public().(ed25519.PublicKey)
	did, err := Ed25519PubKeyToDID(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get DID from public key: %w", err)
	}

	maddr, err := ma.NewMultiaddr(p2pGatewayAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer multiaddr: %w", err)
	}
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer addr info: %w", err)
	}

	return &Client{
		host:           h,
		privKey:        privKey,
		did:            did,
		p2pGatewayAddr: p2pGatewayAddr,
		addrInfo:       info,
	}, nil
}

// DID returns the DID for this client.
func (c *Client) DID() string {
	return c.did
}

// Ed25519PubKeyToDID gets the DID from a public key.
func Ed25519PubKeyToDID(pubKey ed25519.PublicKey) (string, error) {
	// multicodec prefix for Ed25519: 0xED01
	data := append([]byte{0xed, 0x01}, pubKey...)
	// multibase encode (base58btc, 'z' prefix)
	mbStr, err := multibase.Encode(multibase.Base58BTC, data)
	if err != nil {
		return "", err
	}
	return "did:key:" + mbStr, nil
}

// buildPayload returns the canonical payload
func buildPayload(method, path string, body []byte, nonce string, ts int64) string {
	hash := sha256.Sum256(body)
	contentHash := fmt.Sprintf("%x", hash[:])
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%d",
		strings.ToUpper(method),
		path,
		contentHash,
		nonce,
		ts,
	)
}

// createCanonicalHeader builds the DID authorization header for a request.
// It does NOT send the request yet.
func (c *Client) createCanonicalHeader(method, path string, body []byte, nonce string, ts int64) (string, []byte, error) {
	payload := buildPayload(method, path, body, nonce, ts)

	// sign payload
	sig := ed25519.Sign(c.privKey, []byte(payload))
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	// construct header
	authHeader := fmt.Sprintf(
		"DID %s;sig=%s;ts=%d;nonce=%s",
		c.did,
		sigB64,
		ts,
		nonce,
	)

	return authHeader, sig, nil
}

// RawRequest sends a raw request given the method, path, body and other args
func (c *Client) RawRequest(ctx context.Context, method, path string, body []byte, nonce string, ts int64) (*ProxyResponse, error) {
	return c.request(ctx, method, path, body, nonce, ts)
}

// request sends a raw request given all the required params.
func (c *Client) request(ctx context.Context, method, path string, body []byte, nonce string, ts int64) (*ProxyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, sendRequestTimeout)
	defer cancel()

	if err := c.host.Connect(ctx, *c.addrInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	// open a new stream
	s, err := c.host.NewStream(ctx, *&c.addrInfo.ID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer s.Close()

	// buffered writer/reader
	br := bufio.NewReader(s)
	bw := bufio.NewWriter(s)

	canonicalHeader, _, err := c.createCanonicalHeader(method, path, body, nonce, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to create canonical header: %w", err)
	}

	// prepare proxy request
	req := ProxyRequest{
		Method: method,
		Path:   path,
		Headers: map[string]string{
			"Authorization": canonicalHeader,
		},
		Body: body,
	}

	// send request
	if err := writeMessage(bw, req); err != nil {
		return nil, fmt.Errorf("failed to send request:: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush: %w", err)
	}

	// read response
	var resp ProxyResponse
	if err := readMessage(br, &resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	return &resp, nil
}

func (c *Client) Plans() {}

func writeMessage(w io.Writer, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var lenbuf [4]byte
	binary.BigEndian.PutUint32(lenbuf[:], uint32(len(b)))
	if _, err := w.Write(lenbuf[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func readMessage(r io.Reader, dst interface{}) error {
	var lenbuf [4]byte
	if _, err := io.ReadFull(r, lenbuf[:]); err != nil {
		return err
	}
	l := binary.BigEndian.Uint32(lenbuf[:])
	if l == 0 {
		return fmt.Errorf("zero-length message")
	}
	data := make([]byte, l)
	if _, err := io.ReadFull(r, data); err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
