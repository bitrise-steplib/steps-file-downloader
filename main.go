package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/filedownloader"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// Input holds the raw step inputs parsed from environment variables.
type Input struct {
	Source         string `env:"source,required"`
	Destination    string `env:"destination,required"`
	FilePermission string `env:"file_permission"`
}

// Config holds the validated step configuration.
type Config struct {
	Source         string
	Destination    string
	FilePermission *os.FileMode
}

// Step downloads a remote file to the build machine.
type Step struct {
	inputParser  stepconf.InputParser
	logger       log.Logger
	pathModifier pathutil.PathModifier
	downloader   filedownloader.Downloader
}

// NewStep creates a Step with injected dependencies.
func NewStep(inputParser stepconf.InputParser, logger log.Logger, pathModifier pathutil.PathModifier, downloader filedownloader.Downloader) Step {
	return Step{
		inputParser:  inputParser,
		logger:       logger,
		pathModifier: pathModifier,
		downloader:   downloader,
	}
}

// ProcessConfig parses and validates the step inputs.
func (s Step) ProcessConfig() (Config, error) {
	var input Input
	if err := s.inputParser.Parse(&input); err != nil {
		return Config{}, fmt.Errorf("failed to parse inputs: %w", err)
	}
	stepconf.Print(input)

	destination, err := s.pathModifier.AbsPath(input.Destination)
	if err != nil {
		return Config{}, fmt.Errorf("failed to expand destination path (%s): %w", input.Destination, err)
	}

	var filePermission *os.FileMode
	if input.FilePermission != "" {
		parsed, err := strconv.ParseUint(input.FilePermission, 8, 32)
		if err != nil {
			return Config{}, fmt.Errorf("invalid file permission (%s): must be octal bits like 644: %w", input.FilePermission, err)
		}
		mode := os.FileMode(parsed)
		filePermission = &mode
	}

	return Config{
		Source:         input.Source,
		Destination:    destination,
		FilePermission: filePermission,
	}, nil
}

// Run downloads the source to the destination and applies the file permission if set.
func (s Step) Run(cfg Config) error {
	dir := filepath.Dir(cfg.Destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory (%s): %w", dir, err)
	}

	s.logger.Infof("Downloading %s to %s", cfg.Source, cfg.Destination)
	if err := s.downloader.Download(context.Background(), cfg.Destination, cfg.Source); err != nil {
		return fmt.Errorf("failed to download %s: %w", cfg.Source, err)
	}
	s.logger.Donef("Downloaded file to %s", cfg.Destination)

	if cfg.FilePermission != nil {
		s.logger.Infof("Setting file permission to %#o", *cfg.FilePermission)
		if err := os.Chmod(cfg.Destination, *cfg.FilePermission); err != nil {
			return fmt.Errorf("failed to set file permission on %s: %w", cfg.Destination, err)
		}
	}

	return nil
}

func main() {
	os.Exit(run())
}

func run() int {
	logger := log.NewLogger()
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	pathModifier := pathutil.NewPathModifier()
	downloader := filedownloader.NewDownloader(logger)
	step := NewStep(inputParser, logger, pathModifier, downloader)

	config, err := step.ProcessConfig()
	if err != nil {
		logger.Errorf(err.Error())
		return 1
	}

	if err := step.Run(config); err != nil {
		logger.Errorf(err.Error())
		return 1
	}

	return 0
}
