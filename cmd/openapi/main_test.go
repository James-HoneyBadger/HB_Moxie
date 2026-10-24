package main

import (
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestApplyJSONStringOptions_UsesStringForStringEncodedIntegers(t *testing.T) {
	reflector := &jsonschema.Reflector{
		FieldNameTag: "json",
		Anonymous:    true,
	}

	schema := reflector.Reflect(&apiSchemaTypes{})
	applyJSONStringOptions(schema.Definitions)

	definition := schema.Definitions["BackupListResponse"]
	assert.NotNil(t, definition)

	fileName, exists := definition.Properties.Get("fileSize")
	assert.True(t, exists)
	assert.Equal(t, "string", fileName.Type)
}
