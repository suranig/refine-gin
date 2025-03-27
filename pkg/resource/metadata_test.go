package resource

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResource for testing metadata generation
type MetadataMockResource struct {
	mock.Mock
}

func (m *MetadataMockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MetadataMockResource) GetFields() []Field {
	args := m.Called()
	return args.Get(0).([]Field)
}

func (m *MetadataMockResource) GetOperations() []Operation {
	args := m.Called()
	return args.Get(0).([]Operation)
}

func (m *MetadataMockResource) HasOperation(op Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MetadataMockResource) GetDefaultSort() *Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*Sort)
}

func (m *MetadataMockResource) GetFilters() []Filter {
	args := m.Called()
	return args.Get(0).([]Filter)
}

func (m *MetadataMockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MetadataMockResource) GetRelations() []Relation {
	args := m.Called()
	return args.Get(0).([]Relation)
}

func (m *MetadataMockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MetadataMockResource) GetRelation(name string) *Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(Relation)
	return &relation
}

func (m *MetadataMockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetField(name string) *Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(Field)
	return &field
}

func (m *MetadataMockResource) GetSearchable() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MetadataMockResource) GetFilterableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MetadataMockResource) GetSortableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MetadataMockResource) GetRequiredFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MetadataMockResource) GetTableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MetadataMockResource) GetFormFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func TestGenerateResourceMetadata(t *testing.T) {
	// Create a sample resource
	fields := []Field{
		{
			Name:  "id",
			Type:  "int",
			Label: "ID",
			List: &ListConfig{
				Width: 100,
			},
			Form: &FormConfig{
				Placeholder: "Enter ID",
			},
			Validation: &Validation{
				Required: true,
			},
		},
		{
			Name:  "name",
			Type:  "string",
			Label: "Name",
			Validation: &Validation{
				MinLength: 3,
				MaxLength: 50,
			},
		},
	}

	relations := []Relation{
		{
			Name:           "posts",
			Type:           RelationTypeOneToMany,
			Resource:       "posts",
			Field:          "author_id",
			ReferenceField: "id",
		},
	}

	operations := []Operation{
		OperationList,
		OperationCreate,
		OperationRead,
	}

	defaultSort := &Sort{
		Field: "id",
		Order: "asc",
	}

	filters := []Filter{
		{
			Field:    "name",
			Operator: "eq",
			Value:    "John",
		},
	}

	resource := &DefaultResource{
		Name:             "users",
		Label:            "Users",
		Icon:             "user",
		Fields:           fields,
		Operations:       operations,
		DefaultSort:      defaultSort,
		Filters:          filters,
		SearchableFields: []string{"name"},
		IDFieldName:      "id",
		FilterableFields: []string{"id", "name"},
		SortableFields:   []string{"id", "name"},
		RequiredFields:   []string{"id"},
		Relations:        relations,
	}

	// Generate metadata
	metadata := GenerateResourceMetadata(resource)

	// Verify metadata
	assert.Equal(t, "users", metadata.Name)
	assert.Equal(t, "Users", metadata.Label)
	assert.Equal(t, "user", metadata.Icon)
	assert.Len(t, metadata.Operations, 3)
	assert.Len(t, metadata.Fields, 2)
	assert.Len(t, metadata.Relations, 1)
	assert.Equal(t, "id", metadata.DefaultSort.Field)
	assert.Equal(t, "asc", metadata.DefaultSort.Order)
	assert.Len(t, metadata.Filters, 1)
	assert.Equal(t, []string{"name"}, metadata.Searchable)
	assert.Equal(t, "id", metadata.IDFieldName)
}

func TestGenerateFieldsMetadata(t *testing.T) {
	// Create sample fields
	fields := []Field{
		{
			Name:  "id",
			Type:  "int",
			Label: "ID",
			Validation: &Validation{
				Required: true,
			},
		},
		{
			Name:  "name",
			Type:  "string",
			Label: "Name",
			Validation: &Validation{
				MinLength: 3,
				MaxLength: 50,
			},
		},
	}

	// Generate fields metadata
	fieldsMeta := GenerateFieldsMetadata(fields)

	// Verify fields metadata
	assert.Len(t, fieldsMeta, 2)
	assert.Equal(t, "id", fieldsMeta[0].Name)
	assert.Equal(t, "int", fieldsMeta[0].Type)
	assert.Equal(t, "ID", fieldsMeta[0].Label)
	assert.True(t, fieldsMeta[0].Required)
	assert.True(t, fieldsMeta[0].Filterable)  // Default value
	assert.True(t, fieldsMeta[0].Sortable)    // Default value
	assert.False(t, fieldsMeta[0].Searchable) // Default value
	assert.False(t, fieldsMeta[0].Unique)     // Default value

	assert.Equal(t, "name", fieldsMeta[1].Name)
	assert.Equal(t, "string", fieldsMeta[1].Type)
	assert.Equal(t, "Name", fieldsMeta[1].Label)
	assert.False(t, fieldsMeta[1].Required)
	assert.True(t, fieldsMeta[1].Filterable)  // Default value
	assert.True(t, fieldsMeta[1].Sortable)    // Default value
	assert.False(t, fieldsMeta[1].Searchable) // Default value
	assert.False(t, fieldsMeta[1].Unique)     // Default value
}

func TestGenerateValidatorsMetadata(t *testing.T) {
	// Test with empty validators
	emptyValidators := []Validator{}
	emptyMetadata := GenerateValidatorsMetadata(emptyValidators)
	assert.Empty(t, emptyMetadata)

	// Test with string validator
	stringValidator := StringValidator{
		MinLength: 5,
		MaxLength: 100,
		Pattern:   "^[a-zA-Z0-9]+$",
	}

	// Test with number validator
	numberValidator := NumberValidator{
		Min: 10,
		Max: 1000,
	}

	validators := []Validator{stringValidator, numberValidator}
	metadata := GenerateValidatorsMetadata(validators)

	assert.Equal(t, 2, len(metadata))

	// Check string validator metadata
	assert.Equal(t, "string", metadata[0].Type)

	// Check validator rules
	if minLength, ok := metadata[0].Rules["minLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minLengthInt int
		switch v := minLength.(type) {
		case float64:
			minLengthInt = int(v)
		case int:
			minLengthInt = v
		}
		assert.Equal(t, 5, minLengthInt)
	} else {
		t.Error("minLength rule should exist")
	}

	if maxLength, ok := metadata[0].Rules["maxLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxLengthInt int
		switch v := maxLength.(type) {
		case float64:
			maxLengthInt = int(v)
		case int:
			maxLengthInt = v
		}
		assert.Equal(t, 100, maxLengthInt)
	} else {
		t.Error("maxLength rule should exist")
	}

	assert.Equal(t, "^[a-zA-Z0-9]+$", metadata[0].Rules["pattern"])

	// Check number validator metadata
	assert.Equal(t, "number", metadata[1].Type)

	// Check validator rules
	if min, ok := metadata[1].Rules["min"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minInt int
		switch v := min.(type) {
		case float64:
			minInt = int(v)
		case int:
			minInt = v
		}
		assert.Equal(t, 10, minInt)
	} else {
		t.Error("min rule should exist")
	}

	if max, ok := metadata[1].Rules["max"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxInt int
		switch v := max.(type) {
		case float64:
			maxInt = int(v)
		case int:
			maxInt = v
		}
		assert.Equal(t, 1000, maxInt)
	} else {
		t.Error("max rule should exist")
	}
}

func TestGenerateRelationsMetadata(t *testing.T) {
	// Test with empty relations
	emptyRelations := []Relation{}
	emptyMetadata := GenerateRelationsMetadata(emptyRelations)
	assert.Empty(t, emptyMetadata)

	// Test with various relations
	relations := []Relation{
		{
			Name:             "comments",
			Type:             RelationTypeOneToMany,
			Resource:         "comment",
			Field:            "post_id",
			ReferenceField:   "id",
			IncludeByDefault: true,
			MinItems:         1,
			MaxItems:         100,
		},
		{
			Name:           "author",
			Type:           RelationTypeManyToOne,
			Resource:       "user",
			Field:          "author_id",
			ReferenceField: "id",
			DisplayField:   "name",
			ValueField:     "id",
			Required:       true,
			Cascade:        true,
			OnDelete:       "CASCADE",
			OnUpdate:       "CASCADE",
		},
		{
			Name:         "tags",
			Type:         RelationTypeManyToMany,
			Resource:     "tag",
			PivotTable:   "post_tags",
			PivotFields:  map[string]string{"post_id": "id", "tag_id": "id"},
			DisplayField: "name",
			ValueField:   "id",
		},
	}

	metadata := GenerateRelationsMetadata(relations)
	assert.Equal(t, 3, len(metadata))

	// Check one-to-many relation
	assert.Equal(t, "comments", metadata[0].Name)
	assert.Equal(t, RelationTypeOneToMany, metadata[0].Type)
	assert.Equal(t, "comment", metadata[0].Resource)
	assert.Equal(t, "post_id", metadata[0].Field)
	assert.Equal(t, "id", metadata[0].ReferenceField)
	assert.True(t, metadata[0].IncludeByDefault)
	assert.Equal(t, 1, metadata[0].MinItems)
	assert.Equal(t, 100, metadata[0].MaxItems)

	// Check many-to-one relation
	assert.Equal(t, "author", metadata[1].Name)
	assert.Equal(t, RelationTypeManyToOne, metadata[1].Type)
	assert.Equal(t, "user", metadata[1].Resource)
	assert.Equal(t, "author_id", metadata[1].Field)
	assert.Equal(t, "id", metadata[1].ReferenceField)
	assert.Equal(t, "name", metadata[1].DisplayField)
	assert.Equal(t, "id", metadata[1].ValueField)
	assert.True(t, metadata[1].Required)
	assert.True(t, metadata[1].Cascade)
	assert.Equal(t, "CASCADE", metadata[1].OnDelete)
	assert.Equal(t, "CASCADE", metadata[1].OnUpdate)

	// Check many-to-many relation
	assert.Equal(t, "tags", metadata[2].Name)
	assert.Equal(t, RelationTypeManyToMany, metadata[2].Type)
	assert.Equal(t, "tag", metadata[2].Resource)
	assert.Equal(t, "post_tags", metadata[2].PivotTable)
	assert.Equal(t, map[string]string{"post_id": "id", "tag_id": "id"}, metadata[2].PivotFields)
	assert.Equal(t, "name", metadata[2].DisplayField)
	assert.Equal(t, "id", metadata[2].ValueField)
}

func TestGenerateFieldsMetadataWithJson(t *testing.T) {
	// Create a field with JSON configuration
	fields := []Field{
		{
			Name: "config",
			Type: "json",
			Json: &JsonConfig{
				DefaultExpanded: true,
				EditorType:      "form",
				Properties: []JsonProperty{
					{
						Path:  "email",
						Label: "Email Configuration",
						Type:  "object",
						Properties: []JsonProperty{
							{
								Path:  "email.host",
								Label: "SMTP Host",
								Type:  "string",
								Validation: &Validation{
									Required: true,
								},
								Form: &FormConfig{
									Placeholder: "smtp.example.com",
									Help:        "Enter your SMTP server host",
								},
							},
							{
								Path:  "email.port",
								Label: "SMTP Port",
								Type:  "number",
								Validation: &Validation{
									Required: true,
									Min:      0,
									Max:      65535,
								},
							},
						},
					},
					{
						Path:  "oauth",
						Label: "OAuth Settings",
						Type:  "object",
						Properties: []JsonProperty{
							{
								Path:  "oauth.google_client_id",
								Label: "Google Client ID",
								Type:  "string",
							},
							{
								Path:  "oauth.google_client_secret",
								Label: "Google Client Secret",
								Type:  "string",
							},
						},
					},
				},
			},
		},
	}

	// Generate metadata for the field
	metadata := GenerateFieldsMetadata(fields)

	// Verify JSON metadata
	assert.Len(t, metadata, 1)
	assert.Equal(t, "config", metadata[0].Name)
	assert.Equal(t, "json", metadata[0].Type)

	// Check JSON configuration
	assert.NotNil(t, metadata[0].Json)
	assert.True(t, metadata[0].Json.DefaultExpanded)
	assert.Equal(t, "form", metadata[0].Json.EditorType)

	// Check properties
	assert.Len(t, metadata[0].Json.Properties, 2)

	// Check email configuration
	emailProperty := metadata[0].Json.Properties[0]
	assert.Equal(t, "email", emailProperty.Path)
	assert.Equal(t, "Email Configuration", emailProperty.Label)
	assert.Equal(t, "object", emailProperty.Type)
	assert.Len(t, emailProperty.Properties, 2)

	// Check nested email.host property
	hostProperty := emailProperty.Properties[0]
	assert.Equal(t, "email.host", hostProperty.Path)
	assert.Equal(t, "SMTP Host", hostProperty.Label)
	assert.Equal(t, "string", hostProperty.Type)
	assert.NotNil(t, hostProperty.Validation)
	assert.True(t, hostProperty.Validation.Required)
	assert.NotNil(t, hostProperty.Form)
	assert.Equal(t, "smtp.example.com", hostProperty.Form.Placeholder)
	assert.Equal(t, "Enter your SMTP server host", hostProperty.Form.Help)

	// Check nested email.port property
	portProperty := emailProperty.Properties[1]
	assert.Equal(t, "email.port", portProperty.Path)
	assert.Equal(t, "SMTP Port", portProperty.Label)
	assert.Equal(t, "number", portProperty.Type)
	assert.NotNil(t, portProperty.Validation)
	assert.True(t, portProperty.Validation.Required)
	assert.Equal(t, float64(0), portProperty.Validation.Min)
	assert.Equal(t, float64(65535), portProperty.Validation.Max)

	// Check OAuth configuration
	oauthProperty := metadata[0].Json.Properties[1]
	assert.Equal(t, "oauth", oauthProperty.Path)
	assert.Equal(t, "OAuth Settings", oauthProperty.Label)
	assert.Equal(t, "object", oauthProperty.Type)
	assert.Len(t, oauthProperty.Properties, 2)

	// Check Google OAuth properties
	googleClientIdProperty := oauthProperty.Properties[0]
	assert.Equal(t, "oauth.google_client_id", googleClientIdProperty.Path)
	assert.Equal(t, "Google Client ID", googleClientIdProperty.Label)
	assert.Equal(t, "string", googleClientIdProperty.Type)

	googleClientSecretProperty := oauthProperty.Properties[1]
	assert.Equal(t, "oauth.google_client_secret", googleClientSecretProperty.Path)
	assert.Equal(t, "Google Client Secret", googleClientSecretProperty.Label)
	assert.Equal(t, "string", googleClientSecretProperty.Type)
}

// Add test for automatic extraction of JSON properties from struct
func TestExtractJsonSchemaAndProperties(t *testing.T) {
	// Define a test struct similar to domain.go
	type EmailConfig struct {
		Host     string `json:"host,omitempty"`
		Port     int    `json:"port,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}

	type OAuthConfig struct {
		GoogleClientID     string `json:"google_client_id,omitempty"`
		GoogleClientSecret string `json:"google_client_secret,omitempty"`
		GoogleRedirectURL  string `json:"google_redirect_url,omitempty"`
	}

	type DomainConfig struct {
		Email     EmailConfig `json:"email,omitempty"`
		OAuth     OAuthConfig `json:"oauth,omitempty"`
		Active    bool        `json:"active,omitempty"`
		CreatedAt string      `json:"created_at,omitempty"`
	}

	type Domain struct {
		ID     uint         `json:"id"`
		Name   string       `json:"name"`
		Config DomainConfig `json:"config"`
	}

	// Create a model field for testing
	testDomain := Domain{}
	modelType := reflect.TypeOf(testDomain)
	configField := modelType.Field(2) // Config field is the third field

	// Test isJsonField function
	assert.True(t, isJsonField(configField.Type))

	// Extract JSON schema and properties
	schema, properties := extractJsonSchemaAndProperties(configField.Type)

	// Verify schema
	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])
	assert.NotNil(t, schema["properties"])

	// Verify properties
	assert.NotEmpty(t, properties)

	// Get all property paths for easier testing
	var propertyPaths []string
	for _, prop := range properties {
		propertyPaths = append(propertyPaths, prop.Path)
	}

	// Check for expected top-level properties
	assert.Contains(t, propertyPaths, "email")
	assert.Contains(t, propertyPaths, "oauth")
	assert.Contains(t, propertyPaths, "active")
	assert.Contains(t, propertyPaths, "created_at")

	// Check for nested properties
	assert.Contains(t, propertyPaths, "email.host")
	assert.Contains(t, propertyPaths, "email.port")
	assert.Contains(t, propertyPaths, "email.username")
	assert.Contains(t, propertyPaths, "email.password")

	assert.Contains(t, propertyPaths, "oauth.google_client_id")
	assert.Contains(t, propertyPaths, "oauth.google_client_secret")
	assert.Contains(t, propertyPaths, "oauth.google_redirect_url")

	// Check types
	for _, prop := range properties {
		if prop.Path == "active" {
			assert.Equal(t, "boolean", prop.Type)
		} else if prop.Path == "email.port" {
			assert.Equal(t, "number", prop.Type)
		} else if prop.Path == "email" || prop.Path == "oauth" {
			assert.Equal(t, "object", prop.Type)
		} else {
			assert.Equal(t, "string", prop.Type)
		}
	}
}
