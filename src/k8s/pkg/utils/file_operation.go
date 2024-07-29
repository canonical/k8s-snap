package utils

import (
	"fmt"
	"io"
	"os"
)

// FileOperations is a struct that helps in perfoming file operations like
// backup and write multiple files.
type FileOperations struct {
	BackupPath  string
	SourcePath  string
	Content     []byte
	Permissions os.FileMode
}

// BackupFiles backs up the files in the operations slice.
func BackupFiles(operations []FileOperations) error {
	for _, op := range operations {
		if err := backupFile(op.SourcePath, op.BackupPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}
	return nil
}

// WriteFiles writes the files in the operations slice.
func WriteFiles(operations []FileOperations) error {
	for _, op := range operations {
		if err := os.WriteFile(op.SourcePath, op.Content, op.Permissions); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}
	return nil
}

// backupFile backs up the file at sourcePath to backupPath.
func backupFile(sourcePath, backupPath string) error {
	err := copyFile(sourcePath, backupPath)
	if err != nil {
		return fmt.Errorf("failed to backup file: %w", err)
	}
	return nil
}

// CopyFiles copies the files in the operations slice perserving the permissions.
func copyFile(sourcePath, destinationPath string) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer in.Close()

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	out, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	defer out.Close()

	os.Chmod(destinationPath, sourceInfo.Mode())

	if _, err := io.Copy(in, out); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
