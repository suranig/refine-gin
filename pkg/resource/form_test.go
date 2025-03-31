package resource

import (
	"testing"
)

func TestValidateFormLayout(t *testing.T) {
	// Test cases
	tests := []struct {
		name    string
		layout  *FormLayout
		wantErr bool
	}{
		{
			name: "Valid form layout",
			layout: &FormLayout{
				Columns: 2,
				Gutter:  16,
				Sections: []*FormSection{
					{
						ID:    "section1",
						Title: "Personal Information",
					},
				},
				FieldLayouts: []*FormFieldLayout{
					{
						Field:     "name",
						SectionID: "section1",
						Column:    0,
						Row:       0,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil layout",
			layout:  nil,
			wantErr: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormLayout(tt.layout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormLayout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFormDependencies(t *testing.T) {
	// Create test fields
	fields := []Field{
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "address", Type: "string"},
	}

	// Test cases
	tests := []struct {
		name    string
		layout  *FormLayout
		fields  []Field
		wantErr bool
	}{
		{
			name: "Valid dependencies",
			layout: &FormLayout{
				Sections: []*FormSection{
					{
						ID: "section1",
					},
				},
				FieldLayouts: []*FormFieldLayout{
					{
						Field:     "name",
						SectionID: "section1",
					},
				},
			},
			fields:  fields,
			wantErr: false,
		},
		{
			name: "Invalid field reference",
			layout: &FormLayout{
				FieldLayouts: []*FormFieldLayout{
					{
						Field: "nonexistent",
					},
				},
			},
			fields:  fields,
			wantErr: true,
		},
		{
			name: "Invalid section reference",
			layout: &FormLayout{
				FieldLayouts: []*FormFieldLayout{
					{
						Field:     "name",
						SectionID: "nonexistent",
					},
				},
			},
			fields:  fields,
			wantErr: true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormDependencies(tt.layout, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateDefaultFormLayout(t *testing.T) {
	// Create test fields
	fields := []Field{
		{Name: "ID", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "address", Type: "string", Hidden: true},
		{Name: "createdAt", Type: "time"},
	}

	// Generate default layout
	layout := GenerateDefaultFormLayout(fields)

	// Verify layout
	if layout == nil {
		t.Fatal("GenerateDefaultFormLayout() returned nil")
	}

	// Check basic properties
	if layout.Columns != 1 {
		t.Errorf("Expected Columns = 1, got %d", layout.Columns)
	}

	// Should have one section
	if len(layout.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(layout.Sections))
	}

	// Check field layouts - should have 3 fields (ID and hidden fields excluded)
	expectedFieldCount := 3 // name, email, createdAt
	if len(layout.FieldLayouts) != expectedFieldCount {
		t.Errorf("Expected %d field layouts, got %d", expectedFieldCount, len(layout.FieldLayouts))
	}

	// Check if ID field is excluded
	for _, fl := range layout.FieldLayouts {
		if fl.Field == "ID" || fl.Field == "id" {
			t.Error("ID field should be excluded from field layouts")
		}
		if fl.Field == "address" {
			t.Error("Hidden field 'address' should be excluded from field layouts")
		}
	}
}

func TestGenerateFormLayoutMetadata(t *testing.T) {
	// Test case
	layout := &FormLayout{
		Columns: 2,
		Gutter:  16,
		Sections: []*FormSection{
			{
				ID:    "personal",
				Title: "Personal Info",
				Condition: &FormCondition{
					Field:    "showPersonal",
					Operator: "eq",
					Value:    true,
				},
			},
		},
		FieldLayouts: []*FormFieldLayout{
			{
				Field:     "name",
				SectionID: "personal",
				Column:    0,
				Row:       0,
				ColSpan:   1,
			},
		},
	}

	// Generate metadata
	metadata := GenerateFormLayoutMetadata(layout)

	// Verify metadata
	if metadata == nil {
		t.Fatal("GenerateFormLayoutMetadata() returned nil")
	}

	// Check basic properties
	if metadata.Columns != 2 {
		t.Errorf("Expected Columns = 2, got %d", metadata.Columns)
	}
	if metadata.Gutter != 16 {
		t.Errorf("Expected Gutter = 16, got %d", metadata.Gutter)
	}

	// Check sections
	if len(metadata.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(metadata.Sections))
	}
	if metadata.Sections[0].ID != "personal" {
		t.Errorf("Expected section ID = \"personal\", got %s", metadata.Sections[0].ID)
	}
	if metadata.Sections[0].Title != "Personal Info" {
		t.Errorf("Expected section Title = \"Personal Info\", got %s", metadata.Sections[0].Title)
	}
	if metadata.Sections[0].Condition == nil {
		t.Error("Expected section to have a condition, got nil")
	} else {
		if metadata.Sections[0].Condition.Field != "showPersonal" {
			t.Errorf("Expected condition field = \"showPersonal\", got %s",
				metadata.Sections[0].Condition.Field)
		}
		if metadata.Sections[0].Condition.Operator != "eq" {
			t.Errorf("Expected condition operator = \"eq\", got %s",
				metadata.Sections[0].Condition.Operator)
		}
	}

	// Check field layouts
	if len(metadata.FieldLayouts) != 1 {
		t.Errorf("Expected 1 field layout, got %d", len(metadata.FieldLayouts))
	}
	if metadata.FieldLayouts[0].Field != "name" {
		t.Errorf("Expected field = \"name\", got %s", metadata.FieldLayouts[0].Field)
	}
	if metadata.FieldLayouts[0].SectionID != "personal" {
		t.Errorf("Expected sectionID = \"personal\", got %s", metadata.FieldLayouts[0].SectionID)
	}
}
