// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package transformer provides functionality for transforming and parsing data.
package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/itchyny/gojq"
	"k8s.io/client-go/util/jsonpath"
)

var (
	jsonPathCache sync.Map // map[string]*jsonpath.JSONPath
	xpathCache    sync.Map // map[string]*xpath.Expr
	regexCache    sync.Map // map[string]*regexp.Regexp
	jqCache       sync.Map // map[string]*gojq.Query
)

// TextParser provides functionality to parse various text formats (JSON, XML,
// plain text) and extract data into a structured map. It uses a configuration
// map to define the extraction rules for each format, such as JSONPath for
// JSON, XPath for XML, and regex for plain text.
type TextParser struct {
	transformer *Transformer
}

var (
	defaultTextParser     *TextParser
	defaultTextParserOnce sync.Once
)

// NewTextParser returns a shared instance of TextParser.
//
// Returns the result.
func NewTextParser() *TextParser {
	defaultTextParserOnce.Do(func() {
		defaultTextParser = &TextParser{
			transformer: NewTransformer(),
		}
	})
	return defaultTextParser
}

// Transform takes a map of data and a Go template string and returns a byte
// slice containing the transformed output.
//
// templateStr is the Go template to be executed.
// data is the map containing the data to be used in the template.
// It returns the transformed data as a byte slice or an error if the
// transformation fails.
func (p *TextParser) Transform(templateStr string, data any) ([]byte, error) {
	return p.transformer.Transform(templateStr, data)
}

// Parse extracts data from an input byte slice based on the specified input
// type and configuration.
//
// inputType specifies the format of the input data ("json", "xml", "text", or "jq").
// input is the raw byte slice containing the data to be parsed.
// config is a map where keys are the desired output keys and values are the
// extraction rules (JSONPath, XPath, or regex) for the corresponding data.
// jqQuery is the JQ query string (only used when inputType is "jq").
// It returns the extracted data (as a map or any for JQ) or an error if parsing fails.
func (p *TextParser) Parse(inputType string, input []byte, config map[string]string, jqQuery string) (any, error) {
	switch strings.ToLower(inputType) {
	case "json":
		return p.parseJSON(input, config)
	case "xml":
		return p.parseXML(input, config)
	case "text":
		return p.parseText(input, config)
	case "jq":
		return p.parseJQ(input, jqQuery)
	default:
		return nil, fmt.Errorf("unsupported input type: %s", inputType)
	}
}

// parseJSON handles the parsing of JSON data. It uses JSONPath expressions from
// the config map to extract values from the input JSON.
func (p *TextParser) parseJSON(input []byte, config map[string]string) (any, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		var j *jsonpath.JSONPath
		if val, ok := jsonPathCache.Load(path); ok {
			j = val.(*jsonpath.JSONPath)
		} else {
			j = jsonpath.New(key)
			if err := j.Parse(path); err != nil {
				return nil, fmt.Errorf("failed to parse JSONPath for key '%s': %w", key, err)
			}
			jsonPathCache.Store(path, j)
		}

		values, err := j.FindResults(data)
		if err != nil && !strings.Contains(err.Error(), "is not found") {
			return nil, fmt.Errorf("failed to find results for JSONPath '%s': %w", path, err)
		}

		if len(values) > 0 && len(values[0]) > 0 {
			result[key] = values[0][0].Interface()
		}
	}
	return result, nil
}

// parseXML handles the parsing of XML data. It uses XPath expressions from the
// config map to query and extract values from the input XML.
func (p *TextParser) parseXML(input []byte, config map[string]string) (any, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		var expr *xpath.Expr
		if val, ok := xpathCache.Load(path); ok {
			expr = val.(*xpath.Expr)
		} else {
			var err error
			expr, err = xpath.Compile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to compile xpath for key '%s': %w", key, err)
			}
			xpathCache.Store(path, expr)
		}

		node := xmlquery.QuerySelector(doc, expr)
		if node != nil {
			result[key] = node.InnerText()
		}
	}
	return result, nil
}

// parseText handles the parsing of plain text data. It uses regular expressions
// from the config map to find and extract substrings from the input text.
func (p *TextParser) parseText(input []byte, config map[string]string) (any, error) {
	result := make(map[string]any)
	inputText := string(input)

	for key, pattern := range config {
		var re *regexp.Regexp
		if val, ok := regexCache.Load(pattern); ok {
			re = val.(*regexp.Regexp)
		} else {
			var err error
			re, err = regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex for key '%s': %w", key, err)
			}
			regexCache.Store(pattern, re)
		}

		matches := re.FindStringSubmatch(inputText)
		if len(matches) > 1 {
			result[key] = matches[1] // extract first capture group
		}
	}
	return result, nil
}

// parseJQ handles the parsing of JSON data using JQ queries.
func (p *TextParser) parseJQ(input []byte, query string) (any, error) {
	if query == "" {
		return nil, fmt.Errorf("jq query cannot be empty")
	}

	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON input for JQ: %w", err)
	}

	var pq *gojq.Query
	if val, ok := jqCache.Load(query); ok {
		pq = val.(*gojq.Query)
	} else {
		var err error
		pq, err = gojq.Parse(query)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JQ query: %w", err)
		}
		jqCache.Store(query, pq)
	}

	iter := pq.Run(data)
	var results []any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return nil, fmt.Errorf("jq execution failed: %w", err)
		}
		results = append(results, v)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
