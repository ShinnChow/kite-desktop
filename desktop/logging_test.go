package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

func TestDesktopMainLogPath(t *testing.T) {
	paths := desktopPaths{
		LogsDir: filepath.Join("/tmp", "kite", "logs"),
	}

	got := desktopMainLogPath(paths)
	want := filepath.Join(paths.LogsDir, desktopMainLogFileName)
	if got != want {
		t.Fatalf("desktopMainLogPath() = %q, want %q", got, want)
	}
}

func TestDesktopRotatingLogWriterRotatesBySize(t *testing.T) {
	paths := desktopPaths{
		LogsDir: t.TempDir(),
	}

	writer, err := newDesktopRotatingLogWriterWithClock(paths, 32, 7*24*time.Hour, func() time.Time {
		return time.Date(2026, 4, 21, 15, 4, 5, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("newDesktopRotatingLogWriterWithClock() error = %v", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	if _, err := writer.Write([]byte("12345678901234567890\n")); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	if _, err := writer.Write([]byte("abcdefghijklmnopqrstuvwxyz\n")); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}

	entries, err := os.ReadDir(paths.LogsDir)
	if err != nil {
		t.Fatalf("os.ReadDir() error = %v", err)
	}

	var archiveName string
	for _, entry := range entries {
		if isDesktopLogArchiveName(entry.Name()) {
			archiveName = entry.Name()
			break
		}
	}
	if archiveName == "" {
		t.Fatalf("expected rotated archive in %q", paths.LogsDir)
	}

	archiveContent, err := os.ReadFile(filepath.Join(paths.LogsDir, archiveName))
	if err != nil {
		t.Fatalf("os.ReadFile(archive) error = %v", err)
	}
	if string(archiveContent) != "12345678901234567890\n" {
		t.Fatalf("unexpected archive content: %q", string(archiveContent))
	}

	activeContent, err := os.ReadFile(desktopMainLogPath(paths))
	if err != nil {
		t.Fatalf("os.ReadFile(active) error = %v", err)
	}
	if string(activeContent) != "abcdefghijklmnopqrstuvwxyz\n" {
		t.Fatalf("unexpected active content: %q", string(activeContent))
	}
}

func TestDesktopRotatingLogWriterCleansOldArchives(t *testing.T) {
	paths := desktopPaths{
		LogsDir: t.TempDir(),
	}
	oldArchive := filepath.Join(paths.LogsDir, "desktop-20260401-010101.000.log")
	recentArchive := filepath.Join(paths.LogsDir, "desktop-20260420-010101.000.log")
	activeLog := desktopMainLogPath(paths)

	for _, file := range []string{oldArchive, recentArchive, activeLog} {
		if err := os.WriteFile(file, []byte("payload"), 0o644); err != nil {
			t.Fatalf("os.WriteFile(%q) error = %v", file, err)
		}
	}
	oldTime := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	recentTime := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(oldArchive, oldTime, oldTime); err != nil {
		t.Fatalf("os.Chtimes(oldArchive) error = %v", err)
	}
	if err := os.Chtimes(recentArchive, recentTime, recentTime); err != nil {
		t.Fatalf("os.Chtimes(recentArchive) error = %v", err)
	}

	writer, err := newDesktopRotatingLogWriterWithClock(paths, 1024, 7*24*time.Hour, func() time.Time {
		return time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("newDesktopRotatingLogWriterWithClock() error = %v", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	if _, err := os.Stat(oldArchive); !os.IsNotExist(err) {
		t.Fatalf("expected old archive removed, stat err = %v", err)
	}
	if _, err := os.Stat(recentArchive); err != nil {
		t.Fatalf("expected recent archive kept, stat err = %v", err)
	}
	if _, err := os.Stat(activeLog); err != nil {
		t.Fatalf("expected active log kept, stat err = %v", err)
	}
}

func TestSetupDesktopLoggingWritesToDesktopLogFile(t *testing.T) {
	paths := desktopPaths{
		LogsDir: t.TempDir(),
	}

	cleanup, err := setupDesktopLogging(paths)
	if err != nil {
		t.Fatalf("setupDesktopLogging() error = %v", err)
	}

	log.Print("standard log message")
	_, _ = fmt.Fprintln(os.Stdout, "stdout message")
	klog.Info("klog message")
	klog.Flush()

	cleanup()

	content, err := os.ReadFile(desktopMainLogPath(paths))
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"desktop logging initialised:",
		"standard log message",
		"stdout message",
		"klog message",
		"retention=168h0m0s",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("desktop log file missing %q in %q", want, text)
		}
	}
}
