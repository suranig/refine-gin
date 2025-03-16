package resource

// Operation reprezentuje operację na zasobie
type Operation string

const (
	OperationList   Operation = "LIST"
	OperationCreate Operation = "CREATE"
	OperationRead   Operation = "READ"
	OperationUpdate Operation = "UPDATE"
	OperationDelete Operation = "DELETE"
)
