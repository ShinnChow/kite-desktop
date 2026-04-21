package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

const (
	desktopMainLogFileName            = "desktop.log"
	desktopMainLogArchivePrefix       = "desktop-"
	desktopMainLogArchiveSuffix       = ".log"
	desktopLogMaxSizeBytes      int64 = 10 * 1024 * 1024
	desktopLogRetention               = 7 * 24 * time.Hour
)

func desktopMainLogPath(paths desktopPaths) string {
	return filepath.Join(paths.LogsDir, desktopMainLogFileName)
}

type desktopRotatingLogWriter struct {
	dir          string
	activePath   string
	maxSizeBytes int64
	retention    time.Duration
	now          func() time.Time
	mu           sync.Mutex
	file         *os.File
	currentSize  int64
}

func newDesktopRotatingLogWriter(paths desktopPaths) (*desktopRotatingLogWriter, error) {
	return newDesktopRotatingLogWriterWithOptions(paths, desktopLogMaxSizeBytes, desktopLogRetention)
}

func newDesktopRotatingLogWriterWithOptions(paths desktopPaths, maxSizeBytes int64, retention time.Duration) (*desktopRotatingLogWriter, error) {
	return newDesktopRotatingLogWriterWithClock(paths, maxSizeBytes, retention, time.Now)
}

func newDesktopRotatingLogWriterWithClock(paths desktopPaths, maxSizeBytes int64, retention time.Duration, now func() time.Time) (*desktopRotatingLogWriter, error) {
	if maxSizeBytes <= 0 {
		return nil, fmt.Errorf("invalid desktop log max size: %d", maxSizeBytes)
	}
	if now == nil {
		now = time.Now
	}

	w := &desktopRotatingLogWriter{
		dir:          paths.LogsDir,
		activePath:   desktopMainLogPath(paths),
		maxSizeBytes: maxSizeBytes,
		retention:    retention,
		now:          now,
	}
	if err := w.openActiveFile(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *desktopRotatingLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openActiveFileLocked(); err != nil {
			return 0, err
		}
	}
	if w.currentSize > 0 && w.currentSize+int64(len(p)) > w.maxSizeBytes {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *desktopRotatingLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *desktopRotatingLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.currentSize = 0
	return err
}

func (w *desktopRotatingLogWriter) openActiveFile() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.openActiveFileLocked()
}

func (w *desktopRotatingLogWriter) openActiveFileLocked() error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("create desktop log dir %q: %w", w.dir, err)
	}
	if err := w.cleanupOldArchivesLocked(); err != nil {
		return err
	}

	info, err := os.Stat(w.activePath)
	if err == nil && info.Size() >= w.maxSizeBytes {
		if err := w.rotateExistingActiveFileLocked(); err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat desktop log file %q: %w", w.activePath, err)
	}

	file, err := os.OpenFile(w.activePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open desktop log file %q: %w", w.activePath, err)
	}
	info, err = file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("stat desktop log file %q: %w", w.activePath, err)
	}

	w.file = file
	w.currentSize = info.Size()
	return nil
}

func (w *desktopRotatingLogWriter) rotateLocked() error {
	if w.file != nil {
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("sync desktop log file %q: %w", w.activePath, err)
		}
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close desktop log file %q: %w", w.activePath, err)
		}
		w.file = nil
	}

	if err := w.rotateExistingActiveFileLocked(); err != nil {
		return err
	}
	if err := w.cleanupOldArchivesLocked(); err != nil {
		return err
	}

	file, err := os.OpenFile(w.activePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create rotated desktop log file %q: %w", w.activePath, err)
	}
	w.file = file
	w.currentSize = 0
	return nil
}

func (w *desktopRotatingLogWriter) rotateExistingActiveFileLocked() error {
	if _, err := os.Stat(w.activePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat desktop log file %q: %w", w.activePath, err)
	}

	archivePath := filepath.Join(w.dir, w.archiveFileName(w.now()))
	if err := os.Rename(w.activePath, archivePath); err != nil {
		return fmt.Errorf("rotate desktop log file %q -> %q: %w", w.activePath, archivePath, err)
	}
	return nil
}

func (w *desktopRotatingLogWriter) archiveFileName(now time.Time) string {
	return desktopMainLogArchivePrefix + now.Format("20060102-150405.000") + desktopMainLogArchiveSuffix
}

func (w *desktopRotatingLogWriter) cleanupOldArchivesLocked() error {
	if w.retention <= 0 {
		return nil
	}

	entries, err := os.ReadDir(w.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read desktop log dir %q: %w", w.dir, err)
	}

	cutoff := w.now().Add(-w.retention)
	for _, entry := range entries {
		if entry.IsDir() || !isDesktopLogArchiveName(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat desktop log archive %q: %w", entry.Name(), err)
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(w.dir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove desktop log archive %q: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

func isDesktopLogArchiveName(name string) bool {
	return strings.HasPrefix(name, desktopMainLogArchivePrefix) &&
		strings.HasSuffix(name, desktopMainLogArchiveSuffix) &&
		name != desktopMainLogFileName
}

func setupDesktopLogging(paths desktopPaths) (func(), error) {
	logWriter, err := newDesktopRotatingLogWriter(paths)
	if err != nil {
		return nil, err
	}

	prevStdout := os.Stdout
	prevStderr := os.Stderr
	prevLogWriter := log.Writer()
	prevGinWriter := gin.DefaultWriter
	prevGinErrorWriter := gin.DefaultErrorWriter

	restoreStdout, err := teeProcessStream(&os.Stdout, logWriter)
	if err != nil {
		_ = logWriter.Close()
		return nil, fmt.Errorf("redirect stdout to desktop log file: %w", err)
	}
	restoreStderr, err := teeProcessStream(&os.Stderr, logWriter)
	if err != nil {
		restoreStdout()
		_ = logWriter.Close()
		return nil, fmt.Errorf("redirect stderr to desktop log file: %w", err)
	}

	log.SetOutput(os.Stderr)
	gin.DefaultWriter = os.Stdout
	gin.DefaultErrorWriter = os.Stderr
	klog.SetOutput(os.Stderr)

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			log.SetOutput(prevLogWriter)
			gin.DefaultWriter = prevGinWriter
			gin.DefaultErrorWriter = prevGinErrorWriter
			restoreStdout()
			restoreStderr()
			os.Stdout = prevStdout
			os.Stderr = prevStderr
			klog.SetOutput(os.Stderr)
			klog.Flush()
			_ = logWriter.Sync()
			_ = logWriter.Close()
		})
	}

	log.Printf(
		"desktop logging initialised: %s (max_size=%d bytes retention=%s)",
		desktopMainLogPath(paths),
		desktopLogMaxSizeBytes,
		desktopLogRetention,
	)

	return cleanup, nil
}

func teeProcessStream(target **os.File, writer io.Writer) (func(), error) {
	original := *target
	reader, pipeWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	*target = pipeWriter

	var once sync.Once
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(io.MultiWriter(original, writer), reader)
		_ = reader.Close()
	}()

	return func() {
		once.Do(func() {
			*target = original
			_ = pipeWriter.Close()
			<-done
		})
	}, nil
}
