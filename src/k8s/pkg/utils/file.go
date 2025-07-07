package utils

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/canonical/k8s/pkg/log"
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

	// Remove extra surrounding quotations from the value
	value = strings.Trim(value, "\"")

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

// ParseConfigFile parses a configuration file and returns a map of key-value pairs.
// The file is expected to have lines in the format "key=value". Comments (lines starting with '#') are ignored.
func ParseConfigFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	lines := make([]string, 0)

	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan lines in config file: %w", err)
	}

	config := make(map[string]string, len(lines))
	for _, line := range lines {
		// Split the line into key and value based on the first '=' character
		splitIndex := strings.Index(line, "=")
		if splitIndex == -1 {
			continue // Skip lines that do not contain an '=' character
		}

		key := strings.TrimSpace(line[:splitIndex])
		value := strings.TrimSpace(line[splitIndex+1:])
		config[key] = value
	}

	return config, nil
}

// MinConfigFileDiff searches configuration directories to check whether the minimum
// configurations are set in them. Returns a map with the key, value configurations that
// need to be set to enforce the minimum configuration requirements.
func MinConfigFileDiff(dirs []string, minConfig map[string]string) map[string]string {
	newConfig := DeepCopyMap(minConfig)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.L().Error(err, "Could not parse configuration directory")
			}
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				params, err := ParseConfigFile(filepath.Join(dir, entry.Name()))
				if err != nil {
					log.L().Error(err, "could not parse configuration file, skipping file")
					continue
				}

				for key := range minConfig {
					if value, exists := params[key]; exists {
						// Minimum Configuration already set
						if exists && value >= minConfig[key] {
							delete(newConfig, key)
						}
					}
				}
			}
		}
	}

	return newConfig
}

// UpdateConfigFile takes a map with new configurations in the format "key=value" and
// adjusts the file to reflect the new configuration. The file is expected to have
// lines in the format "key=value". Comments (lines starting with '#') are ignored.
func UpdateConfigFile(path string, newConfig map[string]string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	lines := make([]string, 0)
	updatedKeys := make(map[string]bool)

	// Read through 'tokens' until an EOF is encountered.
	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimSpace(line) // Trim leading and trailing white spaces

		// Ignore empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			splitIndex := strings.Index(line, "=")
			if splitIndex != -1 {
				key := strings.TrimSpace(line[:splitIndex])

				// override value if necessary
				if newValue, exists := newConfig[key]; exists {
					line = key + "=" + newValue
					updatedKeys[key] = true
					delete(newConfig, key)
				}
			}
		}
		lines = append(lines, line)
	}

	// Append remaining keys
	for key := range newConfig {
		if !updatedKeys[key] {
			new_line := key + "=" + newConfig[key]
			lines = append(lines, new_line)
		}
	}

	if err := sc.Err(); err != nil {
		return fmt.Errorf("failed to scan lines in config file: %w", err)
	}

	// Write file with updated lines
	file, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open configuration file for writing %w", err)
	}

	defer file.Close()

	for i := 0; i < len(lines); i++ {
		if _, err := file.WriteString(lines[i] + "\n"); err != nil {
			return fmt.Errorf("failed to write updated lines to configuration file %w", err)
		}
	}
	return nil
}

// Deep copy of a map.
func DeepCopyMap(original map[string]string) map[string]string {
	copied := make(map[string]string, len(original))
	for k, v := range original {
		copied[k] = v
	}
	return copied
}

// GetFileMatch returns the path of the file in a dir matching the regex or "" if no
// match was found.
func GetFileMatch(path string, re *regexp.Regexp) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if match := re.FindString(entry.Name()); match != "" {
			return filepath.Join(path, match), nil
		}
	}
	return "", nil
}

// GetFileMatch returns the path of the file in a dir matching a file like 10-xyz.conf
// using the regex `^(\d+)-.*\.conf$` or returns 0s if no match was found.
func GetHighestConfigFileOrder(path string) (int, error) {
	maxOrder := 0
	re := regexp.MustCompile(`^(\d+)-.*\.conf$`)
	entries, err := os.ReadDir(path)
	if err != nil {
		return maxOrder, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		// Check for configuration file order number
		numStr := matches[1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if num > maxOrder {
			maxOrder = num
		}
	}

	return maxOrder, nil
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

type MountPropagationType string

const (
	MountPropagationShared  MountPropagationType = "shared"
	MountPropagationPrivate MountPropagationType = "private"
	MountPropagationUnknown MountPropagationType = "unknown"
)

// GetMountPropagationType returns the propagation type (shared or private)
// GetMountPropagationType returns ErrUnkownMount if the mount path does not exist.
func GetMountPropagationType(path string) (MountPropagationType, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(path))
	if err != nil {
		return MountPropagationUnknown, fmt.Errorf("failed to get mounts: %w", err)
	}

	if len(mounts) == 0 {
		return MountPropagationUnknown, ErrUnknownMount
	}

	mount := mounts[0]
	if strings.Contains(mount.Optional, string(MountPropagationShared)) {
		return MountPropagationShared, nil
	}
	return MountPropagationPrivate, nil
}

// CreateTarball creates tarball at tarballPath, rooted at rootDir and including
// all files in walkDir except those paths found in excludeFiles.
// walkDir and excludeFiles elements are relative to rootDir.
func CreateTarball(tarballPath string, rootDir string, walkDir string, excludeFiles []string) error {
	tarball, err := os.Create(tarballPath)
	if err != nil {
		return err
	}

	gzWriter := gzip.NewWriter(tarball)
	tarWriter := tar.NewWriter(gzWriter)

	filesys := os.DirFS(rootDir)

	err = fs.WalkDir(filesys, walkDir, func(filepath string, stat fs.DirEntry, err error) error {
		if err != nil {
			msg := fmt.Sprintf("failed to read file while creating tarball; skipping, file: %s, error: %v", filepath, err)
			log.L().Info(msg)
			return nil
		}

		if slices.Contains(excludeFiles, filepath) {
			return nil
		}

		info, err := stat.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, filepath)
		if err != nil {
			return fmt.Errorf("create tar header for %q, error: %w", filepath, err)
		}

		// header.Name is the basename of `stat` by default
		header.Name = filepath

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("failed to write tar header, error: %w", err)
		}

		// Only write contents for regular files
		if header.Typeflag == tar.TypeReg {
			fullPath := path.Join(rootDir, filepath)
			file, err := os.Open(fullPath)
			if err != nil {
				return fmt.Errorf("could not open file: %s, error: %w", fullPath, err)
			}

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("tar write failure: %s, error: %w", fullPath, err)
			}

			err = file.Close()
			if err != nil {
				return fmt.Errorf("could not close file: %s, error: %w", fullPath, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tar walk failed: %s, error: %w", walkDir, err)
	}

	err = tarWriter.Close()
	if err != nil {
		return fmt.Errorf("could not close tar writer, error: %w", err)
	}

	err = gzWriter.Close()
	if err != nil {
		return fmt.Errorf("could not close gz writer, error: %w", err)
	}

	err = tarball.Close()
	if err != nil {
		return fmt.Errorf("could not close tarball, error: %w", err)
	}

	return nil
}

// WriteFile writes data to a file with the given name and permissions.
// The file is written to a temporary file in the same directory as the target file
// and then renamed to the target file to avoid partial writes in case of a crash.
func WriteFile(name string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(name)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpFile.Chmod(perm); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), name); err != nil {
		return fmt.Errorf("failed to rename temp file to target file: %w", err)
	}

	return nil
}

// IsYaml returns true if the file has a yaml or yml extension.
func IsYaml(file string) bool {
	return filepath.Ext(strings.ToLower(file)) == ".yaml" || filepath.Ext(strings.ToLower(file)) == ".yml"
}
