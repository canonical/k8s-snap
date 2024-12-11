package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

// TODO: pass as argument
const sourcesFile = "sources.txt"

// GenerateGoStruct recursively generates Go struct definitions from a YAML Node
func GenerateGoStruct(name string, node *yaml.Node, parent string) string {
	var fields []string
	var nestedStructs []string
	processedLists := make(map[string]bool)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		fieldName := strcase.ToCamel(keyNode.Value)
		fieldType := "any"

		structName := name
		if parent != "" {
			structName = parent + "_" + name
		}

		switch valueNode.Kind {
		case yaml.MappingNode:
			// Nested struct
			nestedStructName := structName + "_" + fieldName
			nestedStructs = append(nestedStructs, GenerateGoStruct(nestedStructName, valueNode, structName))
			fieldType = nestedStructName
		case yaml.SequenceNode:
			if len(valueNode.Content) == 0 {
				fieldType = "[]any"
			} else {
				// List with its own struct
				nestedListName := structName + "_" + fieldName + "Item"
				if !processedLists[nestedListName] {
					nestedStructs = append(nestedStructs, GenerateGoStruct(nestedListName, valueNode.Content[0], structName))
					processedLists[nestedListName] = true
				}
				fieldType = "[]" + nestedListName
			}
		case yaml.ScalarNode:
			// Scalar value
			fieldType = "any"
		}

		if lines := extractComments(keyNode); len(lines) > 0 {
			fields = append(fields, strings.Join(lines, "\n"))
		}
		fields = append(fields, fmt.Sprintf("\t%s %s", fieldName, fieldType))
	}

	structDef := fmt.Sprintf("type %s struct {\n%s\n}", name, strings.Join(fields, "\n"))
	return strings.Join(append(nestedStructs, structDef), "\n\n")
}

// extractComments extracts comments from a YAML node
func extractComments(n *yaml.Node) []string {
	totalLines := []string{}
	if hc := n.HeadComment; hc != "" {
		lines := strings.Split(hc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}
	if lc := n.LineComment; lc != "" {
		lines := strings.Split(lc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}
	if fc := n.FootComment; fc != "" {
		lines := strings.Split(fc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}

	return totalLines
}

// ProcessYAMLFile reads a YAML file and generates the corresponding Go struct
func ProcessYAMLFile(filePath, structName, outputFile string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file %s: %w", filePath, err)
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(data, &rootNode); err != nil {
		return fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
	}

	goStruct := GenerateGoStruct(structName, rootNode.Content[0], "")

	output := fmt.Sprintf("package main\n\n%s\n", goStruct)
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write Go file %s: %w", outputFile, err)
	}

	if err := formatGoFile(outputFile); err != nil {
		return fmt.Errorf("failed to format Go file %s: %w", outputFile, err)
	}

	fmt.Printf("Generated %s\n", outputFile)
	return nil
}

// formatGoFile formats a Go file using gofmt
func formatGoFile(filePath string) error {
	cmd := exec.Command("gofmt", "-w", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to format %s: %v\nOutput: %s", filePath, err, out.String())
	}
	return nil
}

func main() {
	file, err := os.Open(sourcesFile)
	if err != nil {
		log.Fatalf("Failed to open sources file: %v\n", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		yamlFile := scanner.Text()

		if yamlFile == "" {
			// empty line
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(yamlFile), filepath.Ext(yamlFile))
		structName := strcase.ToCamel(strings.ReplaceAll(strings.ReplaceAll(baseName, ".", "_"), "-", "_"))
		outputFile := fmt.Sprintf("%s.go", baseName)

		if err := ProcessYAMLFile(yamlFile, structName, outputFile); err != nil {
			log.Fatalf("Error processing file %s: %v\n", yamlFile, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading sources file: %v\n", err)
	}
}
