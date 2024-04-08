package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moby/sys/mountinfo"
)

// ParseArgumentLine parses a command-line argument from a single line.
// The returned key includes any dash prefixes.
func ParseArgumentLine(line string) (key string, value string) {
	line = strings.TrimSpace(line) // Trim leading and trailing white spaces

	// parse "--argument value", "--argument=value", "--argument=value=,othervalue=" variants

	splitIndex := -1
	for i, c := range line {
		if c == ' ' || c == '=' {
			splitIndex = i
			break
		}
	}

	if splitIndex == -1 {
		// If no space or equal sign is found, return the line as key
		return line, ""
	}

	// Split the line into key and value based on the split index
	key = line[:splitIndex]
	value = strings.TrimSpace(line[splitIndex+1:]) // Remove any leading space in value

	return key, value
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
func FileExists(path ...string) (bool, error) {
	if _, err := os.Stat(filepath.Join(path...)); err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to stat: %w", err)
		}
		return false, nil
	}
	return true, nil
}

var ErrUnknownMount = errors.New("mount is unknown")

// GetMountPath returns the first mountpath for a given filesystem type.
// GetMountPath returns ErrUnkownMount if the mount path does not exist.
func GetMountPath(fsType string) (string, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.FSTypeFilter(fsType))
	if err != nil {
		return "", fmt.Errorf("failed to find the mount info for %s: %w", fsType, err)
	}
	if len(mounts) == 0 {
		return "", ErrUnknownMount
	}

	return mounts[0].Mountpoint, nil
}

// GetMountPropagation returns the propagation type (shared or private)
// GetMountPropagation returns ErrUnkownMount if the mount path does not exist.
func GetMountPropagation(path string) (string, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(path))
	if err != nil {
		return "", fmt.Errorf("failed to get mounts: %w", err)
	}

	if len(mounts) == 0 {
		return "", ErrUnknownMount
	}

	mount := mounts[0]
	if strings.Contains(mount.Optional, "shared") {
		return "shared", nil
	}
	return "private", nil
}
