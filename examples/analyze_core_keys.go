// Package main provides analysis utilities for examining SlimAcademy document structure.
// This tool analyzes JSON keys and content patterns to help understand document schemas
// and support development and debugging workflows.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
)

type KeyInfo struct {
	Count      int
	Percentage float64
	DataTypes  []string
	IsCore     bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_core_keys.go <json-file>")
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	analyze(filename, jsonData)
	return nil
}

func analyze(filename string, jsonData any) {
	keyStats := make(map[string]*KeyInfo)
	analyzeJSON(jsonData, keyStats)

	totalObjects := countObjects(jsonData)

	// Calculate percentages
	for key, info := range keyStats {
		info.Percentage = float64(info.Count) / float64(totalObjects) * 100
		info.IsCore = isCoreKey(key, info.Percentage)
	}

	// Core Google Docs API keys
	coreKeys := []string{
		"documentId", "title", "body", "content", "paragraph", "elements", "textRun",
		"textStyle", "paragraphStyle", "namedStyleType", "bold", "italic", "underline",
		"fontSize", "fontFamily", "foregroundColor", "backgroundColor", "link", "url",
		"startIndex", "endIndex", "table", "tableRows", "tableCells", "list", "bullet",
		"headingId", "inlineObjects", "imageProperties", "headers", "footers",
	}

	fmt.Printf("Google Docs JSON Key Analysis - Core Fields Focus\n")
	fmt.Printf("=================================================\n")
	fmt.Printf("File: %s\n", filename)
	fmt.Printf("Total objects: %d\n\n", totalObjects)

	fmt.Println("CORE STRUCTURAL KEYS:")
	fmt.Println("====================")
	fmt.Printf("%-25s %-8s %-8s %-15s %s\n", "Key", "Count", "% Occur", "Required?", "Data Types")
	fmt.Println(strings.Repeat("-", 80))

	for _, key := range coreKeys {
		if info, exists := keyStats[key]; exists {
			required := "OPTIONAL"
			if info.Percentage > 95 {
				required = "REQUIRED"
			} else if info.Percentage > 50 {
				required = "COMMON"
			}

			typesStr := strings.Join(info.DataTypes, ", ")
			if len(typesStr) > 15 {
				typesStr = typesStr[:12] + "..."
			}

			fmt.Printf("%-25s %-8d %-8.1f%% %-15s %s\n",
				key, info.Count, info.Percentage, required, typesStr)
		} else {
			fmt.Printf("%-25s %-8s %-8s %-15s %s\n",
				key, "0", "0.0%", "MISSING", "")
		}
	}

	fmt.Println("\nSTYLING KEYS:")
	fmt.Println("=============")
	styleKeys := []string{
		"bold", "italic", "underline", "strikethrough", "smallCaps",
		"fontSize", "fontFamily", "weight", "foregroundColor", "backgroundColor",
		"alignment", "lineSpacing", "spaceAbove", "spaceBelow", "indentStart", "indentEnd",
	}

	for _, key := range styleKeys {
		if info, exists := keyStats[key]; exists {
			required := "OPTIONAL"
			if info.Percentage > 95 {
				required = "REQUIRED"
			} else if info.Percentage > 50 {
				required = "COMMON"
			}

			typesStr := strings.Join(info.DataTypes, ", ")
			if len(typesStr) > 15 {
				typesStr = typesStr[:12] + "..."
			}

			fmt.Printf("%-25s %-8d %-8.1f%% %-15s %s\n",
				key, info.Count, info.Percentage, required, typesStr)
		}
	}

	fmt.Println("\nHIGH FREQUENCY KEYS (>10% occurrence):")
	fmt.Println("=====================================")

	type keyFreq struct {
		key  string
		info *KeyInfo
	}

	var highFreqKeys []keyFreq
	for key, info := range keyStats {
		if info.Percentage > 10 {
			highFreqKeys = append(highFreqKeys, keyFreq{key, info})
		}
	}

	sort.Slice(highFreqKeys, func(i, j int) bool {
		return highFreqKeys[i].info.Percentage > highFreqKeys[j].info.Percentage
	})

	for _, kf := range highFreqKeys {
		typesStr := strings.Join(kf.info.DataTypes, ", ")
		if len(typesStr) > 15 {
			typesStr = typesStr[:12] + "..."
		}
		fmt.Printf("%-25s %-8d %-8.1f%% %s\n",
			kf.key, kf.info.Count, kf.info.Percentage, typesStr)
	}
}

func analyzeJSON(data any, keyStats map[string]*KeyInfo) {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			if keyStats[key] == nil {
				keyStats[key] = &KeyInfo{
					DataTypes: []string{},
				}
			}

			keyStats[key].Count++

			dataType := getDataType(value)
			found := slices.Contains(keyStats[key].DataTypes, dataType)
			if !found {
				keyStats[key].DataTypes = append(keyStats[key].DataTypes, dataType)
			}

			analyzeJSON(value, keyStats)
		}
	case []any:
		for _, item := range v {
			analyzeJSON(item, keyStats)
		}
	}
}

func countObjects(data any) int {
	count := 0
	switch v := data.(type) {
	case map[string]any:
		count++
		for _, value := range v {
			count += countObjects(value)
		}
	case []any:
		for _, item := range v {
			count += countObjects(item)
		}
	}
	return count
}

func getDataType(value any) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case bool:
		return "bool"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		if len(v) == 0 {
			return "[]"
		}
		return "array"
	case map[string]any:
		return "object"
	default:
		return "unknown"
	}
}

func isCoreKey(key string, percentage float64) bool {
	coreKeys := map[string]bool{
		"documentId": true, "title": true, "body": true, "content": true,
		"paragraph": true, "elements": true, "textRun": true, "textStyle": true,
		"startIndex": true, "endIndex": true,
	}
	return coreKeys[key] || percentage > 50
}
