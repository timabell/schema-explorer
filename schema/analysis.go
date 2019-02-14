package schema

type ColumnAnalysis struct {
	Column      *Column
	ValueCounts []ValueInfo
}

type ValueInfo struct {
	Value    interface{}
	Quantity int
}
