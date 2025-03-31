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
								Validation: &JsonValidation{
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
								Validation: &JsonValidation{
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

func TestGenerateJsonConfigMetadataWithNestedOptions(t *testing.T) {
	// Create a field with nested JSON configuration
	jsonConfig := &JsonConfig{
		DefaultExpanded: true,
		EditorType:      "form",
		Nested:          true,
		RenderAs:        "tabs",
		TabsConfig: &JsonTabsConfig{
			TabPosition:      "top",
			DefaultActiveTab: "basic",
			Tabs: []JsonTab{
				{
					Key:    "basic",
					Title:  "Basic Settings",
					Icon:   "settings",
					Fields: []string{"name", "description", "active"},
				},
				{
					Key:    "advanced",
					Title:  "Advanced",
					Icon:   "advanced-settings",
					Fields: []string{"config.email", "config.oauth"},
				},
			},
		},
		ObjectLabels: map[string]string{
			"config.email": "Email Settings",
			"config.oauth": "OAuth Configuration",
		},
		Properties: []JsonProperty{
			{
				Path:  "name",
				Label: "Name",
				Type:  "string",
				Validation: &JsonValidation{
					Required:  true,
					MinLength: 3,
				},
			},
			{
				Path:  "config",
				Label: "Configuration",
				Type:  "object",
				Properties: []JsonProperty{
					{
						Path:  "config.email",
						Label: "Email",
						Type:  "object",
					},
					{
						Path:  "config.oauth",
						Label: "OAuth",
						Type:  "object",
					},
				},
			},
		},
	}

	// Generate metadata
	metadata := GenerateJsonConfigMetadata(jsonConfig)

	// Verify base metadata
	assert.NotNil(t, metadata)
	assert.True(t, metadata.DefaultExpanded)
	assert.Equal(t, "form", metadata.EditorType)
	assert.True(t, metadata.Nested)
	assert.Equal(t, "tabs", metadata.RenderAs)

	// Verify tabs configuration
	assert.NotNil(t, metadata.TabsConfig)
	assert.Equal(t, "top", metadata.TabsConfig.TabPosition)
	assert.Equal(t, "basic", metadata.TabsConfig.DefaultActiveTab)
	assert.Len(t, metadata.TabsConfig.Tabs, 2)

	// First tab
	assert.Equal(t, "basic", metadata.TabsConfig.Tabs[0].Key)
	assert.Equal(t, "Basic Settings", metadata.TabsConfig.Tabs[0].Title)
	assert.Equal(t, "settings", metadata.TabsConfig.Tabs[0].Icon)
	assert.Equal(t, []string{"name", "description", "active"}, metadata.TabsConfig.Tabs[0].Fields)

	// Second tab
	assert.Equal(t, "advanced", metadata.TabsConfig.Tabs[1].Key)
	assert.Equal(t, "Advanced", metadata.TabsConfig.Tabs[1].Title)
	assert.Contains(t, metadata.TabsConfig.Tabs[1].Fields, "config.email")

	// Verify object labels
	assert.NotNil(t, metadata.ObjectLabels)
	assert.Equal(t, "Email Settings", metadata.ObjectLabels["config.email"])
	assert.Equal(t, "OAuth Configuration", metadata.ObjectLabels["config.oauth"])

	// Verify properties
	assert.Len(t, metadata.Properties, 2)
	assert.Equal(t, "name", metadata.Properties[0].Path)
	assert.Equal(t, "config", metadata.Properties[1].Path)
}

func TestGenerateJsonConfigMetadataWithGridLayout(t *testing.T) {
	// Create a field with grid layout JSON configuration
	jsonConfig := &JsonConfig{
		DefaultExpanded: true,
		EditorType:      "form",
		Nested:          true,
		RenderAs:        "grid",
		GridConfig: &JsonGridConfig{
			Columns: 12,
			Gutter:  16,
			FieldLayouts: map[string]*JsonFieldLayout{
				"name": {
					Column:  1,
					Row:     1,
					ColSpan: 6,
					RowSpan: 1,
				},
				"description": {
					Column:  7,
					Row:     1,
					ColSpan: 6,
					RowSpan: 1,
				},
				"config.email": {
					Column:  1,
					Row:     2,
					ColSpan: 12,
					RowSpan: 2,
				},
			},
		},
		Properties: []JsonProperty{
			{
				Path:  "name",
				Label: "Name",
				Type:  "string",
			},
			{
				Path:  "description",
				Label: "Description",
				Type:  "string",
			},
			{
				Path:  "config.email",
				Label: "Email Configuration",
				Type:  "object",
			},
		},
	}

	// Generate metadata
	metadata := GenerateJsonConfigMetadata(jsonConfig)

	// Verify base metadata
	assert.NotNil(t, metadata)
	assert.True(t, metadata.Nested)
	assert.Equal(t, "grid", metadata.RenderAs)

	// Verify grid configuration
	assert.NotNil(t, metadata.GridConfig)
	assert.Equal(t, 12, metadata.GridConfig.Columns)
	assert.Equal(t, 16, metadata.GridConfig.Gutter)
	assert.Len(t, metadata.GridConfig.FieldLayouts, 3)

	// Verify field layouts
	nameLayout := metadata.GridConfig.FieldLayouts["name"]
	assert.NotNil(t, nameLayout)
	assert.Equal(t, 1, nameLayout.Column)
	assert.Equal(t, 1, nameLayout.Row)
	assert.Equal(t, 6, nameLayout.ColSpan)
	assert.Equal(t, 1, nameLayout.RowSpan)

	emailLayout := metadata.GridConfig.FieldLayouts["config.email"]
	assert.NotNil(t, emailLayout)
	assert.Equal(t, 1, emailLayout.Column)
	assert.Equal(t, 2, emailLayout.Row)
	assert.Equal(t, 12, emailLayout.ColSpan)
	assert.Equal(t, 2, emailLayout.RowSpan)
}

func TestGenerateFileConfigMetadata(t *testing.T) {
	// Create a file config for testing
	fileConfig := &FileConfig{
		AllowedTypes:       []string{"image/jpeg", "image/png", "application/pdf"},
		MaxSize:            10485760, // 10MB
		BaseURL:            "/uploads/images",
		IsImage:            true,
		MaxWidth:           1920,
		MaxHeight:          1080,
		GenerateThumbnails: true,
		ThumbnailSizes: []ThumbnailSize{
			{
				Name:            "small",
				Width:           150,
				Height:          150,
				KeepAspectRatio: true,
			},
			{
				Name:            "medium",
				Width:           400,
				Height:          300,
				KeepAspectRatio: true,
			},
		},
	}

	// Generate metadata
	metadata := GenerateFileConfigMetadata(fileConfig)

	// Verify metadata
	assert.NotNil(t, metadata)
	assert.Equal(t, []string{"image/jpeg", "image/png", "application/pdf"}, metadata.AllowedTypes)
	assert.Equal(t, int64(10485760), metadata.MaxSize)
	assert.Equal(t, "/uploads/images", metadata.BaseURL)
	assert.True(t, metadata.IsImage)
	assert.Equal(t, 1920, metadata.MaxWidth)
	assert.Equal(t, 1080, metadata.MaxHeight)
	assert.True(t, metadata.GenerateThumbnails)

	// Verify thumbnail sizes
	assert.Len(t, metadata.ThumbnailSizes, 2)
	assert.Equal(t, "small", metadata.ThumbnailSizes[0].Name)
	assert.Equal(t, 150, metadata.ThumbnailSizes[0].Width)
	assert.Equal(t, 150, metadata.ThumbnailSizes[0].Height)
	assert.True(t, metadata.ThumbnailSizes[0].KeepAspectRatio)

	assert.Equal(t, "medium", metadata.ThumbnailSizes[1].Name)
	assert.Equal(t, 400, metadata.ThumbnailSizes[1].Width)
	assert.Equal(t, 300, metadata.ThumbnailSizes[1].Height)
	assert.True(t, metadata.ThumbnailSizes[1].KeepAspectRatio)

	// Test null case
	assert.Nil(t, GenerateFileConfigMetadata(nil))
}

func TestGenerateRichTextConfigMetadata(t *testing.T) {
	// Create a rich text config for testing
	richTextConfig := &RichTextConfig{
		Toolbar: []string{
			"bold", "italic", "underline", "link", "image", "heading",
			"bulletList", "orderedList", "blockquote", "codeBlock",
		},
		Height:       "300px",
		Placeholder:  "Enter your content here...",
		EnableImages: true,
		MaxLength:    10000,
		ShowCounter:  true,
		Format:       "html",
	}

	// Generate metadata
	metadata := GenerateRichTextConfigMetadata(richTextConfig)

	// Verify metadata
	assert.NotNil(t, metadata)
	assert.Contains(t, metadata.Toolbar, "bold")
	assert.Contains(t, metadata.Toolbar, "image")
	assert.Contains(t, metadata.Toolbar, "blockquote")
	assert.Equal(t, "300px", metadata.Height)
	assert.Equal(t, "Enter your content here...", metadata.Placeholder)
	assert.True(t, metadata.EnableImages)
	assert.Equal(t, 10000, metadata.MaxLength)
	assert.True(t, metadata.ShowCounter)
	assert.Equal(t, "html", metadata.Format)

	// Test null case
	assert.Nil(t, GenerateRichTextConfigMetadata(nil))
}

func TestGenerateSelectConfigMetadata(t *testing.T) {
	// Create options and dependent options for testing
	userOptions := []Option{
		{Value: 1, Label: "Admin"},
		{Value: 2, Label: "Editor"},
		{Value: 3, Label: "Viewer"},
	}

	categoryOptions := []Option{
		{Value: "tech", Label: "Technology"},
		{Value: "health", Label: "Health & Wellness"},
		{Value: "finance", Label: "Finance"},
	}

	// Create a select config for testing
	selectConfig := &SelectConfig{
		Multiple:    true,
		Searchable:  true,
		Creatable:   false,
		OptionsURL:  "/api/options/users",
		DependsOn:   "category",
		Placeholder: "Select user roles",
		Clearable:   true,
		DisplayMode: "tags",
		DependentOptions: map[string][]Option{
			"admin": userOptions,
			"editor": {
				{Value: 2, Label: "Editor"},
				{Value: 3, Label: "Viewer"},
			},
			"category": categoryOptions,
		},
	}

	// Generate metadata
	metadata := GenerateSelectConfigMetadata(selectConfig)

	// Verify metadata
	assert.NotNil(t, metadata)
	assert.True(t, metadata.Multiple)
	assert.True(t, metadata.Searchable)
	assert.False(t, metadata.Creatable)
	assert.Equal(t, "/api/options/users", metadata.OptionsURL)
	assert.Equal(t, "category", metadata.DependsOn)
	assert.Equal(t, "Select user roles", metadata.Placeholder)
	assert.True(t, metadata.Clearable)
	assert.Equal(t, "tags", metadata.DisplayMode)

	// Verify dependent options
	assert.Len(t, metadata.DependentOptions, 3)
	assert.Len(t, metadata.DependentOptions["admin"], 3)
	assert.Len(t, metadata.DependentOptions["editor"], 2)
	assert.Len(t, metadata.DependentOptions["category"], 3)

	// Check specific option content
	adminValue := metadata.DependentOptions["admin"][0].Value
	assert.Equal(t, 1, adminValue)
	assert.Equal(t, "Admin", metadata.DependentOptions["admin"][0].Label)
	assert.Equal(t, "tech", metadata.DependentOptions["category"][0].Value)
	assert.Equal(t, "Technology", metadata.DependentOptions["category"][0].Label)

	// Test null case
	assert.Nil(t, GenerateSelectConfigMetadata(nil))
}

func TestGenerateComputedFieldConfigMetadata(t *testing.T) {
	// Create a computed field config for testing
	computedConfig := &ComputedFieldConfig{
		DependsOn:    []string{"firstName", "lastName"},
		Expression:   "${firstName} + ' ' + ${lastName}",
		ClientSide:   true,
		Format:       "text",
		Persist:      false,
		ComputeOrder: 10,
	}

	// Generate metadata
	metadata := GenerateComputedFieldConfigMetadata(computedConfig)

	// Verify metadata
	assert.NotNil(t, metadata)
	assert.Equal(t, []string{"firstName", "lastName"}, metadata.DependsOn)
	assert.Equal(t, "${firstName} + ' ' + ${lastName}", metadata.Expression)
	assert.True(t, metadata.ClientSide)
	assert.Equal(t, "text", metadata.Format)
	assert.False(t, metadata.Persist)
	assert.Equal(t, 10, metadata.ComputeOrder)

	// Test null case
	assert.Nil(t, GenerateComputedFieldConfigMetadata(nil))
}

func TestGenerateFieldsMetadataWithSpecialFields(t *testing.T) {
	// Create sample fields with special field types
	fields := []Field{
		{
			Name:  "avatar",
			Type:  "file",
			Label: "User Avatar",
			File: &FileConfig{
				IsImage:      true,
				AllowedTypes: []string{"image/jpeg", "image/png"},
				MaxSize:      2097152, // 2MB
			},
		},
		{
			Name:  "biography",
			Type:  "richtext",
			Label: "User Biography",
			RichText: &RichTextConfig{
				Toolbar:     []string{"bold", "italic", "link"},
				MaxLength:   5000,
				ShowCounter: true,
			},
		},
		{
			Name:  "role",
			Type:  "select",
			Label: "User Role",
			Select: &SelectConfig{
				Multiple:    false,
				Searchable:  true,
				DisplayMode: "dropdown",
			},
		},
		{
			Name:     "fullName",
			Type:     "string",
			Label:    "Full Name",
			ReadOnly: true,
			Computed: &ComputedFieldConfig{
				DependsOn:  []string{"firstName", "lastName"},
				Expression: "${firstName} + ' ' + ${lastName}",
				ClientSide: true,
			},
		},
	}

	// Generate metadata
	metadata := GenerateFieldsMetadata(fields)

	// Verify metadata
	assert.Len(t, metadata, 4)

	// Check avatar field (file type)
	avatarMeta := findFieldMetadata(metadata, "avatar")
	assert.NotNil(t, avatarMeta)
	assert.Equal(t, "file", avatarMeta.Type)
	assert.NotNil(t, avatarMeta.File)
	assert.True(t, avatarMeta.File.IsImage)
	assert.Equal(t, []string{"image/jpeg", "image/png"}, avatarMeta.File.AllowedTypes)
	assert.Equal(t, int64(2097152), avatarMeta.File.MaxSize)

	// Check biography field (richtext type)
	bioMeta := findFieldMetadata(metadata, "biography")
	assert.NotNil(t, bioMeta)
	assert.Equal(t, "richtext", bioMeta.Type)
	assert.NotNil(t, bioMeta.RichText)
	assert.Contains(t, bioMeta.RichText.Toolbar, "bold")
	assert.Equal(t, 5000, bioMeta.RichText.MaxLength)
	assert.True(t, bioMeta.RichText.ShowCounter)

	// Check role field (select type)
	roleMeta := findFieldMetadata(metadata, "role")
	assert.NotNil(t, roleMeta)
	assert.Equal(t, "select", roleMeta.Type)
	assert.NotNil(t, roleMeta.Select)
	assert.False(t, roleMeta.Select.Multiple)
	assert.True(t, roleMeta.Select.Searchable)
	assert.Equal(t, "dropdown", roleMeta.Select.DisplayMode)

	// Check fullName field (computed type)
	nameMeta := findFieldMetadata(metadata, "fullName")
	assert.NotNil(t, nameMeta)
	assert.Equal(t, "string", nameMeta.Type)
	assert.True(t, nameMeta.ReadOnly)
	assert.NotNil(t, nameMeta.Computed)
	assert.Equal(t, []string{"firstName", "lastName"}, nameMeta.Computed.DependsOn)
	assert.Equal(t, "${firstName} + ' ' + ${lastName}", nameMeta.Computed.Expression)
	assert.True(t, nameMeta.Computed.ClientSide)
}

// Helper function to find field metadata by name
func findFieldMetadata(fields []FieldMetadata, name string) *FieldMetadata {
	for _, field := range fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

func TestGenerateAntDesignConfigMetadata(t *testing.T) {
	// Create an Ant Design config for testing
	antDesignConfig := &AntDesignConfig{
		ComponentType: "Select",
		Props: map[string]interface{}{
			"allowClear":  true,
			"mode":        "multiple",
			"placeholder": "Select options",
		},
		Rules: []AntDesignRule{
			{
				Type:    "required",
				Message: "This field is required",
			},
			{
				Type:    "min",
				Value:   2,
				Message: "Please select at least 2 options",
			},
		},
		FormItemProps: map[string]interface{}{
			"tooltip": "Select multiple options",
		},
		Dependencies: []string{"category", "type"},
	}

	// Generate metadata
	metadata := GenerateAntDesignConfigMetadata(antDesignConfig)

	// Verify metadata
	assert.NotNil(t, metadata)
	assert.Equal(t, "Select", metadata.ComponentType)

	// Verify props
	assert.NotNil(t, metadata.Props)
	assert.Equal(t, true, metadata.Props["allowClear"])
	assert.Equal(t, "multiple", metadata.Props["mode"])
	assert.Equal(t, "Select options", metadata.Props["placeholder"])

	// Verify rules
	assert.Len(t, metadata.Rules, 2)
	assert.Equal(t, "required", metadata.Rules[0].Type)
	assert.Equal(t, "This field is required", metadata.Rules[0].Message)
	assert.Equal(t, "min", metadata.Rules[1].Type)
	assert.Equal(t, 2, metadata.Rules[1].Value)
	assert.Equal(t, "Please select at least 2 options", metadata.Rules[1].Message)

	// Verify form item props
	assert.NotNil(t, metadata.FormItemProps)
	assert.Equal(t, "Select multiple options", metadata.FormItemProps["tooltip"])

	// Verify dependencies
	assert.Equal(t, []string{"category", "type"}, metadata.Dependencies)

	// Test null case
	assert.Nil(t, GenerateAntDesignConfigMetadata(nil))
}

func TestMapValidationToAntDesignRules(t *testing.T) {
	// Create validation rules
	validation := &Validation{
		Required:  true,
		Min:       10,
		Max:       100,
		MinLength: 5,
		MaxLength: 50,
		Pattern:   "^[a-zA-Z0-9]+$",
		Message:   "Custom validation message",
	}

	// Map to Ant Design rules
	rules := MapValidationToAntDesignRules(validation)

	// Verify rules
	assert.Len(t, rules, 6)

	// Find and verify each rule type
	var requiredRule, minLengthRule, maxLengthRule, patternRule, minRule, maxRule *AntDesignRule

	for i := range rules {
		rule := &rules[i]
		switch rule.Type {
		case "required":
			requiredRule = rule
		case "min":
			if rule.Value == validation.MinLength {
				minLengthRule = rule
			} else if rule.Value == validation.Min {
				minRule = rule
			}
		case "max":
			if rule.Value == validation.MaxLength {
				maxLengthRule = rule
			} else if rule.Value == validation.Max {
				maxRule = rule
			}
		case "pattern":
			patternRule = rule
		}
	}

	// Verify required rule
	assert.NotNil(t, requiredRule)
	assert.Equal(t, "required", requiredRule.Type)
	assert.Equal(t, validation.Message, requiredRule.Message)

	// Verify min length rule
	assert.NotNil(t, minLengthRule)
	assert.Equal(t, "min", minLengthRule.Type)
	assert.Equal(t, validation.MinLength, minLengthRule.Value)

	// Verify max length rule
	assert.NotNil(t, maxLengthRule)
	assert.Equal(t, "max", maxLengthRule.Type)
	assert.Equal(t, validation.MaxLength, maxLengthRule.Value)

	// Verify pattern rule
	assert.NotNil(t, patternRule)
	assert.Equal(t, "pattern", patternRule.Type)
	assert.Equal(t, validation.Pattern, patternRule.Pattern)

	// Verify min value rule
	assert.NotNil(t, minRule)
	assert.Equal(t, "min", minRule.Type)
	assert.Equal(t, validation.Min, minRule.Value)

	// Verify max value rule
	assert.NotNil(t, maxRule)
	assert.Equal(t, "max", maxRule.Type)
	assert.Equal(t, validation.Max, maxRule.Value)

	// Test null case
	assert.Nil(t, MapValidationToAntDesignRules(nil))
}

func TestAutoDetectAntDesignComponent(t *testing.T) {
	// Test detection for various field types
	testCases := []struct {
		name     string
		field    Field
		expected string
	}{
		{
			name: "String field",
			field: Field{
				Name: "title",
				Type: "string",
			},
			expected: "Input",
		},
		{
			name: "Password field",
			field: Field{
				Name: "password",
				Type: "string",
			},
			expected: "Password",
		},
		{
			name: "Number field",
			field: Field{
				Name: "age",
				Type: "number",
			},
			expected: "InputNumber",
		},
		{
			name: "Boolean field",
			field: Field{
				Name: "active",
				Type: "boolean",
			},
			expected: "Switch",
		},
		{
			name: "Date field",
			field: Field{
				Name: "birthDate",
				Type: "date",
			},
			expected: "DatePicker",
		},
		{
			name: "Select field with options",
			field: Field{
				Name: "category",
				Type: "string",
				Options: []Option{
					{Value: "a", Label: "A"},
					{Value: "b", Label: "B"},
				},
			},
			expected: "Select",
		},
		{
			name: "File field",
			field: Field{
				Name: "document",
				Type: "file",
				File: &FileConfig{
					IsImage: false,
				},
			},
			expected: "Upload",
		},
		{
			name: "Image field",
			field: Field{
				Name: "avatar",
				Type: "file",
				File: &FileConfig{
					IsImage: true,
				},
			},
			expected: "Upload.Image",
		},
		{
			name: "Rich text field",
			field: Field{
				Name:     "content",
				Type:     "string",
				RichText: &RichTextConfig{},
			},
			expected: "TextArea",
		},
		{
			name: "JSON field",
			field: Field{
				Name: "config",
				Type: "json",
			},
			expected: "JsonEditor",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			component := AutoDetectAntDesignComponent(&tc.field)
			assert.Equal(t, tc.expected, component)
		})
	}

	// Test null case
	assert.Equal(t, "Input", AutoDetectAntDesignComponent(nil))
}

func TestGenerateFieldsMetadataWithAntDesign(t *testing.T) {
	// Create fields with explicit and implicit Ant Design configuration
	fields := []Field{
		{
			Name: "name",
			Type: "string",
			Validation: &Validation{
				Required:  true,
				MinLength: 3,
				MaxLength: 50,
			},
			Form: &FormConfig{
				Placeholder: "Enter your name",
			},
			// Explicit Ant Design config
			AntDesign: &AntDesignConfig{
				ComponentType: "Input",
				Props: map[string]interface{}{
					"allowClear": true,
				},
			},
		},
		{
			Name: "age",
			Type: "number",
			Validation: &Validation{
				Required: true,
				Min:      18,
				Max:      120,
			},
			// No explicit Ant Design config - should be auto-generated
		},
		{
			Name: "isActive",
			Type: "boolean",
			// No explicit Ant Design config - should be auto-generated
		},
		{
			Name: "role",
			Type: "select",
			Options: []Option{
				{Value: "admin", Label: "Administrator"},
				{Value: "user", Label: "Regular User"},
			},
			// No explicit Ant Design config - should be auto-generated
		},
	}

	// Generate metadata
	metadata := GenerateFieldsMetadata(fields)

	// Verify metadata
	assert.Len(t, metadata, 4)

	// Check name field with explicit config
	nameMeta := findFieldMetadata(metadata, "name")
	assert.NotNil(t, nameMeta)
	assert.NotNil(t, nameMeta.AntDesign)
	assert.Equal(t, "Input", nameMeta.AntDesign.ComponentType)
	assert.Equal(t, true, nameMeta.AntDesign.Props["allowClear"])

	// Check age field with auto-generated config
	ageMeta := findFieldMetadata(metadata, "age")
	assert.NotNil(t, ageMeta)
	assert.NotNil(t, ageMeta.AntDesign)
	assert.Equal(t, "InputNumber", ageMeta.AntDesign.ComponentType)
	assert.Equal(t, float64(18), ageMeta.AntDesign.Props["min"])
	assert.Equal(t, float64(120), ageMeta.AntDesign.Props["max"])

	// Check rules for age field
	assert.NotEmpty(t, ageMeta.AntDesign.Rules)
	hasRequiredRule := false
	for _, rule := range ageMeta.AntDesign.Rules {
		if rule.Type == "required" {
			hasRequiredRule = true
			break
		}
	}
	assert.True(t, hasRequiredRule)

	// Check isActive field
	isActiveMeta := findFieldMetadata(metadata, "isActive")
	assert.NotNil(t, isActiveMeta)
	assert.NotNil(t, isActiveMeta.AntDesign)
	assert.Equal(t, "Switch", isActiveMeta.AntDesign.ComponentType)
	assert.Equal(t, "checked", isActiveMeta.AntDesign.FormItemProps["valuePropName"])

	// Check role field
	roleMeta := findFieldMetadata(metadata, "role")
	assert.NotNil(t, roleMeta)
	assert.NotNil(t, roleMeta.AntDesign)
	assert.Equal(t, "Select", roleMeta.AntDesign.ComponentType)

	// Verify options in Select
	options, ok := roleMeta.AntDesign.Props["options"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, options, 2)
	assert.Equal(t, "admin", options[0]["value"])
	assert.Equal(t, "Administrator", options[0]["label"])
	assert.Equal(t, "user", options[1]["value"])
	assert.Equal(t, "Regular User", options[1]["label"])
}
