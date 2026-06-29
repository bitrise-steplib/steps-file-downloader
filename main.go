package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/filedownloader"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

// Input holds the raw step inputs parsed from environment variables.
type Input struct {
	Source         string `env:"source,required"`
	Destination    string `env:"destination,required"`
	FilePermission string `env:"file_permission"`
	UseSudo        bool   `env:"use_sudo"`
}

// Config holds the validated step configuration.
type Config struct {
	Source         string
	Destination    string
	FilePermission *os.FileMode
	UseSudo        bool
}

// Step downloads a remote file to the build machine.
type Step struct {
	inputParser    stepconf.InputParser
	logger         log.Logger
	pathModifier   pathutil.PathModifier
	pathProvider   pathutil.PathProvider
	downloader     filedownloader.Downloader
	commandFactory command.Factory
	fileManager    fileutil.FileManager
}

// NewStep creates a Step with injected dependencies.
func NewStep(
	inputParser stepconf.InputParser,
	logger log.Logger,
	pathModifier pathutil.PathModifier,
	pathProvider pathutil.PathProvider,
	downloader filedownloader.Downloader,
	commandFactory command.Factory,
	fileManager fileutil.FileManager,
) Step {
	return Step{
		inputParser:    inputParser,
		logger:         logger,
		pathModifier:   pathModifier,
		pathProvider:   pathProvider,
		downloader:     downloader,
		commandFactory: commandFactory,
		fileManager:    fileManager,
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
		UseSudo:        input.UseSudo,
	}, nil
}

// Run downloads the source to a temporary file, then places it at the
// destination. If placement fails with a permission error and use_sudo is
// enabled, it retries the placement with sudo.
func (s Step) Run(cfg Config) error {
	tmpDir, err := s.pathProvider.CreateTempDir("file-downloader")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			s.logger.Warnf("Failed to remove temporary directory %s: %s", tmpDir, rmErr)
		}
	}()

	tmpFile := filepath.Join(tmpDir, filepath.Base(cfg.Destination))

	s.logger.Infof("Downloading %s", cfg.Source)
	if err := s.downloader.Download(context.Background(), tmpFile, cfg.Source); err != nil {
		return fmt.Errorf("failed to download %s: %w", cfg.Source, err)
	}

	s.logger.Infof("Placing file at %s", cfg.Destination)
	if err := s.place(cfg, tmpFile, false); err != nil {
		if cfg.UseSudo && errors.Is(err, fs.ErrPermission) {
			s.logger.Warnf("Permission denied placing file at %s, retrying with sudo", cfg.Destination)
			if sudoErr := s.place(cfg, tmpFile, true); sudoErr != nil {
				return fmt.Errorf("failed to place file at %s with sudo: %w", cfg.Destination, sudoErr)
			}
		} else {
			return fmt.Errorf("failed to place file at %s: %w", cfg.Destination, err)
		}
	}

	s.logger.Donef("Downloaded file to %s", cfg.Destination)
	return nil
}

// place creates the destination directory, copies the downloaded file into
// place, and applies the file permission, optionally using sudo.
func (s Step) place(cfg Config, src string, useSudo bool) error {
	dir := filepath.Dir(cfg.Destination)
	if useSudo {
		return s.placeWithSudo(cfg, src, dir)
	} else {
		return s.placeDirect(cfg, src, dir)
	}
}

func (s Step) placeDirect(cfg Config, src, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := s.fileManager.CopyFile(src, cfg.Destination, &fileutil.CopyOptions{Overwrite: true}); err != nil {
		return err
	}
	if cfg.FilePermission != nil {
		if err := os.Chmod(cfg.Destination, *cfg.FilePermission); err != nil {
			return err
		}
	}
	return nil
}

func (s Step) placeWithSudo(cfg Config, src, dir string) error {
	commands := [][]string{
		{"mkdir", "-p", dir},
		{"cp", src, cfg.Destination},
	}
	if cfg.FilePermission != nil {
		mode := fmt.Sprintf("%o", *cfg.FilePermission&os.ModePerm)
		commands = append(commands, []string{"chmod", mode, cfg.Destination})
	}

	for _, args := range commands {
		if err := s.runSudo(args...); err != nil {
			return err
		}
	}
	return nil
}

func (s Step) runSudo(args ...string) error {
	cmd := s.commandFactory.Create("sudo", append([]string{"-n"}, args...), nil)
	s.logger.Printf("$ %s", cmd.PrintableCommandArgs())
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w (output: %s)", cmd.PrintableCommandArgs(), err, out)
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
	pathProvider := pathutil.NewPathProvider()
	downloader := filedownloader.NewDownloader(logger)
	commandFactory := command.NewFactory(envRepository)
	fileManager := fileutil.NewFileManager()
	step := NewStep(inputParser, logger, pathModifier, pathProvider, downloader, commandFactory, fileManager)

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
