package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
)

const backlogFilename = ".backlog"

func appendFile(path string, b []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			slog.Error("unable to close file",
				slog.Any("error", err),
			)
		}
	}()

	if b != nil {
		_, err = file.Write(b)
	}

	return err
}

func overwriteFile(path string, b []byte) error {
	var err error

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			slog.Error("unable to close file",
				slog.Any("error", err),
			)
		}
	}()

	if b != nil {
		_, err = file.Write(b)
	}

	return err
}

func readFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = file.Close(); err != nil {
			slog.Error("unable to close file",
				slog.Any("error", err),
			)
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}

func backlogPath(root string) string {
	return filepath.Join(root, backlogFilename)
}

const prefixByteSize = 4 // store up to math.MaxUint32

func addLengthPrefix(data []byte) ([]byte, error) {
	var err error

	var b []byte
	buff := bytes.NewBuffer(b)

	dataSize := len(data)
	if dataSize > math.MaxUint32 {
		return nil, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", math.MaxUint32, dataSize)
	}

	prefix := make([]byte, prefixByteSize)
	binary.LittleEndian.PutUint32(prefix, uint32(dataSize))
	if _, err = buff.Write(prefix); err != nil {
		return nil, err
	}

	if data != nil {
		_, err = buff.Write(data)
	}

	return buff.Bytes(), err
}

func removeLengthPrefix(data []byte) ([]byte, error) {
	dataSize := len(data)
	if dataSize < prefixByteSize {
		return nil, errors.New("no prefixed data found")
	}

	maximumAllowedBytes := math.MaxUint32 + prefixByteSize
	if dataSize > maximumAllowedBytes {
		return nil, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", maximumAllowedBytes, dataSize)
	}

	prefix := binary.LittleEndian.Uint32(data[:prefixByteSize])
	lastByte := prefixByteSize + prefix

	if uint32(dataSize) < lastByte {
		return nil, errors.New("missing bytes")
	}

	return data[prefixByteSize:lastByte], nil
}

type FileMode int

const (
	APPEND FileMode = iota
	OVERWRITE
)

func writePrefixedData(path string, mode FileMode, data []byte) error {
	var err error

	prefixed, err := addLengthPrefix(data)
	if err != nil {
		return err
	}

	switch mode {
	case APPEND:
		return appendFile(path, prefixed)
	case OVERWRITE:
		return overwriteFile(path, prefixed)
	default:
		return fmt.Errorf("uknown %[1]T: %[1]d", mode)
	}
}

func readPrefixedData(path string) ([]byte, error) {
	var err error

	content, err := readFile(path)
	if err != nil {
		return nil, err
	}

	return removeLengthPrefix(content)
}
