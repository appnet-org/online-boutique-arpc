# Message Logger Serialization Size Benchmarking Implementation

## Overview

This implementation extends the existing message logger to compute and log serialization sizes for multiple formats:
- **Protobuf** (used for RPC)
- **FlatBuffers**
- **Cap'n Proto**
- **Protobuf JSON** (protojson)

Every time a message is sent or received, the logger converts it to all four formats, computes the byte size of each, and includes these sizes in the log output.

---

## File Structure

### New Files Created

```
services/messagelogger/
├── messagelogger.go          # Main logger (MODIFIED)
├── converters.go             # FlatBuffers converters + Cap'n Proto entry points
├── converters_capnp.go       # Remaining Cap'n Proto converters
├── sizecomputer.go           # Size computation logic
└── IMPLEMENTATION.md         # This file

proto/
├── onlineboutique.capnp      # Cap'n Proto schema (MODIFIED - package name)
└── capnp/
    └── onlineboutique.capnp.go  # Generated Cap'n Proto code (MOVED here)

proto/proto/                  # FlatBuffers generated code
├── Money.go
├── CartItem.go
├── AddItemRequest.go
└── ... (all other message types)
```

### Key Architectural Decision

**Cap'n Proto and Protobuf type conflict resolution:**
- Both Cap'n Proto and Protobuf generate types with the same names (e.g., `CartItem`, `Money`)
- **Solution:** Moved Cap'n Proto generated code to `proto/capnp/` package
- Changed Cap'n Proto package name from `onlineboutique` to `capnp`
- This allows both to coexist without redeclaration errors

---

## Implementation Details

### 1. FlatBuffers Code Generation

```bash
flatc --go proto/onlineboutique.fbs
```

This generates Go code in `proto/proto/` directory with package name `proto`.

**Import alias in code:** `fb "github.com/appnetorg/online-boutique-arpc/proto/proto"`

### 2. Cap'n Proto Code Organization

**Original location:** `proto/onlineboutique.capnp.go` (package `onlineboutique`)
**New location:** `proto/capnp/onlineboutique.capnp.go` (package `capnp`)

**Import alias in code:** `pbcapnp "github.com/appnetorg/online-boutique-arpc/proto/capnp"`

### 3. Converter Functions

#### Format: `ProtoTo{Format}_{MessageType}`

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

**All converters:**
- Take protobuf message as input
- Return marshaled bytes + error
- Handle nested structures (Money, Address, CartItem, etc.)
- Handle lists/arrays

### 4. Size Computation (`sizecomputer.go`)

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
    ProtoJSON   int               `json:"protojson"`
    Errors      map[string]string `json:"errors,omitempty"`
}
```

**How it works:**
1. Uses type switch to determine message type
2. Calls appropriate converter for each format
3. Computes `len(bytes)` for each format
4. Returns -1 for failed conversions
5. Stores error messages in `Errors` map

### 5. Log Entry Structure

**Structure:**
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

**Example output:**
```json
{
  "timestamp": "2026-01-04T10:30:45Z",
  "direction": "request",
  "method": "CartService.AddItem",
  "message_type": "AddItemRequest",
  "sizes": {
    "protobuf": 45,
    "flatbuffers": 52,
    "capnproto": 48,
    "protojson": 78
  },
  "payload": {
    "user_id": "user123",
    "item": {
      "product_id": "OLJCESPC7Z",
      "quantity": 2
    }
  }
}
```

### 6. Message Logger Changes (`messagelogger.go`)

**Logging Strategy:**
Only log messages when they are **sent**, not when received. This ensures each message is logged exactly once.

**Client-side:**
- `ProcessRequest()` - Logs outgoing requests (SEND)
- `ProcessResponse()` - No logging (RECEIVE)

**Server-side:**
- `ProcessRequest()` - No logging (RECEIVE)
- `ProcessResponse()` - Logs outgoing responses (SEND)

**Log Entry Format:**
```go
sizes := ComputeSizes(req.Payload)
entry := LogEntry{
    Timestamp:   time.Now().UTC().Format(time.RFC3339),
    Direction:   "request",  // or "response"
    Method:      req.Method,
    MessageType: GetMessageTypeName(req.Payload),
    Sizes:       sizes,
    Payload:     req.Payload,
}
data, err := MarshalLogEntry(entry)
l.file.Write(append(data, '\n'))
```

---

## How to Extend

### Adding a New Message Type

If you add a new protobuf message type to the schema:

**1. Generate FlatBuffers code:**
```bash
# Add message to proto/onlineboutique.fbs
flatc --go proto/onlineboutique.fbs
```

**2. Generate Cap'n Proto code:**
```bash
# Add message to proto/onlineboutique.capnp
capnp compile -ogo proto/onlineboutique.capnp
# Move generated file
mv proto/onlineboutique.capnp.go proto/capnp/
# Fix package name
sed -i 's/^package proto$/package capnp/' proto/capnp/onlineboutique.capnp.go
```

**3. Add FlatBuffers converter in `converters.go`:**
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

**4. Add Cap'n Proto converter in `converters_capnp.go`:**
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

**5. Add type case in `sizecomputer.go`:**

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

---

## Important Notes

### Error Handling
- **All errors are caught and logged** - RPC never fails due to logging
- Failed conversions result in size = -1
- Error details stored in `Sizes.Errors` map

### Performance Considerations
- Size computation happens on **every** RPC request/response
- Marshaling happens 4 times per message (protobuf, FlatBuffers, Cap'n Proto, protojson)
- This is synchronous and will add latency to RPCs
- Consider making this optional via environment variable if performance is critical

### FlatBuffers Specifics
- Builder pattern: build nested structures first, then parent
- Vectors must be built in reverse order
- Strings must be created before using them
- `builder.Finish(obj)` must be called before `FinishedBytes()`

### Cap'n Proto Specifics
- Always create message and segment first: `capnp.NewMessage(capnp.SingleSegment(nil))`
- Use `NewRoot*` for top-level types
- Use `New*` for nested types
- Lists use `New*_List(seg, int32(len(items)))`
- Text lists use `capnp.NewTextList(seg, int32(len(items)))`

### Package Naming
- **Protobuf:** package `onlineboutique` at `proto/`
- **FlatBuffers:** package `proto` at `proto/proto/`
- **Cap'n Proto:** package `capnp` at `proto/capnp/`
- **Import aliases:**
  - `pb` for protobuf
  - `fb` for FlatBuffers
  - `pbcapnp` for Cap'n Proto

---

## Dependencies

**Added to `go.mod`:**
```
github.com/google/flatbuffers v25.12.19+incompatible
capnproto.org/go/capnp/v3 (already present)
google.golang.org/protobuf (already present)
```

---

## Testing

**Test with a simple RPC:**
```bash
# Start a service (e.g., CartService)
# Make an RPC call
# Check log file in /var/log/arpc-messages/ or MESSAGE_LOG_DIR
# Verify JSON output contains sizes for all formats
```

**Example log location:**
```
/var/log/arpc-messages/client-messages-20260104-103045.jsonl  # Contains outgoing requests
/var/log/arpc-messages/server-messages-20260104-103045.jsonl  # Contains outgoing responses
```

**Note:** Each message is logged exactly once when sent:
- Requests are logged in client log files (when client sends them)
- Responses are logged in server log files (when server sends them)

---

## Troubleshooting

### Build Error: "redeclared in this block"
- **Cause:** Cap'n Proto and Protobuf types in same package
- **Fix:** Ensure Cap'n Proto code is in `proto/capnp/` with package `capnp`

### Build Error: "undefined: pb.NewCartItem"
- **Cause:** Missing `pbcapnp` import alias
- **Fix:** Use `pbcapnp.NewCartItem` for Cap'n Proto types

### Size always -1 for a format
- **Cause:** Converter function failing
- **Fix:** Check error in `Sizes.Errors` field in log output
- Common issues: nil pointers, missing nested structure initialization

### FlatBuffers "vector offset is before current position"
- **Cause:** Building structures in wrong order
- **Fix:** Build all nested structures and vectors BEFORE starting parent structure

---

## Summary of All Message Types Supported

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

All 33 message types have complete converters for both FlatBuffers and Cap'n Proto.
