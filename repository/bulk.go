package repository

// BulkOpType represents the type of bulk operation.
type BulkOpType int

const (
	BulkOpInsert BulkOpType = iota
	BulkOpUpdate
	BulkOpReplace
	BulkOpDelete
)

// BulkOp represents a single operation in a bulk write.
type BulkOp struct {
	Type    BulkOpType
	Filter  any // For update, replace, delete operations
	Doc     any // For insert and replace operations
	Update  any // For update operations
	Upsert  bool
	Collation *Collation
}

// Collation specifies collation options for string comparison.
type Collation struct {
	Locale          string
	CaseLevel       bool
	CaseFirst       string
	Strength        int
	NumericOrdering bool
	Alternate       string
	MaxVariable     string
	Backwards       bool
}

// InsertOp creates a bulk insert operation.
func InsertOp(doc any) BulkOp {
	return BulkOp{
		Type: BulkOpInsert,
		Doc:  doc,
	}
}

// UpdateOp creates a bulk update operation.
func UpdateOp(filter, update any) BulkOp {
	return BulkOp{
		Type:   BulkOpUpdate,
		Filter: filter,
		Update: update,
	}
}

// UpdateOpWithUpsert creates a bulk update operation with upsert enabled.
func UpdateOpWithUpsert(filter, update any) BulkOp {
	return BulkOp{
		Type:   BulkOpUpdate,
		Filter: filter,
		Update: update,
		Upsert: true,
	}
}

// ReplaceOp creates a bulk replace operation.
func ReplaceOp(filter, doc any) BulkOp {
	return BulkOp{
		Type:   BulkOpReplace,
		Filter: filter,
		Doc:    doc,
	}
}

// ReplaceOpWithUpsert creates a bulk replace operation with upsert enabled.
func ReplaceOpWithUpsert(filter, doc any) BulkOp {
	return BulkOp{
		Type:   BulkOpReplace,
		Filter: filter,
		Doc:    doc,
		Upsert: true,
	}
}

// DeleteOp creates a bulk delete operation.
func DeleteOp(filter any) BulkOp {
	return BulkOp{
		Type:   BulkOpDelete,
		Filter: filter,
	}
}
