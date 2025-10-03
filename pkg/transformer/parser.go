/*
 * Copyright 2025 Author(s) of MCPX
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"k8s.io/client-go/util/jsonpath"
)

// TextParser provides functionality to parse various text formats (JSON, XML,
// plain text) and extract data into a structured map. It uses a configuration
// map to define the extraction rules for each format, such as JSONPath for
// JSON, XPath for XML, and regex for plain text.
type TextParser struct{}

// NewTextParser creates and returns a new instance of TextParser.
func NewTextParser() *TextParser {
	return &TextParser{}
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

func (p *TextParser) parseJSON(input []byte, config map[string]string) (map[string]any, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		j := jsonpath.New(key)
		if err := j.Parse(path); err != nil {
			return nil, fmt.Errorf("failed to parse JSONPath for key '%s': %w", key, err)
		}
		values, err := j.FindResults(data)
		if err != nil {
			return nil, fmt.Errorf("failed to find results for JSONPath '%s': %w", path, err)
		}

		if len(values) > 0 && len(values[0]) > 0 {
			result[key] = values[0][0].Interface()
		}
	}
	return result, nil
}

func (p *TextParser) parseXML(input []byte, config map[string]string) (map[string]any, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	result := make(map[string]any)
	for key, path := range config {
		expr, err := xpath.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to compile xpath for key '%s': %w", key, err)
		}
		node := xmlquery.QuerySelector(doc, expr)
		if node != nil {
			result[key] = node.InnerText()
		}
	}
	return result, nil
}

func (p *TextParser) parseText(input []byte, config map[string]string) (map[string]any, error) {
	result := make(map[string]any)
	inputText := string(input)

	for key, pattern := range config {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex for key '%s': %w", key, err)
		}

		matches := re.FindStringSubmatch(inputText)
		if len(matches) > 1 {
			result[key] = matches[1] // extract first capture group
		}
	}
	return result, nil
}
