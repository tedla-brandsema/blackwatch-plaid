package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendFile(t *testing.T) {
	t.Run("Append to existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "append_test")

		initialContent := []byte("Hello, ")
		additionalContent := []byte("World!")

		// Write initial content
		if err := os.WriteFile(testFilePath, initialContent, 0644); err != nil {
			t.Fatalf("Failed to write initial content: %v", err)
		}

		// Append additional content
		if err := appendFile(testFilePath, additionalContent); err != nil {
			t.Fatalf("Failed to append content: %v", err)
		}

		// Read and verify the result
		finalContent, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		expectedContent := append(initialContent, additionalContent...)
		if !bytes.Equal(finalContent, expectedContent) {
			t.Fatalf("Expected content %q, got %q", expectedContent, finalContent)
		}
	})

	t.Run("Append to non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "non_existent_append_test")

		content := []byte("New file content")
		if err := appendFile(testFilePath, content); err != nil {
			t.Fatalf("Failed to append content to new file: %v", err)
		}

		// Verify content
		finalContent, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if !bytes.Equal(finalContent, content) {
			t.Fatalf("Expected content %q, got %q", content, finalContent)
		}
	})
}

func TestOverwriteFile(t *testing.T) {
	t.Run("Overwrite existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "overwrite_test")

		initialContent := []byte("Old content")
		newContent := []byte("New content")

		// Write initial content
		if err := os.WriteFile(testFilePath, initialContent, 0644); err != nil {
			t.Fatalf("Failed to write initial content: %v", err)
		}

		// Overwrite with new content
		if err := overwriteFile(testFilePath, newContent); err != nil {
			t.Fatalf("Failed to overwrite content: %v", err)
		}

		// Read and verify
		finalContent, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if !bytes.Equal(finalContent, newContent) {
			t.Fatalf("Expected content %q, got %q", newContent, finalContent)
		}
	})

	t.Run("Overwrite non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "non_existent_overwrite_test")

		content := []byte("New content")
		if err := overwriteFile(testFilePath, content); err != nil {
			t.Fatalf("Failed to overwrite non-existent file: %v", err)
		}

		// Verify content
		finalContent, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if !bytes.Equal(finalContent, content) {
			t.Fatalf("Expected content %q, got %q", content, finalContent)
		}
	})
}

func TestReadFile(t *testing.T) {
	t.Run("Read existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "read_test")

		content := []byte("Test content")
		if err := os.WriteFile(testFilePath, content, 0644); err != nil {
			t.Fatalf("Failed to write content: %v", err)
		}

		// Read and verify
		readContent, err := readFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if !bytes.Equal(readContent, content) {
			t.Fatalf("Expected content %q, got %q", content, readContent)
		}
	})

	t.Run("Read non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "non_existent_read_test")

		_, err := readFile(testFilePath)
		if err == nil {
			t.Fatalf("Expected error, got none")
		}
	})
}

func TestAddLengthPrefix(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedError error
		expectedSize  int
	}{
		{
			name:          "Valid small data",
			input:         []byte("hello"),
			expectedError: nil,
			expectedSize:  prefixByteSize + 5, // 4 bytes for prefix + 5 bytes of data
		},
		{
			name:          "Empty data",
			input:         []byte{},
			expectedError: nil,
			expectedSize:  prefixByteSize, // Only the prefix
		},
		{
			name:          "Maximum allowed size",
			input:         bytes.Repeat([]byte{0xAA}, math.MaxUint32),
			expectedError: nil,
			expectedSize:  prefixByteSize + int(math.MaxUint32),
		},
		{
			name:          "Exceeding maximum size",
			input:         bytes.Repeat([]byte{0xBB}, math.MaxUint32+1),
			expectedError: fmt.Errorf("maximum allowed bytes %d exceeded: found %d", math.MaxUint32, math.MaxUint32+1),
			expectedSize:  0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := addLengthPrefix(test.input)

			if test.expectedError != nil {
				if err == nil || err.Error() != test.expectedError.Error() {
					t.Fatalf("expected error %v, got %v", test.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(output) != test.expectedSize {
					t.Fatalf("expected size %d, got %d", test.expectedSize, len(output))
				}
				// Verify prefix
				prefix := binary.LittleEndian.Uint32(output[:prefixByteSize])
				if int(prefix) != len(test.input) {
					t.Fatalf("expected prefix %d, got %d", len(test.input), prefix)
				}
			}
		})
	}
}

func TestRemoveLengthPrefix(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedError error
		expectedData  []byte
	}{
		{
			name:          "Valid prefixed data",
			input:         append([]byte{0x05, 0x00, 0x00, 0x00}, []byte("hello")...), // 5 in Little Endian
			expectedError: nil,
			expectedData:  []byte("hello"),
		},
		{
			name:          "Empty prefixed data",
			input:         append([]byte{0x00, 0x00, 0x00, 0x00}, []byte{}...), // 0 length prefix
			expectedError: nil,
			expectedData:  []byte{},
		},
		{
			name:          "Missing bytes",
			input:         []byte{0x05, 0x00, 0x00, 0x00}, // Indicates 5 bytes of data but none present
			expectedError: errors.New("missing bytes"),
			expectedData:  nil,
		},
		{
			name:          "No prefixed data found",
			input:         []byte{0x01}, // Less than prefixByteSize
			expectedError: errors.New("no prefixed data found"),
			expectedData:  nil,
		},
		{
			name:          "Exceeding maximum allowed data size",
			input:         append([]byte{0xFF, 0xFF, 0xFF, 0xFF}, bytes.Repeat([]byte{0xAA}, math.MaxUint32+1)...),
			expectedError: fmt.Errorf("maximum allowed bytes %d exceeded: found %d", math.MaxUint32+prefixByteSize, math.MaxUint32+1+prefixByteSize),
			expectedData:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := removeLengthPrefix(test.input)

			if test.expectedError != nil {
				if err == nil || err.Error() != test.expectedError.Error() {
					t.Fatalf("expected error %v, got %v", test.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !bytes.Equal(output, test.expectedData) {
					t.Fatalf("expected data %v, got %v", test.expectedData, output)
				}
			}
		})
	}
}

func TestWriteAndReadPrefixedData(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test_file")

	tests := []struct {
		name        string
		mode        FileMode
		data        []byte
		expectedErr error
	}{
		{
			name:        "Write append mode with valid data",
			mode:        APPEND,
			data:        []byte("test data"),
			expectedErr: nil,
		},
		{
			name:        "Write overwrite mode with valid data",
			mode:        OVERWRITE,
			data:        []byte("new data"),
			expectedErr: nil,
		},
		{
			name:        "Write with empty data in append mode",
			mode:        APPEND,
			data:        []byte{},
			expectedErr: nil,
		},
		{
			name:        "Write with unknown mode",
			mode:        FileMode(999),
			data:        []byte("unknown mode"),
			expectedErr: errors.New("uknown FileMode: 999"),
		},
		{
			name:        "Write data exceeding MaxUint32",
			mode:        OVERWRITE,
			data:        make([]byte, math.MaxUint32+1),
			expectedErr: errors.New("maximum allowed bytes 4294967295 exceeded: found 4294967296"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writePrefixedData(testFilePath, tt.mode, tt.data)
			if (err != nil && tt.expectedErr == nil) || (err == nil && tt.expectedErr != nil) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedErr, err)
			} else if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("Expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}

	t.Run("Read back prefixed data after overwrite", func(t *testing.T) {
		data := []byte("hello world")
		if err := writePrefixedData(testFilePath, OVERWRITE, data); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		readData, err := readPrefixedData(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read prefixed data: %v", err)
		}

		if string(readData) != string(data) {
			t.Errorf("Expected data: %s, got: %s", string(data), string(readData))
		}
	})

	t.Run("Append multiple pieces of data and read last", func(t *testing.T) {
		data1 := []byte("first piece")
		data2 := []byte("second piece")
		if err := writePrefixedData(testFilePath, APPEND, data1); err != nil {
			t.Fatalf("Failed to append first data: %v", err)
		}
		if err := writePrefixedData(testFilePath, APPEND, data2); err != nil {
			t.Fatalf("Failed to append second data: %v", err)
		}

		// Read the entire file and split into prefixed chunks for validation
		content, err := readFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		readData1, err := removeLengthPrefix(content)
		if err != nil {
			t.Fatalf("Failed to remove length prefix for first piece: %v", err)
		}

		readData2, err := removeLengthPrefix(content[len(data1)+8:]) // Skip prefix and data
		if err != nil {
			t.Fatalf("Failed to remove length prefix for second piece: %v", err)
		}

		if string(readData1) != string(data1) || string(readData2) != string(data2) {
			t.Errorf("Appended data mismatch. Got: %s, %s. Expected: %s, %s", string(readData1), string(readData2), string(data1), string(data2))
		}
	})

	t.Run("Read from a file with corrupted prefix", func(t *testing.T) {
		corruptedData := append([]byte{0xFF, 0xFF, 0xFF, 0xFF}, []byte("corrupted")...)
		if err := overwriteFile(testFilePath, corruptedData); err != nil {
			t.Fatalf("Failed to write corrupted data: %v", err)
		}

		_, err := readPrefixedData(testFilePath)
		if err == nil || err.Error() != "missing bytes" {
			t.Errorf("Expected 'missing bytes' error, got: %v", err)
		}
	})
}
