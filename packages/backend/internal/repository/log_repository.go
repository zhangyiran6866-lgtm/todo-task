package repository

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"todotask/backend/internal/model"
	"todotask/backend/pkg/config"
)

var (
	ErrLogNotFound = errors.New("log not found")
)

type LogFilter struct {
	Channel string
	Level   string
	Module  string
	Keyword string
	StartAt *time.Time
	EndAt   *time.Time
}

type LogRepository interface {
	List(ctx context.Context, filter LogFilter, offset int, limit int) ([]*model.LogEntry, int, error)
	FindByID(ctx context.Context, id string, channel string) (*model.LogEntry, error)
}

type logRepository struct {
	channelPaths map[string]string
}

func NewLogRepository(logCfg *config.LogConfig) LogRepository {
	paths := map[string]string{
		"app":   strings.TrimSpace(logCfg.AppPath),
		"error": strings.TrimSpace(logCfg.ErrorPath),
		"audit": strings.TrimSpace(logCfg.AuditPath),
	}
	return &logRepository{channelPaths: paths}
}

func (r *logRepository) List(ctx context.Context, filter LogFilter, offset int, limit int) ([]*model.LogEntry, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	entries, err := r.loadEntries(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Timestamp.Equal(entries[j].Timestamp) {
			return entries[i].ID > entries[j].ID
		}
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	total := len(entries)
	if offset >= total {
		return []*model.LogEntry{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return entries[offset:end], total, nil
}

func (r *logRepository) FindByID(ctx context.Context, id string, channel string) (*model.LogEntry, error) {
	filter := LogFilter{Channel: channel}
	entries, err := r.loadEntries(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.ID == id {
			return entry, nil
		}
	}
	return nil, ErrLogNotFound
}

func (r *logRepository) loadEntries(ctx context.Context, filter LogFilter) ([]*model.LogEntry, error) {
	channels := r.resolveChannels(filter.Channel)
	entries := make([]*model.LogEntry, 0, 256)

	for _, channel := range channels {
		path := r.channelPaths[channel]
		if path == "" {
			continue
		}

		files, err := discoverLogFiles(path)
		if err != nil {
			return nil, err
		}

		for _, filePath := range files {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			fileEntries, err := readLogFile(ctx, channel, filePath, filter)
			if err != nil {
				return nil, err
			}
			entries = append(entries, fileEntries...)
		}
	}

	return entries, nil
}

func (r *logRepository) resolveChannels(channel string) []string {
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel != "" {
		if _, ok := r.channelPaths[channel]; ok {
			return []string{channel}
		}
		return []string{}
	}

	return []string{"app", "error", "audit"}
}

func discoverLogFiles(basePath string) ([]string, error) {
	ext := filepath.Ext(basePath)
	if ext == "" {
		ext = ".log"
	}
	prefix := strings.TrimSuffix(filepath.Base(basePath), filepath.Ext(basePath))
	if prefix == "" {
		prefix = filepath.Base(basePath)
	}

	dir := filepath.Dir(basePath)
	if dir == "." {
		dir = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read log dir failed: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, prefix+"-") {
			continue
		}
		if !(strings.HasSuffix(name, ext) || strings.HasSuffix(name, ext+".gz")) {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	return files, nil
}

func readLogFile(ctx context.Context, channel, filePath string, filter LogFilter) ([]*model.LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open log file failed: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("open gzip log failed: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	result := make([]*model.LogEntry, 0, 128)
	lineNo := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		lineNo++
		line := scanner.Bytes()
		entry, ok := parseLogLine(channel, line, lineNo)
		if !ok {
			continue
		}
		if !matchFilter(entry, filter) {
			continue
		}
		result = append(result, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan log file failed: %w", err)
	}

	return result, nil
}

func parseLogLine(channel string, line []byte, lineNo int) (*model.LogEntry, bool) {
	raw := make(map[string]interface{})
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil, false
	}

	ts := parseTime(raw["timestamp"])
	if ts.IsZero() {
		return nil, false
	}

	recordID := computeLogID(channel, lineNo, line)
	return &model.LogEntry{
		ID:         recordID,
		Channel:    channel,
		Timestamp:  ts,
		Level:      strings.ToLower(stringValue(raw["level"])),
		Module:     stringValue(raw["module"]),
		Action:     stringValue(raw["action"]),
		Message:    stringValue(raw["message"]),
		RequestID:  stringValue(raw["request_id"]),
		UserID:     stringValue(raw["user_id"]),
		Method:     stringValue(raw["method"]),
		Path:       stringValue(raw["path"]),
		Route:      stringValue(raw["route"]),
		ClientIP:   stringValue(raw["client_ip"]),
		StatusCode: int(numberValue(raw["status_code"])),
		LatencyMS:  int64(numberValue(raw["latency_ms"])),
		Error:      stringValue(raw["error"]),
		Raw:        raw,
	}, true
}

func matchFilter(entry *model.LogEntry, filter LogFilter) bool {
	if filter.Level != "" && entry.Level != strings.ToLower(filter.Level) {
		return false
	}
	if filter.Module != "" && strings.ToLower(entry.Module) != strings.ToLower(filter.Module) {
		return false
	}
	if filter.StartAt != nil && entry.Timestamp.Before(*filter.StartAt) {
		return false
	}
	if filter.EndAt != nil && entry.Timestamp.After(*filter.EndAt) {
		return false
	}

	keyword := strings.TrimSpace(strings.ToLower(filter.Keyword))
	if keyword == "" {
		return true
	}

	haystack := strings.ToLower(strings.Join([]string{
		entry.Message,
		entry.Error,
		entry.Path,
		entry.Route,
		entry.RequestID,
		entry.Module,
		entry.Action,
		entry.UserID,
		entry.Level,
		strconv.Itoa(entry.StatusCode),
	}, " "))
	return strings.Contains(haystack, keyword)
}

func parseTime(v interface{}) time.Time {
	s := stringValue(v)
	if s == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05-0700",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func computeLogID(channel string, lineNo int, line []byte) string {
	h := sha1.New() //nolint:gosec
	h.Write([]byte(channel))
	h.Write([]byte("|"))
	h.Write([]byte(strconv.Itoa(lineNo)))
	h.Write([]byte("|"))
	h.Write(line)
	return hex.EncodeToString(h.Sum(nil))
}

func stringValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case fmt.Stringer:
		return strings.TrimSpace(val.String())
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func numberValue(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case json.Number:
		n, _ := val.Float64()
		return n
	default:
		return 0
	}
}
