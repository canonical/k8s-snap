package docgen

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"reflect"
)

var packageDocCache = make(map[string]*doc.Package)

func findTypeSpec(decl *ast.GenDecl, symbol string) *ast.TypeSpec {
	for _, spec := range decl.Specs {
		typeSpec := spec.(*ast.TypeSpec)
		if symbol == typeSpec.Name.Name {
			return typeSpec
		}
	}
	return nil
}

func getStructTypeFromDoc(packageDoc *doc.Package, structName string) *ast.StructType {
	for _, docType := range packageDoc.Types {
		if structName != docType.Name {
			continue
		}
		typeSpec := findTypeSpec(docType.Decl, docType.Name)
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			// Not a structure.
			continue
		}
		return structType
	}
	return nil
}

func parsePackageDir(packageDir string) (*ast.Package, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, packageDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse go package: %s", packageDir)
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no go package found: %s", packageDir)
	}
	if len(packages) > 1 {
		return nil, fmt.Errorf("multiple go package found: %s", packageDir)
	}

	// We have a map containing a single entry and we need to return it.
	for _, pkg := range packages {
		return pkg, nil
	}

	// shouldn't really get here.
	return nil, fmt.Errorf("failed to parse go package")
}

func getAstStructField(structType *ast.StructType, fieldName string) *ast.Field {
	for _, field := range structType.Fields.List {
		for _, fieldIdent := range field.Names {
			if fieldIdent.Name == fieldName {
				return field
			}
		}
	}
	return nil
}

func getPackageDoc(packagePath string, projectDir string) (*doc.Package, error) {
	packageDoc, found := packageDocCache[packagePath]
	if found {
		return packageDoc, nil
	}

	packageDir, err := getGoPackageDir(packagePath, projectDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve package dir, error: %w", err)
	}

	pkg, err := parsePackageDir(packageDir)
	if err != nil {
		return nil, err
	}

	packageDoc = doc.New(pkg, packageDir, doc.AllDecls|doc.PreserveAST)
	packageDocCache[packagePath] = packageDoc

	return packageDoc, nil
}

func getFieldDocstring(i any, field reflect.StructField, projectDir string) (string, error) {
	inType := reflect.TypeOf(i)

	packageDoc, err := getPackageDoc(inType.PkgPath(), projectDir)
	if err != nil {
		return "", err
	}

	structType := getStructTypeFromDoc(packageDoc, inType.Name())
	if structType == nil {
		return "", fmt.Errorf("could not find %s structure definition", inType.Name())
	}

	astField := getAstStructField(structType, field.Name)
	if astField == nil {
		return "", fmt.Errorf("could not find %s.%s field definition", inType.Name(), field.Name)
	}

	return astField.Doc.Text(), nil
}
