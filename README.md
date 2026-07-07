# 🌐 CIPHER

> **CDNs are owned by a handful of companies. CIPHER is the alternative — a decentralized content delivery protocol where math replaces the middleman.**

CIPHER is a peer-to-peer content delivery network that enables secure and verifiable file transfer between peers. It combines libp2p networking, encrypted chunk delivery, signed lottery tickets, and Merkle-based integrity verification.

Whether peers are behind different NATs or isolated networks, CIPHER uses Circuit Relay v2 to establish connectivity and securely deliver verifiable content.

## Highlights

* **Decentralized by Design** — Content is transferred directly between peers without relying on a centralized CDN.
* **Cross-Network Connectivity** — libp2p Circuit Relay v2 enables communication across isolated networks and NATs.
* **Cryptographic Integrity** — Keccak256-based Merkle proofs verify every received chunk.
* **Secure Transport** — Chunks are encrypted using XChaCha20-Poly1305 authenticated encryption.
* **Multi-Protocol Networking** — Built on libp2p with TCP, QUIC, and Secure WebSocket support.

---

## How CIPHER Works

```text
Provider (Network A)
        │
        │ Reserve Relay Slot
        ▼
Public Circuit Relay v2
        ▲
        │ /p2p-circuit/
        │
Client (Network B)
```

For every requested chunk:

```text
ChunkRequest
      ↓
ChunkResponse
      ↓
LotteryTicket
      ↓
KeyReveal
      ↓
Decrypt Chunk
      ↓
Verify HResp
      ↓
Verify Merkle Proof
      ↓
Write Verified Data
```

---

# Usage

CIPHER uses three components:

* **Relay** — provides public cross-network reachability.
* **Provider** — serves encrypted and verifiable file chunks.
* **Client** — downloads, decrypts, and verifies the chunks.

## 1. Deploy the Relay

The relay executable is located at:

```text
cmd/relay/main.go
```

For a Render Web Service, configure:

```text
Language: Go
Branch: main
Root Directory: Leave blank

Build Command:
go build -o cipher-relay ./cmd/relay

Start Command:
./cipher-relay
```

After deployment, check the service logs:

```text
CIPHER RELAY STARTED
Relay Peer ID: <RELAY_PEER_ID>
Port: 10000
Circuit Relay v2 service: ACTIVE
```

Save:

```text
RELAY_HOST
RELAY_PEER_ID
```

The relay multiaddress is:

```text
/dns4/<RELAY_HOST>/tcp/443/wss/p2p/<RELAY_PEER_ID>
```

> **Note:** The current relay identity may change after a restart or redeployment. Always use the latest Relay Peer ID from the deployment logs.

---

## 2. Start the Provider

On **Network A**:

```bash
go run ./cmd/provider \
  -relay /dns4/<RELAY_HOST>/tcp/443/wss/p2p/<RELAY_PEER_ID> \
  --verbose
```

Successful startup should show:

```text
Connected to relay <RELAY_PEER_ID>
Successfully reserved slot on relay
Provider Peer ID: <PROVIDER_PEER_ID>
```

Also note:

```text
Root: <MERKLE_ROOT>
Chunks: <CHUNK_COUNT>
```

Keep the provider running.

---

## 3. Start the Client

On **Network B**, clone the repository if required:

```bash
git clone https://github.com/anshul090706/CIPHER.git
cd CIPHER
go mod download
```

Run:

```bash
go run ./cmd/client \
  -provider /dns4/<RELAY_HOST>/tcp/443/wss/p2p/<RELAY_PEER_ID>/p2p-circuit/p2p/<PROVIDER_PEER_ID> \
  -root <MERKLE_ROOT> \
  -chunks <CHUNK_COUNT> \
  --verbose
```

The client will:

```text
Connect through Relay
        ↓
Request Encrypted Chunks
        ↓
Receive Key Reveal
        ↓
Decrypt Chunks
        ↓
Verify HResp
        ↓
Verify Merkle Proofs
        ↓
Create downloaded_file.txt
```

The client does **not** need the provider's original file.

---

## Verify the Transfer

On Windows PowerShell:

**Provider**

```powershell
Get-FileHash test_file.txt -Algorithm SHA256
```

**Client**

```powershell
Get-FileHash downloaded_file.txt -Algorithm SHA256
```

Both hashes should match.

---

# Bugs Fixed

During cross-network integration testing, the following networking issues were identified and resolved.

### 1. Broken Public Relay

The previously documented relay failed during the Secure WebSocket handshake:

```text
websocket: bad handshake
```

A dedicated deployable Circuit Relay v2 service was added under:

```text
cmd/relay
```

### 2. Relay Reservation Failure

The provider connected to the relay but failed to reserve a slot because the HOP protocol was unavailable:

```text
protocols not supported:
[/libp2p/circuit/relay/0.2.0/hop]
```

The Circuit Relay v2 service is now explicitly initialized on the relay host.

### 3. mDNS Interference

When provider and client were on the same LAN, mDNS could establish an unintended direct connection instead of using the explicit relay circuit.

mDNS was disabled in provider and client configuration for deterministic relay-based connectivity.

### 4. Stream Timeout over Relay

The client connected to the provider through the relay, but opening the CIPHER application stream failed with:

```text
context deadline exceeded
```

The client stream context now explicitly allows the CIPHER chunk protocol to use libp2p limited relay connections.

After these fixes, the tested flow is:

```text
Provider
   ↓
Connect to Relay
   ↓
Reserve Relay Slot
   ↓
Client Connects via /p2p-circuit/
   ↓
CIPHER Stream Opens
   ↓
Encrypted Chunk Transfer
   ↓
HResp Verification
   ↓
Merkle Proof Verification
   ↓
downloaded_file.txt
```

---

## Diagnostic Logging

Use `--verbose` for detailed transport logs.

**Provider**

```bash
go run ./cmd/provider -relay <RELAY_ADDRESS> --verbose
```

**Client**

```bash
go run ./cmd/client \
  -provider <PROVIDER_CIRCUIT_ADDRESS> \
  -root <MERKLE_ROOT> \
  -chunks <CHUNK_COUNT> \
  --verbose
```
