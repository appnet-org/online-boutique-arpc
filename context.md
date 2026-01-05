# Message Logger & Serialization Size Benchmarking - Complete Context

## Overview

This document provides complete context for the message logger implementation that logs all RPC messages and computes serialization sizes across multiple formats: Protobuf, FlatBuffers, Cap'n Proto, and Symphony.

### Purpose
- Log every RPC message sent through the system (requests and responses)
- Compute and compare serialization sizes across 4 different formats
- Output structured JSON logs with message payloads and size metadata

### Serialization Formats Tracked
1. **Protobuf** - Used for actual RPC communication
2. **FlatBuffers** - Zero-copy serialization format
3. **Cap'n Proto** - Another zero-copy format
4. **Symphony** - Custom serialization format (generated methods on protobuf types)

---

## What We Accomplished

### 1. Integrated Message Logger from Old Repository

Copied `services/messagelogger/` directory with all files:
- `messagelogger.go` - Main logger implementation with lazy file initialization
- `sizecomputer.go` - Computes serialization sizes for all 4 formats
- `converters.go` - FlatBuffers converters for 33 message types
- `converters_capnp.go` - Cap'n Proto converters for 33 message types

### 2. Generated Required Schema Code

**FlatBuffers:**
- Generated Go code from `proto/onlineboutique.fbs`
- Output: `proto/proto/*.go` (package `proto`)
- Changed package name from `onlineboutique` to `proto` to avoid conflicts
- Command: `flatc --go proto/onlineboutique.fbs`

**Cap'n Proto:**
- Generated Go code from `proto/onlineboutique.capnp`
- Output: `proto/capnp/onlineboutique.capnp.go` (package `capnp`)
- Changed package name from `onlineboutique` to `capnp` to avoid conflicts
- Command: `capnp compile -ogo -I/users/aruj/go/pkg/mod/capnproto.org/go/capnp/v3@v3.1.0-alpha.1/std proto/onlineboutique.capnp`

**Symphony:**
- No separate generation needed - `MarshalSymphony()` methods are generated directly on protobuf types in `proto/onlineboutique.syn.go`
- Already available in the codebase

**Dependencies Added:**
- `github.com/google/flatbuffers v25.12.19+incompatible`
- `capnproto.org/go/capnp/v3` (already present)
- `google.golang.org/protobuf` (already present)

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
- **Client side**: `ProcessRequest()` logs outgoing requests, `ProcessResponse()` does nothing
- **Server side**: `ProcessRequest()` does nothing, `ProcessResponse()` logs outgoing responses

This ensures each message is logged exactly once in the system.

### 5. Kubernetes Volume Mounts

Added hostPath volume mounts to **58 Kubernetes YAML files** across 7 configuration variants:
- `kubernetes/apply/`
- `kubernetes/apply-basic/`
- `kubernetes/apply-proxy/`
- `kubernetes/apply-reliable/`
- `kubernetes/apply-reliable-cc/`
- `kubernetes/apply-reliable-cc-fc/`
- `kubernetes/apply-reliable-cc-fc-encryption/`

**Volume mount configuration:**
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

- Image name: `appnetorg/onlineboutique-arpc-variants-logging:latest`
- All Kubernetes YAML files updated to use this image
- Build script: `build_images.sh`
- Volume mount script: `scripts/add_volume_mounts.sh`

---

## Critical Bug Found and Fixed

### Problem: Deleted File Issue

When containers started, log files were created but immediately became "deleted":
```
/var/log/arpc-messages/server-messages-20260105-055122.jsonl (deleted)
```

**Root cause:**
1. Container starts and logger creates log file during initialization
2. Kubernetes then mounts the hostPath volume to `/var/log/arpc-messages/`
3. Volume mount overwrites the directory, deleting files created before mount
4. File descriptor stays open but writes go to a "deleted" file that doesn't exist in directory

### Solution: Lazy File Initialization

Modified `services/messagelogger/messagelogger.go`:
- Changed logger structs to store `filename` and `logDir` instead of opening file immediately
- Added `ensureFileOpen()` methods to both `ClientMessageLogger` and `ServerMessageLogger`
- File is now created on **first write** (after volume is mounted) instead of during initialization

**Key changes:**
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

---

## Technical Implementation Details

### File Structure

```
services/messagelogger/
├── messagelogger.go          # Main logger with lazy file initialization
├── converters.go             # FlatBuffers converters + some Cap'n Proto
├── converters_capnp.go       # Remaining Cap'n Proto converters
└── sizecomputer.go           # Size computation logic for all formats

proto/
├── onlineboutique.proto      # Protobuf schema (original)
├── onlineboutique.fbs        # FlatBuffers schema
├── onlineboutique.capnp      # Cap'n Proto schema
├── onlineboutique.syn.go     # Symphony generated code (MarshalSymphony methods)
├── proto/                    # FlatBuffers generated code
│   ├── Money.go
│   ├── CartItem.go
│   └── ... (all message types)
└── capnp/                    # Cap'n Proto generated code
    └── onlineboutique.capnp.go
```

### Key Architectural Decisions

**1. Type Conflict Resolution:**
- Both Cap'n Proto and Protobuf generate types with same names (e.g., `CartItem`, `Money`)
- **Solution:** Moved Cap'n Proto generated code to `proto/capnp/` package
- Changed Cap'n Proto package name from `onlineboutique` to `capnp`
- FlatBuffers also in separate package: `proto` at `proto/proto/`

**2. Symphony Integration:**
- Unlike FlatBuffers/Cap'n Proto, Symphony generates methods **directly on protobuf types**
- No separate converter functions needed - just call `msg.MarshalSymphony()`
- Much simpler than the other formats

**Import aliases used in code:**
- `pb` for protobuf (`github.com/appnetorg/online-boutique-arpc/proto`)
- `fb` for FlatBuffers (`github.com/appnetorg/online-boutique-arpc/proto/proto`)
- `pbcapnp` for Cap'n Proto (`github.com/appnetorg/online-boutique-arpc/proto/capnp`)

### Size Computation (`sizecomputer.go`)

**Main function:**
```go
func ComputeSizes(msg interface{}) SerializationSizes
```

**Data structure:**
```go
type SerializationSizes struct {
    Protobuf    int               `json:"protobuf"`
    FlatBuffers int               `json:"flatbuffers"`
    CapnProto   int               `json:"capnproto"`
    Symphony    int               `json:"symphony"`
    Errors      map[string]string `json:"errors,omitempty"`
}
```

**How it works:**
1. **Protobuf**: Uses `proto.Marshal(msg)` to get wire format bytes
2. **FlatBuffers**: Type switch to call appropriate `ProtoToFB_MessageType()` converter
3. **Cap'n Proto**: Type switch to call appropriate `ProtoToCapnp_MessageType()` converter
4. **Symphony**: Interface check for `MarshalSymphony()` method and calls it
5. Computes `len(bytes)` for each format
6. Returns -1 for failed conversions
7. Stores error messages in `Errors` map

### Converter Functions

**FlatBuffers converters** (in `converters.go`):
```go
func ProtoToFB_Empty(pb *pb.Empty) ([]byte, error)
func ProtoToFB_Money(pb *pb.Money) ([]byte, error)
func ProtoToFB_AddItemRequest(pb *pb.AddItemRequest) ([]byte, error)
// ... 33 total message types
```

**Cap'n Proto converters** (in `converters.go` and `converters_capnp.go`):
```go
func ProtoToCapnp_Empty(pbMsg *pb.Empty) ([]byte, error)
func ProtoToCapnp_Money(pbMsg *pb.Money) ([]byte, error)
func ProtoToCapnp_AddItemRequest(pbMsg *pb.AddItemRequest) ([]byte, error)
// ... 33 total message types
```

**Symphony** (no converters needed):
```go
// Just call the method directly on any protobuf message:
data, err := msg.MarshalSymphony()
```

**Converter characteristics:**
- Take protobuf message as input
- Return marshaled bytes + error
- Handle nested structures (Money, Address, CartItem, etc.)
- Handle lists/arrays

### Log Entry Structure

```go
type LogEntry struct {
    Timestamp   string             `json:"timestamp"`
    Direction   string             `json:"direction"` // "request" or "response"
    Method      string             `json:"method,omitempty"`
    MessageType string             `json:"message_type"`
    Sizes       SerializationSizes `json:"sizes"`
    Payload     interface{}        `json:"payload"`
}
```

**Example log output:**
```json
{
  "timestamp": "2026-01-05T06:00:35Z",
  "direction": "request",
  "method": "GetCart",
  "message_type": "GetCartRequest",
  "sizes": {
    "protobuf": 37,
    "flatbuffers": 112,
    "capnproto": 120,
    "symphony": 50
  },
  "payload": {
    "user_id": "test123"
  }
}
```

### Message Logger Implementation (`messagelogger.go`)

**Logging flow:**

**Client-side (logs outgoing requests):**
```go
func (l *ClientMessageLogger) ProcessRequest(ctx context.Context, req *element.RPCRequest) {
    l.mu.Lock()
    defer l.mu.Unlock()

    // Lazy-create the log file on first write
    if err := l.ensureFileOpen(); err != nil {
        return req, ctx, nil // Don't fail the RPC
    }

    // Compute sizes for all formats
    sizes := ComputeSizes(req.Payload)

    // Create and write log entry
    entry := LogEntry{
        Timestamp:   time.Now().UTC().Format(time.RFC3339),
        Direction:   "request",
        Method:      req.Method,
        MessageType: GetMessageTypeName(req.Payload),
        Sizes:       sizes,
        Payload:     req.Payload,
    }

    data, _ := MarshalLogEntry(entry)
    l.file.Write(append(data, '\n'))

    return req, ctx, nil
}

func (l *ClientMessageLogger) ProcessResponse(ctx context.Context, resp *element.RPCResponse) {
    // Don't log receives - only log sends
    return resp, ctx, nil
}
```

**Server-side (logs outgoing responses):**
```go
func (l *ServerMessageLogger) ProcessRequest(ctx context.Context, req *element.RPCRequest) {
    // Don't log receives - only log sends
    return req, ctx, nil
}

func (l *ServerMessageLogger) ProcessResponse(ctx context.Context, resp *element.RPCResponse) {
    l.mu.Lock()
    defer l.mu.Unlock()

    // Lazy-create the log file on first write
    if err := l.ensureFileOpen(); err != nil {
        return resp, ctx, nil // Don't fail the RPC
    }

    // Compute sizes and log (same as client)
    sizes := ComputeSizes(resp.Result)
    // ... create and write log entry

    return resp, ctx, nil
}
```

---

## All Supported Message Types (33 total)

1. Empty
2. EmptyUser
3. Money
4. Address
5. CartItem
6. CreditCardInfo
7. Ad
8. AddItemRequest
9. GetCartRequest
10. EmptyCartRequest
11. Cart
12. ListRecommendationsRequest
13. ListRecommendationsResponse
14. Product
15. GetProductRequest
16. SearchProductsRequest
17. ListProductsResponse
18. SearchProductsResponse
19. GetQuoteRequest
20. GetQuoteResponse
21. ShipOrderRequest
22. ShipOrderResponse
23. GetSupportedCurrenciesResponse
24. CurrencyConversionRequest
25. ChargeRequest
26. ChargeResponse
27. OrderItem
28. OrderResult
29. SendOrderConfirmationRequest
30. PlaceOrderRequest
31. PlaceOrderResponse
32. AdRequest
33. AdResponse

All 33 message types have complete converters for FlatBuffers and Cap'n Proto. Symphony works automatically via generated methods.

---

## File Locations

- **Log output**: `/users/aruj/online-boutique-arpc/logs/` (on host machine)
- **Container log path**: `/var/log/arpc-messages/`
- **Message logger code**: `services/messagelogger/`
- **Generated schemas**:
  - Protobuf: `proto/onlineboutique.pb.go`
  - FlatBuffers: `proto/proto/`
  - Cap'n Proto: `proto/capnp/`
  - Symphony: `proto/onlineboutique.syn.go`
- **Volume mount script**: `scripts/add_volume_mounts.sh`
- **Build script**: `build_images.sh`
- **Context doc**: `context.md` (this file)

---

## How to Extend

### Adding a New Message Type

If you add a new protobuf message type to the schema:

**1. Update protobuf schema:**
```bash
# Add message to proto/onlineboutique.proto
# Regenerate protobuf code
protoc --go_out=. --go_opt=paths=source_relative proto/onlineboutique.proto
```

**2. Generate FlatBuffers code:**
```bash
# Add message to proto/onlineboutique.fbs
flatc --go proto/onlineboutique.fbs
```

**3. Generate Cap'n Proto code:**
```bash
# Add message to proto/onlineboutique.capnp
capnp compile -ogo -I/users/aruj/go/pkg/mod/capnproto.org/go/capnp/v3@v3.1.0-alpha.1/std proto/onlineboutique.capnp
# Move generated file
mv proto/onlineboutique.capnp.go proto/capnp/
# Fix package name if needed
sed -i 's/^package proto$/package capnp/' proto/capnp/onlineboutique.capnp.go
```

**4. Add FlatBuffers converter in `converters.go`:**
```go
func ProtoToFB_NewMessageType(pbMsg *pb.NewMessageType) ([]byte, error) {
    builder := flatbuffers.NewBuilder(256)

    // Build nested structures first
    // Then build main structure

    fb.NewMessageTypeStart(builder)
    // Add fields...
    obj := fb.NewMessageTypeEnd(builder)

    builder.Finish(obj)
    return builder.FinishedBytes(), nil
}
```

**5. Add Cap'n Proto converter in `converters_capnp.go`:**
```go
func ProtoToCapnp_NewMessageType(pbMsg *pb.NewMessageType) ([]byte, error) {
    msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
    if err != nil {
        return nil, fmt.Errorf("failed to create message: %w", err)
    }

    newMsg, err := pbcapnp.NewRootNewMessageType(seg)
    if err != nil {
        return nil, fmt.Errorf("failed to create NewMessageType: %w", err)
    }

    // Set fields...

    data, err := msg.Marshal()
    if err != nil {
        return nil, fmt.Errorf("failed to marshal message: %w", err)
    }

    return data, nil
}
```

**6. Add type cases in `sizecomputer.go`:**

In `convertToFlatBuffers()`:
```go
case *pb.NewMessageType:
    return ProtoToFB_NewMessageType(v)
```

In `convertToCapnProto()`:
```go
case *pb.NewMessageType:
    return ProtoToCapnp_NewMessageType(v)
```

**7. Symphony requires no changes** - if the Symphony code was regenerated, the `MarshalSymphony()` method will already be available.

---

## Important Implementation Notes

### Error Handling
- **All errors are caught and logged** - RPC never fails due to logging errors
- Failed conversions result in size = -1
- Error details stored in `Sizes.Errors` map
- Logging is best-effort and non-blocking

### Performance Considerations
- Size computation happens on **every** RPC request/response
- Marshaling happens 4 times per message (protobuf, FlatBuffers, Cap'n Proto, Symphony)
- This is synchronous and will add latency to RPCs
- All marshaling is done inline during RPC processing
- Consider making this optional via environment variable if performance becomes critical

### FlatBuffers Specifics
- Builder pattern: build nested structures first, then parent
- Vectors must be built in reverse order
- Strings must be created before using them in structures
- `builder.Finish(obj)` must be called before `FinishedBytes()`
- Fixed-length fields can be set directly in the table

### Cap'n Proto Specifics
- Always create message and segment first: `capnp.NewMessage(capnp.SingleSegment(nil))`
- Use `NewRoot*` for top-level types
- Use `New*` for nested types
- Lists use `New*_List(seg, int32(len(items)))`
- Text lists use `capnp.NewTextList(seg, int32(len(items)))`
- Must marshal the entire message at the end

### Symphony Specifics
- Methods are generated directly on protobuf types
- No separate converter functions needed
- Just call `msg.MarshalSymphony()` via interface
- Simpler than FlatBuffers/Cap'n Proto integration

### Lazy File Initialization
- **Critical for Kubernetes deployments**
- Files must be created AFTER Kubernetes mounts the volume
- `ensureFileOpen()` is called on first log write
- Subsequent writes reuse the open file handle

---

## Deployment & Testing

### Build and Deploy

**Build Docker image:**
```bash
./build_images.sh
```

**Deploy to Kubernetes:**
```bash
kubectl delete -f kubernetes/apply/
kubectl apply -f kubernetes/apply/
```

**Check deployment:**
```bash
# Check pods are running
kubectl get pods

# Check if logger initialized
kubectl logs deployment/cart | grep "message logging"

# Check files inside container
kubectl exec $(kubectl get pod -l app=cart -o jsonpath='{.items[0].metadata.name}') -- ls -la /var/log/arpc-messages/
```

### Testing the Logging

**Send test request:**
```bash
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
```

**Check logs on host:**
```bash
# List log files
ls -la /users/aruj/online-boutique-arpc/logs/

# View client logs (outgoing requests)
cat /users/aruj/online-boutique-arpc/logs/client-messages-*.jsonl | head -5

# View server logs (outgoing responses)
cat /users/aruj/online-boutique-arpc/logs/server-messages-*.jsonl | head -5

# Pretty print with jq
cat /users/aruj/online-boutique-arpc/logs/client-messages-*.jsonl | jq .
```

**Log file naming:**
- Client logs: `client-messages-YYYYMMDD-HHMMSS.jsonl`
- Server logs: `server-messages-YYYYMMDD-HHMMSS.jsonl`
- Each service instance creates its own log file with timestamp
- JSONL format: one JSON object per line

---

## Troubleshooting

### Build Error: "redeclared in this block"
- **Cause:** Cap'n Proto and Protobuf types in same package
- **Fix:** Ensure Cap'n Proto code is in `proto/capnp/` with package `capnp`

### Build Error: "undefined: pb.NewCartItem"
- **Cause:** Missing `pbcapnp` import alias or wrong package reference
- **Fix:** Use `pbcapnp.NewCartItem` for Cap'n Proto types, `pb.CartItem` for protobuf

### Size always -1 for a format
- **Cause:** Converter function failing
- **Fix:** Check error in `Sizes.Errors` field in log output
- Common issues: nil pointers, missing nested structure initialization

### FlatBuffers "vector offset is before current position"
- **Cause:** Building structures in wrong order
- **Fix:** Build all nested structures and vectors BEFORE starting parent structure

### Log files show as "(deleted)" in container
- **Cause:** File created before volume mount
- **Fix:** Ensure lazy file initialization is implemented (should already be fixed)

### Empty payload in logs
- **Cause:** Message field is empty string or default value
- **Example:** `GetCartRequest{UserId: ""}` shows as `{}` due to `omitempty` JSON tag
- **Fix:** This is expected behavior - protobuf omits default values in JSON

### No log files appearing
- **Check:** Verify volume mount is correct in YAML files
- **Check:** Verify pod has write permissions to `/var/log/arpc-messages/`
- **Check:** Look for errors in pod logs: `kubectl logs <pod-name> | grep -i log`

### Volume mount not working
- **Verify:** Check hostPath exists: `ls -la /users/aruj/online-boutique-arpc/logs/`
- **Verify:** Create test file from container: `kubectl exec <pod> -- touch /var/log/arpc-messages/test.txt`
- **Verify:** Check if test file appears on host

---

## Current State & Known Issues

### What's Working ✅
- Message logger code integrated and compiles
- FlatBuffers, Cap'n Proto, and Symphony size computation
- All services have logger integrated (client and server side)
- Kubernetes volume mounts configured across all deployment variants
- Lazy file initialization implemented and deployed
- Log files being created and populated correctly

### Known Issues 🐛
- **Frontend bug**: `placeOrderHandler` uses `sessionID(r)` instead of `userId` variable from form (line 290 in `services/frontend.go`)
  - This causes GetCartRequest to have empty user_id even when curl provides it
  - Frontend logs show: `user_id: test` but PlaceOrder gets `user_id: ""`
  - Should change `UserId: sessionID(r)` to `UserId: userId`

### Services Deployed
All 10 services must be running for checkout to work:
1. frontend
2. cart
3. checkout
4. currency
5. email
6. payment
7. productcatalog
8. recommendation
9. shipping
10. ad

---

## Quick Reference

**Important paths:**
- Host log directory: `/users/aruj/online-boutique-arpc/logs/`
- Container log directory: `/var/log/arpc-messages/`
- Message logger: `services/messagelogger/`
- Build script: `./build_images.sh`
- Kubernetes configs: `kubernetes/apply/`

**Key commands:**
```bash
# Build and deploy
./build_images.sh
kubectl delete -f kubernetes/apply/ && kubectl apply -f kubernetes/apply/

# View logs
cat logs/*.jsonl | jq .

# Check pods
kubectl get pods
kubectl logs deployment/cart | grep -i message
```

**Important notes:**
- Each message logged exactly once (on send, not receive)
- Lazy file initialization prevents deleted file issue
- Symphony is simplest - no converters needed
- FlatBuffers/Cap'n Proto need manual converters for each type
- All errors are caught - logging never fails RPC calls
