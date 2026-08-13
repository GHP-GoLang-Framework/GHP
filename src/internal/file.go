// Package internal provides shared helpers used across ghp subcommands.
package internal

import "os"

// filePerm is the permission bits used when writing new files.
const filePerm os.FileMode = 0o644

// ReadFile reads the whole file at path and returns its raw bytes. Bytes are
// returned as-is so callers can run regexp directly on []byte or pipe the
// content back into WriteFile without extra copies.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to path with filePerm, creating or truncating the
// file as needed. It returns the underlying filesystem error, if any.
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, filePerm)
}
