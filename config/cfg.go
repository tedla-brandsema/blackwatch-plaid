package config

import (
	"encoding/json"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/tedla-brandsema/tribble/internal/fio"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	self *Config
)

func init() {
	var err error

	err = initFolders(tribblePath)
	if err != nil {
		msg := fmt.Sprintf("unable to create folder %s", rootFolder)
		slog.Error(msg,
			slog.Any("error", err),
		)
		os.Exit(1)
		return
	}

	err = load()
	if err != nil {
		slog.Warn("unable to load config file",
			slog.Any("error", err),
		)
		os.Exit(1)
		return
	}
}

func initFolders(paths ...string) error {
	for _, path := range paths {
		if !fio.FileExists(path) {
			err := fio.MakeDir(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	rootFolder    = "."
	tribbleFolder = ".tribble"
	configFile    = "tribble.cfg"
	backlogFile   = "backlog.md"
)

type SerializeMode int

const (
	SerializeMarkdown SerializeMode = iota
	SerializeJSON
	SerializeBinary
)

// Internal vars
var (
	tribblePath = filepath.Join(rootFolder, tribbleFolder)
	configPath  = filepath.Join(tribblePath, configFile)
)

// CFG vars with defaults
var (
	serializeMode = SerializeMarkdown
	backlogPath   = filepath.Join(rootFolder, backlogFile)
)

type Config struct {
	//SerializeMode SerializeMode
	BacklogPath string
}

func NewDefaultConfig() *Config {
	return &Config{
		//SerializeMode: serializeMode,
		BacklogPath: backlogPath,
	}
}

func Get() *Config {
	if self == nil {
		_ = load()
	}

	return self
}

func load() error {
	var err error

	self, err = readConfig()
	if err != nil {
		return err
	}

	return nil
}

func readConfig() (*Config, error) {
	var cfg Config

	if !fio.FileExists(configPath) {
		slog.Info("no config file found: creating config file")
		err := writeConfig(NewDefaultConfig())
		if err != nil {
			return nil, err
		}
	}

	file, err := fio.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func writeConfig(config *Config) error {
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	err = fio.OverwriteFile(configPath, b)
	if err != nil {
		return err
	}
	return nil
}

func filePath(path, file string) string {
	fileSlug := slug.Make(file)
	return filepath.Join(path, fileSlug)
}
