package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

const (
	toolName               = "CHART_VALUES_STRUCT_GENERATOR"
	unsafeFieldName        = "UNSAFE_MISC_FIELDS"
	rootStructDocStringFmt = "// %s represents the values of the %s chart"
)

// StructMeta represents a struct in a Go file
type StructMeta struct {
	// IsRoot is true if the struct is the root struct of the file.
	// This struct represents the complete set of values of the yaml file.
	IsRoot bool
	// Name is the Name of the struct.
	Name string
	// DocString is the docstring of the struct.
	// different lines should be separated by \n.
	DocString string
	// Fields is a list of Fields in the struct.
	Fields []*FieldMeta
}

// FieldMeta represents a field in a Go struct
type FieldMeta struct {
	// Name is the Name of the field.
	Name string
	// OriginalYamlName is the original name of the field in the YAML file.
	OriginalYamlName string
	// DocString is the docstring of the field.
	// different lines should be separated by \n.
	DocString string
	// Type is the go type of the field.
	Type string
}

// Import represents an import in a Go file
type Import struct {
	// Alias is the alias of the import.
	Alias string
	// Path is the path of the import.
	Path string
}

// GoRecipe is a recipe to generate a Go file from a YAML file
type GoRecipe struct {
	// RootStructName is the name of the root struct.
	RootStructName string
	// advancedTypesEnabled is true if advanced type inference for fields is enabled.
	advancedTypesEnabled bool
	// templateFilePath is the path to the template file.
	templateFilePath string

	// GenerateCmd is the command that generated the file.
	GenerateCmd string
	// GenerateDate is the date the file was generated.
	GenerateDate string
	// ToolName is the name of the tool that generated the file.
	ToolName string

	// UnsafeFieldEnabled is true if the unsafe field is enabled.
	// The unsafe field is a map[string]any field that can be used to handle any additional fields.
	UnsafeFieldEnabled bool
	// UnsafeFieldName is the name of the unsafe field.
	UnsafeFieldName string
	// PkgName is the name of the package.
	PkgName string
	Imports []Import
	// Structs is a list of Structs in the file.
	Structs []*StructMeta
}

// fill recursively generates Go recipe definitions from a YAML Node
func (recipe *GoRecipe) fill(structName string, node *yaml.Node, docString string, isRoot bool) {
	stMeta := &StructMeta{
		IsRoot:    isRoot,
		Name:      structName,
		DocString: docString,
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		fieldName := strcase.ToCamel(keyNode.Value)

		field := &FieldMeta{
			Name:             fieldName,
			OriginalYamlName: keyNode.Value,
			DocString:        strings.Join(extractComments(keyNode, valueNode), "\n"),
		}

		// TODO: handle such cases:
		// controller:
		// 	<<: *defaults
		if keyNode.Value == "<<" {
			continue
		}

		switch valueNode.Kind {
		case yaml.MappingNode:
			// nested struct
			if len(valueNode.Content) == 0 {
				field.Type = infereTypeString(valueNode, recipe.advancedTypesEnabled, false)
			} else {
				// struct of known type, the type will be the name of the struct
				nestedStructName := structName + "_" + fieldName
				field.Type = "*" + nestedStructName
				recipe.fill(nestedStructName, valueNode, field.DocString, false)
			}
		case yaml.SequenceNode:
			if len(valueNode.Content) == 0 || len(valueNode.Content[0].Content) == 0 {
				field.Type = infereTypeString(valueNode, recipe.advancedTypesEnabled, false)
			} else {
				// list with its own struct
				nestedListName := structName + "_" + fieldName + "Item"
				field.Type = "*[]" + nestedListName
				recipe.fill(nestedListName, valueNode.Content[0], field.DocString, false)
			}
		case yaml.ScalarNode:
			// scalar value
			field.Type = infereTypeString(valueNode, recipe.advancedTypesEnabled, false)
		}

		stMeta.Fields = append(stMeta.Fields, field)
	}

	recipe.Structs = append(recipe.Structs, stMeta)
}

// generateGoFile generates a Go file from a recipe
func (recipe *GoRecipe) generateGoFile(outputFilePath string) error {
	tmpl, err := template.New(recipe.templateFilePath).ParseFiles(recipe.templateFilePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", recipe.templateFilePath, err)
	}

	var out *os.File
	out, err = os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create Go file %s: %w", outputFilePath, err)
	}

	if err := out.Chmod(0644); err != nil {
		return fmt.Errorf("failed to change permissions of Go file %s: %w", outputFilePath, err)
	}

	if err := tmpl.Execute(out, recipe); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return fmt.Errorf("failed to format Go file %s: %w", outputFilePath, err)
	}

	fmt.Printf("Generated %s\n", outputFilePath)
	return nil
}

// extractComments extracts comments from a YAML node
func extractComments(keyNode, valNode *yaml.Node) []string {
	totalLines := []string{}

	if hc := keyNode.HeadComment; hc != "" {
		lines := strings.Split(hc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}

	if lc := keyNode.LineComment; lc != "" {
		lines := strings.Split(lc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}

	if fc := keyNode.FootComment; fc != "" {
		lines := strings.Split(fc, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimLeft(l, "#")
			totalLines = append(totalLines, fmt.Sprintf("// %s", l))
		}
	}

	if valNode.Value != "" {
		if len(totalLines) != 0 {
			totalLines = append(totalLines, "//")
		}
		totalLines = append(totalLines, fmt.Sprintf("// Default value in yaml: %s", valNode.Value))
	} else if valNode.Kind == yaml.SequenceNode {
		if len(valNode.Content) != 0 && len(valNode.Content[0].Content) == 0 {
			if len(totalLines) != 0 {
				totalLines = append(totalLines, "//")
			}
			totalLines = append(totalLines, "// Default value in yaml:")
			for _, c := range valNode.Content {
				totalLines = append(totalLines, fmt.Sprintf("// - %s", c.Value))
			}
		}
	}

	return totalLines
}

// formatGoFile formats a Go file using gofmt
func formatGoFile(filePath string) error {
	if err := runCmd("gofmt", "-w", filePath); err != nil {
		return fmt.Errorf("failed to format %s: %w", filePath, err)
	}
	return nil
}

// runCmd runs a command
func runCmd(parts ...string) error {
	if len(parts) == 0 {
		return fmt.Errorf("no command provided")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w\nOutput: %s", err, out.String())
	}

	return nil
}

// infereTypeString infers the Go type of a YAML node
func infereTypeString(n *yaml.Node, advanced bool, isNested bool) string {
	switch n.Kind {
	case yaml.ScalarNode:
		if !advanced {
			return "any"
		}

		switch n.Tag {
		case "!!bool":
			if isNested {
				return "bool"
			}
			return "*bool"
		case "!!int":
			if isNested {
				return "int"
			}
			return "*int64"
		case "!!float":
			if isNested {
				return "float64"
			}
			return "*float64"
		default:
			if isNested {
				return "string"
			}
			return "*string"
		}
	case yaml.SequenceNode:
		if len(n.Content) == 0 || !advanced {
			return "*[]any"
		}
		return "*[]" + infereTypeString(n.Content[0], true, true) // advanced has to be true
	case yaml.MappingNode:
		// advanced inference for maps should be handled by the upper level
		return "*map[string]any"
	default:
		return "any"
	}
}

func main() {
	var (
		templateFilePath     string
		yamlFilesStr         string
		pkgName              string
		outDir               string
		advancedTypesEnabled bool
		unsafeFieldEnabled   bool
	)

	flag.StringVar(&templateFilePath, "template", "struct.go.tmpl", "Path to the template file")
	flag.StringVar(&yamlFilesStr, "files", "", "Comma separated list of YAML files to generate Go structs from")
	flag.StringVar(&pkgName, "pkg", "main", "Name of the package to generate")
	flag.StringVar(&outDir, "out-dir", ".", "Directory where the generated files will be saved")
	flag.BoolVar(&advancedTypesEnabled, "advanced-types", false, "Enable advanced types (e.g. string instead of any where possible)")
	flag.BoolVar(&unsafeFieldEnabled, "unsafe-field", false, "Add a map[string]any field to the root struct to handle any additional fields")
	flag.Parse()

	yamlFilesPaths := strings.Split(yamlFilesStr, ",")
	if len(yamlFilesPaths) == 0 {
		log.Fatalf("No YAML files provided\n")
	}

	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err := os.Mkdir(outDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v\n", err)
		}
	}

	generateCmd := fmt.Sprintf("./%s %s", toolName, strings.Join(os.Args[1:], " "))

	for _, yamlFilePath := range yamlFilesPaths {
		yamlFile, err := os.Open(yamlFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("File %s does not exist\n", yamlFilePath)
			} else {
				log.Fatalf("Error reading file %s: %v\n", yamlFilePath, err)
			}
		}
		defer yamlFile.Close()

		baseName := strings.TrimSuffix(filepath.Base(yamlFilePath), filepath.Ext(yamlFilePath))
		rootStructName := strcase.ToCamel(strings.ReplaceAll(strings.ReplaceAll(baseName, ".", "_"), "-", "_"))
		outputFilePath := path.Join(outDir, fmt.Sprintf("%s.go", baseName))

		recipe := &GoRecipe{
			RootStructName:       rootStructName,
			templateFilePath:     templateFilePath,
			PkgName:              pkgName,
			advancedTypesEnabled: advancedTypesEnabled,
			UnsafeFieldEnabled:   unsafeFieldEnabled,
			GenerateCmd:          generateCmd,
			GenerateDate:         time.Now().Format(time.DateOnly),
			ToolName:             toolName,
			UnsafeFieldName:      unsafeFieldName,
			Imports: []Import{
				{
					Path: "fmt",
				},
				{
					Path: "encoding/json",
				},
				{
					Path: "reflect",
				},
				{
					Path: "strings",
				},
			},
		}

		rootNode := yaml.Node{}
		if err := yaml.NewDecoder(yamlFile).Decode(&rootNode); err != nil {
			log.Fatalf("Error decoding yaml value from file %s: %v\n", yamlFilePath, err)
		}

		if len(rootNode.Content) == 0 {
			log.Fatalf("Empty file %s\n", yamlFilePath)
		}

		docString := fmt.Sprintf(rootStructDocStringFmt, rootStructName, yamlFilePath)
		recipe.fill(rootStructName, rootNode.Content[0], docString, true)

		if err := recipe.generateGoFile(outputFilePath); err != nil {
			log.Fatalf("Failed to generate Go file: %v\n", err)
		}
	}
}
