# Serialization Format Size Comparison

The Online Boutique application includes a message logger that computes serialization sizes for all RPC messages across four different formats: Protobuf, FlatBuffers, Cap'n Proto, and Symphony. This allows for direct comparison of how efficiently each format serializes the same data.

## Example Log Entry

Here's an example log entry showing size differences for a `CurrencyConversionRequest`:

```json
{
  "timestamp": "2026-01-05T23:57:40Z",
  "direction": "request",
  "method": "Convert",
  "message_type": "CurrencyConversionRequest",
  "sizes": {
    "protobuf": 26,
    "flatbuffers": 96,
    "capnproto": 88,
    "symphony": 82
  },
  "payload": {
    "from": {
      "currency_code": "USD",
      "units": 18,
      "nanos": 490000000
    },
    "to_code": "CNY",
    "user_id": "test"
  }
}
```

## Actual Data Content

The message contains:
- **Money object** (`from`):
  - `currency_code`: "USD" (3 bytes)
  - `units`: 18 (int64, small value)
  - `nanos`: 490000000 (int32)
- **to_code**: "CNY" (3 bytes)
- **user_id**: "test" (4 bytes)

## Size Breakdown by Format

### 1. Protobuf: 26 bytes (Smallest)

**Why it's the smallest:**
- Uses variable-length encoding (varints) for integers
- Compact field tags (1 byte per field)
- No padding or alignment requirements
- Minimal structural overhead

**Byte breakdown:**
```
Field 1 (from - Money):
  - Field tag (1 byte): field number 1, wire type 2 (length-delimited)
  - Length prefix (1 byte): size of nested Money message
  - Money fields:
    - Field 1 (currency_code): tag (1) + length (1) + "USD" (3) = 5 bytes
    - Field 2 (units): tag (1) + varint(18) (1) = 2 bytes
    - Field 3 (nanos): tag (1) + varint(490000000) (4) = 5 bytes
  - Total Money: 5 + 2 + 5 = 12 bytes
  - Total for field 1: 1 (tag) + 1 (length) + 12 = 14 bytes

Field 2 (to_code):
  - Tag (1) + length (1) + "CNY" (3) = 5 bytes

Field 3 (user_id):
  - Tag (1) + length (1) + "test" (4) = 6 bytes

Total: 14 + 5 + 6 = 25 bytes + 1 byte overhead ≈ 26 bytes
```

### 2. FlatBuffers: 96 bytes (3.7x larger)

**Why it's larger:**
- VTable (virtual function table) overhead for each object
- 4-byte offsets for all references
- Separate string storage with offsets
- Alignment requirements (8-byte boundaries)
- Table-based structure adds overhead

**Byte breakdown:**
```
Structure:
1. Root object table (CurrencyConversionRequest):
   - VTable offset (4 bytes)
   - Field offsets (3 fields × 4 bytes = 12 bytes)
   - VTable itself:
     - Size (2 bytes)
     - Object size (2 bytes)
     - Field offsets (3 × 2 bytes = 6 bytes)
   - Total: ~20 bytes

2. Nested Money object:
   - Object offset (4 bytes)
   - VTable offset (4 bytes)
   - Field offsets (3 × 4 bytes = 12 bytes)
   - VTable (similar structure)
   - Data: currency_code offset, units (8), nanos (4)
   - Total: ~30 bytes

3. String data (stored separately):
   - "USD" (3) + offset (4) = 7 bytes
   - "CNY" (3) + offset (4) = 7 bytes
   - "test" (4) + offset (4) = 8 bytes
   - Total: ~22 bytes

4. Alignment and padding: ~24 bytes

Total: ~96 bytes
```

### 3. Cap'n Proto: 88 bytes (3.4x larger)

**Why it's larger:**
- 8-byte pointers for all references
- Alignment to 8-byte boundaries
- Segment header overhead
- String length prefixes + padding

**Byte breakdown:**
```
Structure:
1. Segment header:
   - Segment count (4 bytes)
   - Root pointer (8 bytes)
   - Total: ~12 bytes

2. CurrencyConversionRequest struct:
   - Pointer to Money (8 bytes)
   - String pointers (2 × 8 bytes = 16 bytes)
   - Struct data: ~8 bytes
   - Total: ~32 bytes

3. Money struct:
   - String pointer (8 bytes)
   - units (8 bytes)
   - nanos (4 bytes)
   - Padding (4 bytes for alignment)
   - Total: ~24 bytes

4. String data:
   - "USD" (3) + length (1) + padding = 8 bytes
   - "CNY" (3) + length (1) + padding = 8 bytes
   - "test" (4) + length (1) + padding = 8 bytes
   - Total: ~24 bytes

Total: ~88 bytes
```

### 4. Symphony: 82 bytes (3.2x larger)

**Why it's larger:**
- Public/private segment headers (13 bytes each for nested messages)
- 4-byte length prefixes for all strings
- Table structure overhead
- Nested messages duplicate header overhead

**Byte breakdown:**
```
Structure:
1. Public segment header:
   - Version (1 byte)
   - Offset to private (4 bytes)
   - Service ID (4 bytes)
   - Method ID (4 bytes)
   - Total: 13 bytes

2. Private segment (CurrencyConversionRequest):
   - Version (1 byte)
   - Table entries (12 bytes: 3 fields × 4 bytes)
   - Field offsets and data
   - Total: ~25 bytes

3. Nested Money (with its own headers):
   - Public segment (13 bytes)
   - Private segment (~20 bytes)
   - Total: ~33 bytes

4. String data with length prefixes:
   - "USD": 4 (length) + 3 = 7 bytes
   - "CNY": 4 (length) + 3 = 7 bytes
   - "test": 4 (length) + 4 = 8 bytes
   - Total: ~22 bytes

Total: ~82 bytes
```

## Summary Comparison

| Format | Size | Overhead Type | Best For |
|--------|------|---------------|----------|
| **Protobuf** | 26 bytes | Minimal (varints, tags) | Small messages, network efficiency |
| **Symphony** | 82 bytes | Headers + tables | Custom protocol needs |
| **Cap'n Proto** | 88 bytes | Pointers + alignment | Zero-copy reads |
| **FlatBuffers** | 96 bytes | VTable + offsets | Zero-copy reads, game engines |

## Key Insights

1. **Protobuf is smallest for small messages** - Uses varint encoding and minimal overhead, making it ideal for network transmission where size matters most.

2. **Zero-copy formats trade size for performance** - FlatBuffers and Cap'n Proto add overhead (VTables, pointers, alignment) to enable zero-copy reads, which is valuable for large messages that are read multiple times.

3. **Symphony adds protocol headers** - The public/private segment structure includes service and method IDs, adding overhead but providing protocol-level features.

4. **Overhead ratio decreases with message size** - For this 26-byte payload, Protobuf is 3.7x smaller than FlatBuffers, but the absolute difference (70 bytes) is small. For larger messages (e.g., product catalogs with many items), the relative difference typically narrows.

5. **Choose format based on use case**:
   - **Network efficiency**: Protobuf
   - **Zero-copy performance**: FlatBuffers or Cap'n Proto
   - **Protocol features**: Symphony

