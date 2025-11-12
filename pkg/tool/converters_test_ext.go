
package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertJSONSchemaToStruct_InvalidType(t *testing.T) {
	jsonSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("invalid-type"),
		},
	}
	_, err := convertJSONSchemaToStruct(jsonSchema)
	assert.Error(t, err, "Should return an error for an invalid schema type")
}
