package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/filedownloader"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

func newTestStep() Step {
	logger := log.NewLogger()
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	return NewStep(
		inputParser,
		logger,
		pathutil.NewPathModifier(),
		pathutil.NewPathProvider(),
		filedownloader.NewDownloader(logger),
		command.NewFactory(envRepository),
		fileutil.NewFileManager(),
	)
}

func TestProcessConfig_ParsesInputs(t *testing.T) {
	t.Setenv("source", "https://example.com/file.apk")
	t.Setenv("destination", "out/file.apk")
	t.Setenv("file_permission", "600")

	cfg, err := newTestStep().ProcessConfig()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if cfg.Source != "https://example.com/file.apk" {
		t.Errorf("Source = %q, want %q", cfg.Source, "https://example.com/file.apk")
	}
	if !filepath.IsAbs(cfg.Destination) {
		t.Errorf("Destination = %q, want an absolute path", cfg.Destination)
	}
	if cfg.FilePermission == nil {
		t.Fatal("FilePermission = nil, want 0600")
	}
	if *cfg.FilePermission != 0o600 {
		t.Errorf("FilePermission = %#o, want %#o", *cfg.FilePermission, os.FileMode(0o600))
	}
}

func TestProcessConfig_OmittedPermissionIsNil(t *testing.T) {
	t.Setenv("source", "https://example.com/file.apk")
	t.Setenv("destination", "out/file.apk")
	t.Setenv("file_permission", "")

	cfg, err := newTestStep().ProcessConfig()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg.FilePermission != nil {
		t.Errorf("FilePermission = %v, want nil when not set", cfg.FilePermission)
	}
}

func TestProcessConfig_UseSudo(t *testing.T) {
	t.Setenv("source", "https://example.com/file.apk")
	t.Setenv("destination", "out/file.apk")

	for _, tc := range []struct {
		value string
		want  bool
	}{
		{value: "true", want: true},
		{value: "false", want: false},
	} {
		t.Setenv("use_sudo", tc.value)
		cfg, err := newTestStep().ProcessConfig()
		if err != nil {
			t.Fatalf("use_sudo=%q: unexpected error: %s", tc.value, err)
		}
		if cfg.UseSudo != tc.want {
			t.Errorf("use_sudo=%q: UseSudo = %v, want %v", tc.value, cfg.UseSudo, tc.want)
		}
	}
}

func TestProcessConfig_InvalidPermissionFails(t *testing.T) {
	t.Setenv("source", "https://example.com/file.apk")
	t.Setenv("destination", "out/file.apk")
	t.Setenv("file_permission", "not-octal")

	if _, err := newTestStep().ProcessConfig(); err == nil {
		t.Fatal("expected an error for an invalid file permission, got nil")
	}
}

func TestProcessConfig_MissingRequiredInputFails(t *testing.T) {
	t.Setenv("source", "")
	t.Setenv("destination", "out/file.apk")

	if _, err := newTestStep().ProcessConfig(); err == nil {
		t.Fatal("expected an error for a missing required input, got nil")
	}
}
