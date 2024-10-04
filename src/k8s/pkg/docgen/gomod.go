package docgen

import (
	"fmt"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"os"
	"path"
	"strings"
)

func getGoDepModulePath(name string, version string) (string, error) {
	cachePath := os.Getenv("GOMODCACHE")
	if cachePath == "" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			goPath = path.Join(os.Getenv("HOME"), "/go")
		}
		cachePath = path.Join(goPath, "pkg", "mod")
	}

	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", fmt.Errorf(
			"couldn't escape module path: %s %v", name, err)
	}

	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", fmt.Errorf(
			"couldn't escape module version: %s %v", version, err)
	}

	path := path.Join(cachePath, escapedPath+"@"+escapedVersion)

	// Validate the path.
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf(
			"Go module path not accessible: %s %s %s. Error: %v.",
			name, version, path, err)
	}

	return path, nil
}

func getDependencyVersionFromGoMod(goModPath string, packageName string, directOnly bool) (string, string, error) {
	goModContents, err := os.ReadFile(goModPath)
	if err != nil {
		return "", "", fmt.Errorf("could not read go.mod file %s. Error: ", goModPath, err)
	}
	goModFile, err := modfile.ParseLax(goModPath, goModContents, nil)
	if err != nil {
		return "", "", fmt.Errorf("could not parse go.mod file %s. Error: ", goModPath, err)
	}

	for _, dep := range goModFile.Require {
		if directOnly && dep.Indirect {
			continue
		}
		if strings.HasPrefix(packageName, dep.Mod.Path) {
			return dep.Mod.Path, dep.Mod.Version, nil
		}
	}

	return "", "", fmt.Errorf("could not find dependency %s in %s", packageName, goModPath)
}

func getGoModPath(projectDir string) (string, error) {
	return path.Join(projectDir, "go.mod"), nil
}

func getGoPackageDir(packageName string, projectDir string) (string, error) {
	if packageName == "" {
		return "", fmt.Errorf("could not retrieve package dir, no package name specified.")
	}

	if strings.HasPrefix(packageName, "github.com/canonical/k8s/") {
		return strings.Replace(packageName, "github.com/canonical/k8s", projectDir, 1), nil
	}

	// Dependency, need to retrieve its version from go.mod.
	goModPath, err := getGoModPath(projectDir)
	if err != nil {
		return "", err
	}

	basePackageName, version, err := getDependencyVersionFromGoMod(goModPath, packageName, false)
	if err != nil {
		return "", err
	}

	basePath, err := getGoDepModulePath(basePackageName, version)
	if err != nil {
		return "", err
	}

	subPath := strings.TrimPrefix(packageName, basePackageName)
	return path.Join(basePath, subPath), nil
}
