package utils

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

// TODO (KU-167): Replace this with a proper snap helper/interface
var (
	SNAP_DATA   = os.Getenv("SNAP_DATA")
	SNAP_COMMON = os.Getenv("SNAP_COMMON")
	SNAP        = os.Getenv("SNAP")
)

// TemplateAndSave compiles a template with the data and saves it to the given target path.
func TemplateAndSave(tmplFile string, data any, target string) error {
	tmpl := template.Must(template.ParseFiles(tmplFile))

	f, err := os.Create(target)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

// ChmodRecursive changes permissions of files and folders recursively.
func ChmodRecursive(name string, mode fs.FileMode) error {
	err := filepath.WalkDir(name, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk into path: %w", err)
		}

		err = os.Chmod(path, mode)
		if err != nil {
			return fmt.Errorf("failed to change permissions: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to change permissions recursively: %w", err)
	}

	return nil
}

// GetServiceArgument returns the value from `--argument=value` in a service arguments file.
func GetServiceArgument(service string, argument string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf("%s=(.+)", argument))

	b, err := os.ReadFile(filepath.Join(SNAP_DATA, "args", service)) // just pass the file name
	if err != nil {
		return "", fmt.Errorf("failed to read args file: %w", err)
	}

	matches := re.FindStringSubmatch(string(b))

	if len(matches) < 2 {
		return "", fmt.Errorf("failed to find argument in args file: %w", err)
	}

	return matches[1], nil
}

// CopyDirectory recursively copies files and directories from the given srcDir.
//
// https://stackoverflow.com/a/56314145
func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := CopyFile(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("could not read symlink: %w", err)
	}
	return os.Symlink(link, dest)
}
