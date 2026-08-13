package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (path string)
		want    []byte
		wantErr bool
	}{
		{
			name: "reads existing file",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "input.txt")
				if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
					t.Fatalf("setup: %v", err)
				}
				return path
			},
			want: []byte("hello"),
		},
		{
			name: "empty file returns empty bytes",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "empty.txt")
				if err := os.WriteFile(path, nil, 0o644); err != nil {
					t.Fatalf("setup: %v", err)
				}
				return path
			},
			want: []byte{},
		},
		{
			name: "missing file returns error",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nope.txt")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			got, err := ReadFile(path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadFile(%q) err = %v, wantErr %v", path, err, tt.wantErr)
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("ReadFile(%q) = %q, want %q", path, got, tt.want)
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantPerm os.FileMode
	}{
		{
			name:     "writes data to new file",
			data:     []byte("file contents"),
			wantPerm: 0o644,
		},
		{
			name:     "empty data creates empty file",
			data:     nil,
			wantPerm: 0o644,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out.txt")

			if err := WriteFile(path, tt.data); err != nil {
				t.Fatalf("WriteFile(%q) err = %v", path, err)
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read back %q: %v", path, err)
			}
			if string(got) != string(tt.data) {
				t.Errorf("WriteFile(%q) = %q, want %q", path, got, tt.data)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat %q: %v", path, err)
			}
			if perm := info.Mode().Perm(); perm != tt.wantPerm {
				t.Errorf("WriteFile(%q) perm = %v, want %v", path, perm, tt.wantPerm)
			}
		})
	}
}

func TestWriteFileError(t *testing.T) {
	err := WriteFile("/no/such/dir/out.txt", []byte("x"))
	if err == nil {
		t.Fatal("WriteFile() on missing dir err = nil, want error")
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roundtrip.txt")
	want := []byte("written and read back")

	if err := WriteFile(path, want); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("round-trip = %q, want %q", got, want)
	}
}
