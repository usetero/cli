# PowerSync Upload Protocol - Investigation Notes

## The Problem

After writing to SQLite and uploading via the CRUD queue, the database ends up in a corrupt state where sync hangs forever.

**Symptoms:**
- `ps_crud` table is empty (upload completed)
- But `ps_buckets` shows `target_op > last_op` for the `$local` bucket
- This causes `sync_local` to return 0 forever, blocking checkpoint completion

## Root Cause

We were only deleting from `ps_crud` after upload. PowerSync requires a specific protocol to signal upload completion.

## The PowerSync Upload Protocol

### Key Tables

**`ps_crud`** - Queue of pending local changes to upload
- `id`: Auto-incrementing client ID
- `tx_id`: Transaction ID (groups operations in same transaction)  
- `data`: JSON-encoded CRUD entry with operation type, table, data

**`ps_buckets`** - Tracks sync state per bucket
- `name`: Bucket name (or `'$local'` for local/write state)
- `last_op`: Last operation ID confirmed by server
- `target_op`: Target operation ID we're waiting for server to confirm

### The `$local` Bucket

The `$local` bucket tracks local write state. The sync engine checks:

```sql
SELECT 1 FROM ps_buckets WHERE target_op > last_op AND name = '$local'
```

If this returns a row, sync is blocked (waiting for server confirmation).

### Protocol Flow

1. **Local Write**
   - Data written to table
   - Trigger inserts entry into `ps_crud`
   - Trigger does: `INSERT OR REPLACE INTO ps_buckets(name, last_op, target_op) VALUES('$local', 0, MAX_OP_ID)`
   - `MAX_OP_ID = 9223372036854775807` (max int64, sentinel for "no specific checkpoint")

2. **Upload to Backend**
   - Read entries from `ps_crud`
   - Send each to your backend API

3. **Fetch Write Checkpoint** (THIS IS WHAT WE WERE MISSING)
   - Call PowerSync server: `GET /write-checkpoint2.json?client_id={clientID}`
   - Server returns: `{"data": {"write_checkpoint": "12345"}}`
   - This checkpoint represents "server has processed all your writes up to this point"

4. **Complete the Batch** (atomic transaction)
   - Delete from `ps_crud`: `DELETE FROM ps_crud WHERE id <= ?`
   - Set target_op to checkpoint: `UPDATE ps_buckets SET target_op = ? WHERE name = '$local'`

5. **Server Confirms via Sync Stream**
   - Server sends checkpoint with `write_checkpoint` field
   - Client sets: `UPDATE ps_buckets SET last_op = ? WHERE name = '$local'`
   - Now `target_op == last_op`, sync can proceed

### Why MAX_OP_ID Doesn't Work

Setting `target_op = MAX_OP_ID` after upload seems logical ("no checkpoint to wait for"), but:
- `MAX_OP_ID > last_op (0)` is always TRUE
- So sync stays blocked

The correct approach is to get an actual checkpoint from the server.

## Key Code Locations

### JS Client (reference implementation)

**Upload orchestration:** `/tmp/powersync-js/packages/common/src/client/sync/stream/AbstractStreamingSyncImplementation.ts`
- Lines 368-453: `_uploadAllCrud()` - the upload loop
- Line 417: After uploads complete, calls `updateLocalTarget(() => this.getWriteCheckpoint())`

**Fetch write checkpoint:** Same file, lines 359-367:
```typescript
async getWriteCheckpoint(): Promise<string> {
  const clientId = await this.options.adapter.getClientId();
  let path = `/write-checkpoint2.json?client_id=${clientId}`;
  const response = await this.options.remote.get(path);
  const checkpoint = response['data']['write_checkpoint'] as string;
  return checkpoint;
}
```

**Batch completion:** `/tmp/powersync-js/packages/common/src/client/sync/bucket/SqliteBucketStorage.ts`
- Lines 320-350: `getCrudBatch()` returns a batch with a `complete()` callback
- The `complete()` callback atomically deletes entries and updates `target_op`

### Rust Core (SQLite extension)

**Sync blocking check:** `/tmp/powersync-sqlite-core/crates/core/src/sync_local.rs`
- Lines 95-120: `can_apply_sync_changes()`
- Checks both `target_op > last_op` AND `ps_crud` not empty

**Local bucket creation:** `/tmp/powersync-sqlite-core/crates/core/src/crud_vtab.rs`
- Line 250: On first write, creates `$local` bucket with `target_op = MAX_OP_ID`

### PowerSync Service

Located at `/tmp/powersync-service/`

**Endpoints:** `/tmp/powersync-service/packages/service-core/src/routes/endpoints/`
- `checkpointing.ts` - Write checkpoint endpoint
- `sync-stream.ts` - Sync stream endpoint

## PowerSync Server API

Based on the JS client, these are the endpoints:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/sync/stream` | POST | SSE stream for sync data |
| `/write-checkpoint2.json?client_id=X` | GET | Fetch write checkpoint |
| `/write-checkpoint2.json?client_id=X` | PUT | Fetch checkpoint with bucket validation |

## Getting Client ID

The client ID is stored in the PowerSync SQLite extension:

```sql
SELECT powersync_client_id()
```

## Correct Implementation

```go
// 1. Upload all entries to backend
for _, entry := range entries {
    uploadToBackend(entry)
}

// 2. Get client ID
var clientID string
db.QueryRow("SELECT powersync_client_id()").Scan(&clientID)

// 3. Fetch write checkpoint from PowerSync server
resp := http.Get(powersyncEndpoint + "/write-checkpoint2.json?client_id=" + clientID)
checkpoint := resp.Data.WriteCheckpoint  // e.g., "12345"

// 4. Complete batch atomically
tx.Exec("DELETE FROM ps_crud WHERE id <= ?", lastID)
tx.Exec("UPDATE ps_buckets SET target_op = ? WHERE name = '$local'", checkpoint)
```

## Files Changed in Our Fix

1. **`internal/powersync/upload.go`** (NEW) - `CrudUploader` handles the complete protocol
2. **`internal/powersync/crud.go`** - Simplified, protocol handling moved to CrudUploader
3. **`internal/upload/upload.go`** - Uses CrudUploader instead of manual queue management
4. **`internal/sqlite/database.go`** - Added `WithTx` for transactions

## Open Questions

1. Should we have a generic PowerSync client for all HTTP endpoints?
2. How should we structure the code for proper separation of concerns?
