package utils

import (
	"bufio"
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

func FileExists(path ...string) (bool, error) {
	if _, err := os.Stat(filepath.Join(path...)); err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to stat: %w", err)
		}
		return false, nil
	}
	return true, nil
}
