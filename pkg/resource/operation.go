package resource

// Operation represents a CRUD operation
type Operation string

// Available operations
const (
	// OperationList represents the LIST operation (GET /resources)
	OperationList Operation = "list"

	// OperationCreate represents the CREATE operation (POST /resources)
	OperationCreate Operation = "create"

	// OperationRead represents the READ operation (GET /resources/:id)
	OperationRead Operation = "read"

	// OperationUpdate represents the UPDATE operation (PUT /resources/:id)
	OperationUpdate Operation = "update"

	// OperationDelete represents the DELETE operation (DELETE /resources/:id)
	OperationDelete Operation = "delete"

	// OperationCount represents the COUNT operation for counting resources
	OperationCount Operation = "count"
)
