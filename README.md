# P2P Client

A lightweight Go client for communicating with a **p2p-gateway** over [libp2p](https://libp2p.io/). This package handles:

- Generating and managing [DID](https://www.w3.org/TR/did-core/) identities from Ed25519 keys.
- Constructing and signing canonical request payloads.
- Sending `ProxyRequest` messages to a remote peer using libp2p streams.
- Receiving and decoding `ProxyResponse` messages.

---

## Features

- **DID Support**: Automatically derives a DID from an Ed25519 public key.
- **Signed Requests**: Canonical request signing using Ed25519 signatures.
- **Proxy Communication**: Serialize, send, and receive requests/responses over libp2p streams.
- **Timeout Handling**: Built-in request timeout for reliability.

---

## Installation

```bash
go get github.com/serverista/p2p-client
```

---

## Usage

### 1. Create a client

```go
import (
    "crypto/ed25519"
    "github.com/libp2p/go-libp2p"
    "github.com/serverista/p2p-client"
)

func main() {
    // Generate a new Ed25519 key pair
    pub, priv, _ := ed25519.GenerateKey(nil)

    // Start a libp2p host
    h, _ := libp2p.New()

    // Proxy peer multiaddress (example)
    proxyAddr := "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooW..."

    // Create client
    client, err := p2pclient.New(h, priv, proxyAddr)
    if err != nil {
        panic(err)
    }

    println("Client DID:", client.DID())
}
```

### 2. Send a request

```go
ctx := context.Background()
resp, err := client.Request(ctx, "GET", "/v1/services", nil, "random-nonce", time.Now().Unix())
if err != nil {
    panic(err)
}

fmt.Printf("Response status: %d\n", resp.Status)
fmt.Printf("Response body: %s\n", string(resp.Body))
```

---

## Data Structures

### ProxyRequest
```go
type ProxyRequest struct {
    Method  string            `json:"method"`
    Path    string            `json:"path"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    []byte            `json:"body,omitempty"`
}
```

### ProxyResponse

```go
type ProxyResponse struct {
    Status  int               `json:"status"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    []byte            `json:"body,omitempty"`
    Error   string            `json:"error,omitempty"`
}
```