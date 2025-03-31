package repository

// IDSetter defines an interface for models that can set their own ID field value
type IDSetter interface {
	SetID(id interface{})
}

// TrySetID attempts to set ID on the model if it implements IDSetter interface
func TrySetID(model interface{}, id interface{}) bool {
	if setter, ok := model.(IDSetter); ok {
		setter.SetID(id)
		return true
	}
	return false
}
