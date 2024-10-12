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
	"slices"
	"sort"
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
