package db

import "fmt"

// Operation is the PowerSync mutation operation type in ps_crud.
type Operation string

const (
	OperationPut    Operation = "PUT"
	OperationPatch  Operation = "PATCH"
	OperationDelete Operation = "DELETE"
)

// TableName identifies a table in ps_crud payloads.
type TableName string

const (
	TableMessages      TableName = "messages"
	TableConversations TableName = "conversations"
	TableAccounts      TableName = "accounts"
	TableWorkspaces    TableName = "workspaces"
)

// MutationID is the auto-increment queue ID in ps_crud.
type MutationID int64

// TransactionID groups queued mutations created in the same transaction.
type TransactionID int64

// BucketName identifies a PowerSync bucket row in ps_buckets.
type BucketName string

const (
	// LocalBucket is the upload-ack bucket used by PowerSync local mutation flow.
	LocalBucket BucketName = "$local"
)

// OpID is a PowerSync operation/checkpoint ID.
type OpID int64

// Mutation is one queued mutation entry from ps_crud.
type Mutation struct {
	ID       MutationID
	TxID     *TransactionID
	Op       Operation
	Table    TableName
	RowID    string
	Data     map[string]any
	Old      map[string]any
	Metadata *string
}

func (m Mutation) String() string {
	if m.TxID == nil {
		return fmt.Sprintf("id=%d op=%s table=%s row=%s", m.ID, m.Op, m.Table, m.RowID)
	}
	return fmt.Sprintf("id=%d tx=%d op=%s table=%s row=%s", m.ID, *m.TxID, m.Op, m.Table, m.RowID)
}
