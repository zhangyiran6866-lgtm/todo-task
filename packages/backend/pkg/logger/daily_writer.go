package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const dayLayout = "2006-01-02"

type dailyRotateWriter struct {
	mu            sync.Mutex
	dir           string
	prefix        string
	ext           string
	currentDay    string
	file          *os.File
	retentionDays int
	compress      bool
}

func newDailyRotateWriter(path string, retentionDays int, compress bool) (*dailyRotateWriter, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	if retentionDays <= 0 {
		retentionDays = 7
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	prefix := strings.TrimSuffix(base, ext)
	if ext == "" {
		ext = ".log"
	}
	if prefix == "" {
		return nil, fmt.Errorf("invalid path prefix: %s", path)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir failed: %w", err)
	}

	writer := &dailyRotateWriter{
		dir:           dir,
		prefix:        prefix,
		ext:           ext,
		retentionDays: retentionDays,
		compress:      compress,
	}

	if err := writer.rotateIfNeededLocked(time.Now()); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *dailyRotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeededLocked(time.Now()); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *dailyRotateWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *dailyRotateWriter) rotateIfNeededLocked(now time.Time) error {
	currentDay := now.Format(dayLayout)
	if w.file != nil && currentDay == w.currentDay {
		return nil
	}

	if w.file != nil {
		_ = w.file.Sync()
		_ = w.file.Close()
	}

	filePath := filepath.Join(w.dir, fmt.Sprintf("%s-%s%s", w.prefix, currentDay, w.ext))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}

	w.file = file
	w.currentDay = currentDay

	if err := w.cleanupLocked(now); err != nil {
		return fmt.Errorf("cleanup log file failed: %w", err)
	}

	return nil
}

func (w *dailyRotateWriter) cleanupLocked(now time.Time) error {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return err
	}

	cutoff := dateOnly(now).AddDate(0, 0, -w.retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		day, isCompressed, ok := w.parseDay(name)
		if !ok {
			continue
		}

		filePath := filepath.Join(w.dir, name)

		// Optional compression for historical logs.
		if w.compress && !isCompressed && day.Before(dateOnly(now)) {
			if err := gzipFile(filePath); err == nil {
				_ = os.Remove(filePath)
				name = name + ".gz"
				isCompressed = true
			}
		}

		if day.Before(cutoff) {
			_ = os.Remove(filepath.Join(w.dir, name))
		}
	}

	return nil
}

func (w *dailyRotateWriter) parseDay(name string) (time.Time, bool, bool) {
	if !strings.HasPrefix(name, w.prefix+"-") {
		return time.Time{}, false, false
	}

	if strings.HasSuffix(name, w.ext) {
		raw := strings.TrimSuffix(strings.TrimPrefix(name, w.prefix+"-"), w.ext)
		day, err := time.Parse(dayLayout, raw)
		return day, false, err == nil
	}

	gzExt := w.ext + ".gz"
	if strings.HasSuffix(name, gzExt) {
		raw := strings.TrimSuffix(strings.TrimPrefix(name, w.prefix+"-"), gzExt)
		day, err := time.Parse(dayLayout, raw)
		return day, true, err == nil
	}

	return time.Time{}, false, false
}

func dateOnly(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func gzipFile(path string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path + ".gz")
	if err != nil {
		return err
	}
	defer dst.Close()

	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, src); err != nil {
		return err
	}
	return nil
}
