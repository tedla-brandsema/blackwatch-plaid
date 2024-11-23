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

const (
	maxTotalBytes = math.MaxUint32
	prefixBytes   = 3 // store up to math.MaxUint32
	maxDataBytes  = maxTotalBytes - prefixBytes
)

func addLengthPrefix(data []byte) ([]byte, error) {
	var err error

	var b []byte
	buff := bytes.NewBuffer(b)

	dataSize := len(data)
	if dataSize > maxDataBytes {
		return nil, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", maxDataBytes, dataSize)
	}

	prefix := make([]byte, prefixBytes)
	binary.LittleEndian.PutUint32(prefix, uint32(dataSize))
	if _, err = buff.Write(prefix); err != nil {
		return nil, err
	}

	if data != nil {
		_, err = buff.Write(data)
	}

	return buff.Bytes(), err
}

var corruptDataError = errors.New("corrupt data")

func removeLengthPrefix(data []byte) (uint32, []byte, error) {
	dataSize := len(data)
	if dataSize < prefixBytes {
		return 0, nil, errors.New("no prefixed data found")
	}

	if dataSize > maxTotalBytes {
		return 0, nil, fmt.Errorf("maximum allowed bytes %d exceeded: found %d", maxTotalBytes, dataSize)
	}

	prefix := binary.LittleEndian.Uint32(data[:prefixBytes])
	offset := uint32(prefixBytes) + prefix

	if uint32(dataSize) < offset || prefixBytes >= offset {
		return 0, nil, corruptDataError
	}

	return offset, data[prefixBytes:offset], nil
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
		return fmt.Errorf("uknown FileMode: %d", mode)
	}
}

func readPrefixedData(path string) (uint32, []byte, error) {
	var err error

	content, err := readFile(path)
	if err != nil {
		return 0, nil, err
	}

	return removeLengthPrefix(content)
}
