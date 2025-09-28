# P2P Client

A lightweight Go client for communicating with a **p2p-gateway** over [libp2p](https://libp2p.io/). This package handles:

- Generating and managing [DID](https://www.w3.org/TR/did-core/) identities from Ed25519 keys.
- Constructing and signing canonical request payloads.
- Sending `ProxyRequest` messages to a remote peer using libp2p streams.
- Receiving and decoding `ProxyResponse` messages.

## Canonical Header Protocol

All requests must include an `Authorization` header that cryptographically proves the request’s origin and prevents tampering.

### 1. Canonical Payload

A deterministic payload string is constructed as follows:

- **METHOD**: The HTTP-like method, uppercased (e.g., `GET`, `POST`).  
- **PATH**: The request path (e.g., `/v1/resource`).  
- **SHA256(BODY)**: A lowercase hex-encoded SHA-256 digest of the request body.  
- **NONCE**: A unique client-generated string to prevent replay attacks.  
- **TIMESTAMP**: Unix timestamp (in seconds) when the request was created.  

### 2. Signature

- The canonical payload is signed using the client’s **Ed25519 private key**.  
- The resulting signature is encoded in **Base64**.  

### 3. Authorization Header

The final header is constructed as:

Authorization: DID did:key:...;sig=<base64_signature>;ts=<timestamp>;nonce=<nonce>

### 4. Verification Process

On the **p2p-gateway**, the request is verified by:

1. Reconstructing the canonical payload.  
2. Extracting the DID from the header and resolving the public key.  
3. Verifying the Ed25519 signature.  
4. Validating the timestamp is within acceptable skew.  
5. Checking that the nonce has not been reused.  

If any step fails, the request is rejected.

## Security Properties

- **Integrity**: The signed payload ensures no tampering with method, path, body, or metadata.  
- **Authentication**: Only the holder of the DID private key can generate valid signatures.  
- **Replay Protection**: Timestamps and nonces ensure requests cannot be reused.  


---

## Features

- **Canonical Request Signing**: Builds deterministic payloads for signature generation.  
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
resp, err := client.RawRequest(ctx, "GET", "/v1/services", nil, "random-nonce", time.Now().Unix())
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