package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("server/tests/integration/e2e_helpers.go")
	if err != nil {
		panic(err)
	}

	str := string(content)

	// Fix the first lint error (nilerr)
	str = strings.Replace(str, `func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	return os.Symlink(src, dst)
}`, `func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		// Ignore if the source directory does not exist.
		return nil //nolint:nilerr
	}
	return os.Symlink(src, dst)
}`, 1)

	str = strings.Replace(str, `	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
		Parameters:   params,
	}.Build()
	toolDef.SetCallId(callID)`, `	callID := "call-" + toolDef.GetName()
	_ = configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
		Parameters:   params,
	}.Build()
	toolDef.SetCallId(callID)`, 1)

	str = strings.Replace(str, `	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()
	callID := "call-" + toolDef.GetName()
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()
	toolDef.SetCallId(callID)`, `	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()
	callID := "call-" + toolDef.GetName()
	_ = configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()
	toolDef.SetCallId(callID)`, 1)

	err = os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Patched successfully")
}
