// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

// Record describes database record interface without data.
type Record interface {
	// GetTable returns table name.
	GetTable() string
	// GetKeyPartitionField returns partition field name.
	GetKeyPartitionField() string
	// GetKeySortField returns sort field name.
	GetKeySortField() string
}

// RecordBuffer describes database record for write database response.
type RecordBuffer interface {
	Record
	Clear()
}

// Key describes key methods.
type Key interface{ GetKey() interface{} }

// KeyRecord describes database record interface without data but with key.
type KeyRecord interface {
	Record
	Key
}

// KeyRecordBuffer describes database record interface without data but with
// key and allows to write database response.
type KeyRecordBuffer interface {
	RecordBuffer
	Key
}

// DataRecord describes database record interface with data.
type DataRecord interface {
	Record
	GetData() interface{}
}

// IndexRecord describes database index interface.
type IndexRecord interface {
	RecordBuffer
	// GetIndex returns index name.
	GetIndex() string
	// GetIndexPartitionField returns index partition field name.
	GetIndexPartitionField() string
	// GetIndexSortField returns index sort field name.
	GetIndexSortField() string
	// GetIndexSortField returns index additional fields for record.
	GetProjection() []string
}

////////////////////////////////////////////////////////////////////////////////
