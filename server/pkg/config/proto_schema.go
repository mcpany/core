package config

import (
	"encoding/json"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type schemaGenerator struct {
	defs map[string]interface{}
}

// GenerateSchemaFromProto generates a jsonschema from a protobuf message using reflection.
//
// Parameters:
//   msg: The protobuf message to generate the schema from.
//
// Returns:
//   *jsonschema.Schema: The generated JSON schema.
//   error: An error if the schema generation fails.
func GenerateSchemaFromProto(msg protoreflect.Message) (*jsonschema.Schema, error) {
	schemaMap := GenerateSchemaMapFromProto(msg)
	return CompileSchema(schemaMap)
}

// GenerateSchemaMapFromProto generates a raw JSON schema map from a protobuf message using reflection.
// This is useful if you want to export the schema as JSON.
//
// Parameters:
//   msg: The protobuf message to generate the schema from.
//
// Returns:
//   map[string]interface{}: The generated JSON schema map.
func GenerateSchemaMapFromProto(msg protoreflect.Message) map[string]interface{} {
	gen := &schemaGenerator{
		defs: make(map[string]interface{}),
	}

	rootRef := gen.getOrAddDefinition(msg.Descriptor())

	return map[string]interface{}{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$defs":   gen.defs,
		"$ref":    rootRef["$ref"],
	}
}

// CompileSchema compiles a raw JSON schema map into a jsonschema.Schema object.
//
// schemaMap is the schemaMap.
//
// Returns the result.
// Returns an error if the operation fails.
func CompileSchema(schemaMap map[string]interface{}) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	url := "config.schema.json"

	b, err := json.Marshal(schemaMap)
	if err != nil {
		return nil, err
	}

	if err := compiler.AddResource(url, strings.NewReader(string(b))); err != nil {
		return nil, err
	}
	return compiler.Compile(url)
}

func (g *schemaGenerator) getOrAddDefinition(
	desc protoreflect.MessageDescriptor,
) map[string]interface{} {
	fullName := string(desc.FullName())
	ref := map[string]interface{}{
		"$ref": "#/$defs/" + fullName,
	}

	if _, exists := g.defs[fullName]; exists {
		return ref
	}

	// Create placeholder to prevent infinite recursion
	g.defs[fullName] = map[string]interface{}{}

	schema := g.protoMessageToSchema(desc)
	g.defs[fullName] = schema

	return ref
}

func (g *schemaGenerator) protoMessageToSchema(
	desc protoreflect.MessageDescriptor,
) map[string]interface{} {
	if desc.FullName() == "google.protobuf.Duration" {
		return map[string]interface{}{
			"type":    "string",
			"pattern": "^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$",
		}
	}
	if desc.FullName() == "google.protobuf.Struct" {
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": true,
		}
	}

	properties := make(map[string]interface{})
	fields := desc.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := field.JSONName()
		if fieldName == "" {
			fieldName = string(field.Name())
		}

		properties[fieldName] = g.protoFieldToSchema(field)
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		// strict validation? additionalProperties: false
		// For now let's be strict to prove it works.
		"additionalProperties": false,
	}
}

func (g *schemaGenerator) protoFieldToSchema(
	field protoreflect.FieldDescriptor,
) map[string]interface{} {
	if field.IsMap() {
		valSchema := g.protoFieldTypeToSchema(field.MapValue())
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": valSchema,
		}
	}

	if field.IsList() {
		return map[string]interface{}{
			"type":  "array",
			"items": g.protoFieldTypeToSchema(field),
		}
	}

	return g.protoFieldTypeToSchema(field)
}

func (g *schemaGenerator) protoFieldTypeToSchema(
	field protoreflect.FieldDescriptor,
) map[string]interface{} {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return g.getOrAddDefinition(field.Message())
	case protoreflect.StringKind:
		return map[string]interface{}{"type": "string"}
	case protoreflect.BoolKind:
		return map[string]interface{}{"type": "boolean"}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		return map[string]interface{}{"type": "integer"}
	case protoreflect.Int64Kind, protoreflect.Uint64Kind,
		protoreflect.Sint64Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed64Kind:
		return map[string]interface{}{"type": "string"} // ProtoJSON: 64-bit ints are strings
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return map[string]interface{}{"type": "number"}
	case protoreflect.BytesKind:
		return map[string]interface{}{"type": "string", "contentEncoding": "base64"}
	case protoreflect.EnumKind:
		return map[string]interface{}{"type": "string"} // Simplified
	}
	return map[string]interface{}{}
}
