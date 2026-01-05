# Message Logger Integration - Context Summary

## What We Accomplished

### 1. Integrated Message Logger from Old Repository
- Copied `services/messagelogger/` directory with all files:
  - `messagelogger.go` - Main logger implementation
  - `sizecomputer.go` - Computes serialization sizes for multiple formats
  - `converters.go` - FlatBuffers converters
  - `converters_capnp.go` - Cap'n Proto converters
  - `IMPLEMENTATION.md` - Documentation

### 2. Generated Required Schema Code
- **FlatBuffers**: Generated Go code from `proto/onlineboutique.fbs`
  - Output: `proto/proto/*.go` (package `proto`)
  - Changed package name from `onlineboutique` to `proto`

- **Cap'n Proto**: Generated Go code from `proto/onlineboutique.capnp`
  - Output: `proto/capnp/onlineboutique.capnp.go` (package `capnp`)
  - Changed package name from `onlineboutique` to `capnp`
  - Used command: `capnp compile -ogo -I/users/aruj/go/pkg/mod/capnproto.org/go/capnp/v3@v3.1.0-alpha.1/std proto/onlineboutique.capnp`

- **Dependencies**: Added `github.com/google/flatbuffers` to go.mod

### 3. Added Message Logger to All Services

**Server-side logging** (logs outgoing responses):
- `services/ad.go`
- `services/cart.go`
- `services/checkout.go`
- `services/currency.go`
- `services/email.go`
- `services/payment.go`
- `services/productcatalog.go`
- `services/recommendation.go`
- `services/shipping.go`

**Client-side logging** (logs outgoing requests):
- `services/util.go` in `mustConnARPC()` function
  - This handles all client connections (frontend, checkout, recommendation)

### 4. Logging Strategy
**Only log SENDS, not receives** (each message logged exactly once):
- Client side: `ProcessRequest()` logs outgoing requests, `ProcessResponse()` does nothing
- Server side: `ProcessRequest()` does nothing, `ProcessResponse()` logs outgoing responses

### 5. Kubernetes Volume Mounts
Added hostPath volume mounts to **58 Kubernetes YAML files**:
- Updated all service deployments across 7 configuration variants:
  - `kubernetes/apply/`
  - `kubernetes/apply-basic/`
  - `kubernetes/apply-proxy/`
  - `kubernetes/apply-reliable/`
  - `kubernetes/apply-reliable-cc/`
  - `kubernetes/apply-reliable-cc-fc/`
  - `kubernetes/apply-reliable-cc-fc-encryption/`

Volume mount configuration:
```yaml
volumeMounts:
- name: message-logs
  mountPath: /var/log/arpc-messages
volumes:
- name: message-logs
  hostPath:
    path: /users/aruj/online-boutique-arpc/logs
    type: DirectoryOrCreate
```

### 6. Docker Image Configuration
- Updated image name to: `appnetorg/onlineboutique-arpc-variants-logging:latest`
- Updated all Kubernetes YAML files to use the new image name
- Created script: `scripts/add_volume_mounts.sh`

## Critical Bug Found and Fixed

### Problem: Deleted File Issue
When containers started, log files were created but immediately became "deleted":
```
/var/log/arpc-messages/server-messages-20260105-055122.jsonl (deleted)
```

**Root cause**:
1. Container starts and logger creates log file during initialization
2. Kubernetes then mounts the hostPath volume to `/var/log/arpc-messages/`
3. Volume mount overwrites the directory, deleting files created before mount
4. File descriptor stays open but writes go to a "deleted" file that doesn't exist in directory

### Solution: Lazy File Initialization
Modified `services/messagelogger/messagelogger.go`:
- Changed logger structs to store `filename` and `logDir` instead of opening file immediately
- Added `ensureFileOpen()` methods to both `ClientMessageLogger` and `ServerMessageLogger`
- File is now created on **first write** (after volume is mounted) instead of during initialization

Key changes:
```go
// Old: Create file during initialization
file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
return &ClientMessageLogger{file: file}, nil

// New: Lazy initialization
return &ClientMessageLogger{
    filename: filename,
    logDir:   logDir,
}, nil

// File created on first write in ProcessRequest/ProcessResponse:
if err := l.ensureFileOpen(); err != nil {
    log.Printf("Failed to open log file: %v", err)
    return req, ctx, nil
}
```

## Current State

### What's Working
✅ Message logger code integrated and compiles
✅ FlatBuffers and Cap'n Proto schemas generated
✅ All services have logger integrated
✅ Kubernetes volume mounts configured
✅ Lazy file initialization implemented (NEEDS REBUILD)

### What Needs to Be Done

**NEXT STEP: Rebuild Docker Image**
```bash
./build_images.sh
```

This will:
1. Build new image with lazy file initialization fix
2. Push to `appnetorg/onlineboutique-arpc-variants-logging:latest`

**Then: Redeploy Services**
```bash
kubectl delete -f kubernetes/apply/
kubectl apply -f kubernetes/apply/
```

**Test the Logging**
```bash
# Send test request
curl -X POST http://10.96.88.88/cart/checkout \
  -d "email=test@example.com" \
  -d "street_address=123 Main St" \
  -d "zip_code=98101" \
  -d "city=Seattle" \
  -d "state=WA" \
  -d "country=USA" \
  -d "credit_card_number=4111111111111111" \
  -d "credit_card_expiration_month=12" \
  -d "credit_card_expiration_year=2030" \
  -d "credit_card_cvv=123" \
  -d "user_id=test123"

# Check logs
ls -la /users/aruj/online-boutique-arpc/logs/
cat /users/aruj/online-boutique-arpc/logs/client-messages-*.jsonl | head -5
cat /users/aruj/online-boutique-arpc/logs/server-messages-*.jsonl | head -5
```

## Log Format

Each log entry is JSON with:
```json
{
  "timestamp": "2026-01-05T05:47:51Z",
  "direction": "request",  // or "response"
  "method": "GetCart",
  "message_type": "GetCartRequest",
  "sizes": {
    "protobuf": 37,
    "flatbuffers": 112,
    "capnproto": 120,
    "protojson": 105
  },
  "payload": { /* actual message data */ }
}
```

## File Locations

- **Log output**: `/users/aruj/online-boutique-arpc/logs/` (on host machine)
- **Message logger code**: `services/messagelogger/`
- **Generated schemas**:
  - FlatBuffers: `proto/proto/`
  - Cap'n Proto: `proto/capnp/`
- **Volume mount script**: `scripts/add_volume_mounts.sh`
- **Build script**: `build_images.sh`

## Important Notes

1. **Each message is logged exactly once** when sent (not when received)
2. **Volume mount is working** - verified by creating test.txt in container and seeing it appear on host
3. **All 10 services must be deployed** for checkout to work:
   - ad, cart, checkout, currency, email, frontend, payment, productcatalog, recommendation, shipping
4. **Lazy file initialization is critical** - files must be created AFTER Kubernetes mounts the volume

## Verification Commands

```bash
# Check pods are running
kubectl get pods

# Check if logger initialized
kubectl logs deployment/cart | grep "message logging"

# Check files inside container
kubectl exec $(kubectl get pod -l app=cart -o jsonpath='{.items[0].metadata.name}') -- ls -la /var/log/arpc-messages/

# Check files on host
ls -la /users/aruj/online-boutique-arpc/logs/

# View log content
cat /users/aruj/online-boutique-arpc/logs/*.jsonl
```
