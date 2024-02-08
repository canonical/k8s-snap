package utils

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/moby/sys/mountinfo"
)

// TemplateAndSave compiles a template with the data and saves it to the given target path.
func TemplateAndSave(tmplFile string, data any, target string) error {
	tmpl := template.Must(template.ParseFiles(tmplFile))

	f, err := os.Create(target)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

// FileExists returns true if the specified path exists.
func FileExists(path ...string) (bool, error) {
	if _, err := os.Stat(filepath.Join(path...)); err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to stat: %w", err)
		}
		return false, nil
	}
	return true, nil
}

// ReadFile returns the file contents as a string.
func ReadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}
	return string(b), nil
}

// ParseArgumentLine parses a command-line argument from a single line.
// The returned key includes any dash prefixes.
func ParseArgumentLine(line string) (key string, value string) {
	line = strings.TrimSpace(line)

	// parse "--argument value" and "--argument=value" variants
	if parts := strings.Split(line, "="); len(parts) >= 2 {
		key = parts[0]
		value = parts[1]
	} else if parts := strings.Split(line, " "); len(parts) >= 2 {
		key = parts[0]
		value = strings.Join(parts[1:], " ")
	} else {
		key = line
	}

	return
}

// Reads an argument file and parses the lines to an <arg, value> map.
func ParseArgumentFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read argument file %s: %w", path, err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	lines := make([]string, 0)

	// Read through 'tokens' until an EOF is encountered.
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan lines in argument file: %w", err)
	}

	args := make(map[string]string, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			a, v := ParseArgumentLine(line)
			args[a] = v
		}
	}
	return args, nil
}

// Serializes a map of service arguments in the format "argument=value" to file.
func SerializeArgumentFile(arguments map[string]string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to write argument file %s: %w", path, err)
	}
	defer file.Close()

	// Order the argument keys alphabetically to make the output deterministic
	keys := make([]string, 0)
	for k := range arguments {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		file.WriteString(fmt.Sprintf("%s=%s\n", k, arguments[k]))
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
			if err := CreateIfNotExists(destPath, 0o755); err != nil {
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

// GetMountPath returns the first mountpath for a given filesystem type.
func GetMountPath(fsType string) (string, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.FSTypeFilter(fsType))
	if err != nil {
		return "", fmt.Errorf("failed to find the mount info for %s: %w", fsType, err)
	}
	if len(mounts) == 0 {
		return "", fmt.Errorf("could not find any %s filesystem mount", fsType)
	}

	return mounts[0].Mountpoint, nil
}
