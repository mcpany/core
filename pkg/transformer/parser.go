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
	"k8s.io/client-go/util/jsonpath"
)

// TextParser provides functionality to parse various text formats (JSON, XML,
// plain text) and extract data into a structured map. It uses a configuration
// map to define the extraction rules for each format, such as JSONPath for
// JSON, XPath for XML, and regex for plain text.
type TextParser struct {
	transformer   *Transformer
	regexCache    sync.Map
	xpathCache    sync.Map
	jsonPathCache sync.Map
}

// NewTextParser creates and returns a new instance of TextParser.
func NewTextParser() *TextParser {
	return &TextParser{
		transformer: NewTransformer(),
	}
}

// Transform takes a map of data and a Go template string and returns a byte
// slice containing the transformed output.
//
// templateStr is the Go template to be executed.
// data is the map containing the data to be used in the template.
// It returns the transformed data as a byte slice or an error if the
// transformation fails.
func (p *TextParser) Transform(templateStr string, data map[string]any) ([]byte, error) {
	return p.transformer.Transform(templateStr, data)
}

// Parse extracts data from an input byte slice based on the specified input
// type and configuration.
//
// inputType specifies the format of the input data ("json", "xml", or "text").
// input is the raw byte slice containing the data to be parsed.
// config is a map where keys are the desired output keys and values are the
// extraction rules (JSONPath, XPath, or regex) for the corresponding data.
// It returns a map containing the extracted data or an error if parsing fails.
func (p *TextParser) Parse(inputType string, input []byte, config map[string]string) (map[string]any, error) {
	switch strings.ToLower(inputType) {
	case "json":
		return p.parseJSON(input, config)
	case "xml":
		return p.parseXML(input, config)
	case "text":
		return p.parseText(input, config)
	default:
		return nil, fmt.Errorf("unsupported input type: %s", inputType)
	}
}

// parseJSON handles the parsing of JSON data. It uses JSONPath expressions from
// the config map to extract values from the input JSON.
func (p *TextParser) parseJSON(input []byte, config map[string]string) (map[string]any, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		var j *jsonpath.JSONPath
		if val, ok := p.jsonPathCache.Load(path); ok {
			j = val.(*jsonpath.JSONPath)
		} else {
			j = jsonpath.New(key)
			if err := j.Parse(path); err != nil {
				return nil, fmt.Errorf("failed to parse JSONPath for key '%s': %w", key, err)
			}
			p.jsonPathCache.Store(path, j)
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
func (p *TextParser) parseXML(input []byte, config map[string]string) (map[string]any, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		var expr *xpath.Expr
		if val, ok := p.xpathCache.Load(path); ok {
			expr = val.(*xpath.Expr)
		} else {
			var err error
			expr, err = xpath.Compile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to compile xpath for key '%s': %w", key, err)
			}
			p.xpathCache.Store(path, expr)
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
func (p *TextParser) parseText(input []byte, config map[string]string) (map[string]any, error) {
	result := make(map[string]any)
	inputText := string(input)

	for key, pattern := range config {
		var re *regexp.Regexp
		if val, ok := p.regexCache.Load(pattern); ok {
			re = val.(*regexp.Regexp)
		} else {
			var err error
			re, err = regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex for key '%s': %w", key, err)
			}
			p.regexCache.Store(pattern, re)
		}

		matches := re.FindStringSubmatch(inputText)
		if len(matches) > 1 {
			result[key] = matches[1] // extract first capture group
		}
	}
	return result, nil
}
