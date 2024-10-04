package docgen

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type JsonTag struct {
	Name    string
	Options []string
}

type Field struct {
	Name         string
	TypeName     string
	JsonTag      JsonTag
	FullJsonPath string
	Docstring    string
}

// Generate Markdown documentation for a JSON or YAML based on
// the Go structure definition, parsing field annotations.
func MarkdownFromJsonStruct(i any, projectDir string) (string, error) {
	fields, err := ParseStruct(i, projectDir)
	if err != nil {
		return "", err
	}

	entryTemplate := `### %s
**Type:** ` + "`%s`" + `<br>

%s
`

	var out strings.Builder
	for _, field := range fields {
		outFieldType := strings.Replace(field.TypeName, "*", "", -1)
		entry := fmt.Sprintf(entryTemplate, field.FullJsonPath, outFieldType, field.Docstring)
		out.WriteString(entry)
	}

	return out.String(), nil
}

// Generate Markdown documentation for a JSON or YAML based on
// the Go structure definition, parsing field annotations.
// Write the output to the specified file path.
// The project dir is used to identify dependencies based on the go.mod file.
func MarkdownFromJsonStructToFile(i any, outFilePath string, projectDir string) error {
	content, err := MarkdownFromJsonStruct(i, projectDir)
	if err != nil {
		return err
	}

	err = os.WriteFile(outFilePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write markdown documentation to %s, error: %v.",
			outFilePath, err)
	}
	return nil
}

func getJsonTag(field reflect.StructField) JsonTag {
	jsonTag := JsonTag{}

	jsonTagStr := field.Tag.Get("json")
	if jsonTagStr == "" {
		// Use yaml tags as fallback, which have the same format.
		jsonTagStr = field.Tag.Get("yaml")
	}
	if jsonTagStr != "" {
		jsonTagSlice := strings.Split(jsonTagStr, ",")
		if len(jsonTagSlice) > 0 {
			jsonTag.Name = jsonTagSlice[0]
		}
		if len(jsonTagSlice) > 1 {
			jsonTag.Options = jsonTagSlice[1:]
		}
	}

	return jsonTag
}

func ParseStruct(i any, projectDir string) ([]Field, error) {
	inType := reflect.TypeOf(i)

	if inType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("structure parsing failed, not a structure: %s", inType.Name)
	}

	outFields := []Field{}
	fields := reflect.VisibleFields(inType)
	for _, field := range fields {
		jsonTag := getJsonTag(field)
		docstring, err := getFieldDocstring(i, field, projectDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not retrieve field docstring: %s.%s, error: %v",
				inType.Name, field.Name, err)
		}

		if field.Type.Kind() == reflect.Struct {
			fieldIface := reflect.ValueOf(i).FieldByName(field.Name).Interface()
			nestedFields, err := ParseStruct(fieldIface, projectDir)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse %s.%s, error: %v", inType, field.Name, err)
			}

			outField := Field{
				Name:         field.Name,
				TypeName:     "object",
				JsonTag:      jsonTag,
				FullJsonPath: jsonTag.Name,
				Docstring:    docstring,
			}
			outFields = append(outFields, outField)

			for _, nestedField := range nestedFields {
				// Update the json paths of the nested fields based on the field name.
				nestedField.FullJsonPath = jsonTag.Name + "." + nestedField.FullJsonPath
				outFields = append(outFields, nestedField)
			}
		} else {
			outField := Field{
				Name:         field.Name,
				TypeName:     field.Type.String(),
				JsonTag:      jsonTag,
				FullJsonPath: jsonTag.Name,
				Docstring:    docstring,
			}
			outFields = append(outFields, outField)
		}
	}

	return outFields, nil
}
