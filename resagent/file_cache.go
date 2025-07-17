package resagent

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/lamlv2305/sentinel/types"
)

var _ Cache = (*FileCache)(nil)

type FileCache struct {
	file   *os.File
	mu     *sync.RWMutex
	buffer map[string]types.Resource
}

func NewFileCache(filepath string) (*FileCache, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	// Seek to the beginning before reading
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	fc := &FileCache{
		file:   file,
		mu:     &sync.RWMutex{},
		buffer: make(map[string]types.Resource),
	}

	for _, line := range lines {
		var res types.Resource
		if err := res.UnmarshalText([]byte(line)); err != nil {
			return nil, err
		}

		fc.buffer[res.ResourceId] = res
	}

	return fc, nil
}

// Delete implements Cache.
func (f *FileCache) Delete(ctx context.Context, rid string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.buffer, rid)
}

// Get implements Cache.
func (f *FileCache) Get(ctx context.Context, rid string) *types.Resource {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if res, exists := f.buffer[rid]; exists {
		return &res
	}

	return nil
}

// List implements Cache.
func (f *FileCache) List(ctx context.Context) []types.Resource {
	f.mu.RLock()
	defer f.mu.RUnlock()

	resources := make([]types.Resource, 0, len(f.buffer))
	for _, res := range f.buffer {
		resources = append(resources, res)
	}

	return resources
}

// Set implements Cache.
func (f *FileCache) Set(ctx context.Context, data types.Resource) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.buffer[data.ResourceId] = data
	if err := f.flush(); err != nil {
		return err
	}
	return nil
}

func (f *FileCache) Start(ctx context.Context) {
	flushTicker := time.NewTicker(5 * time.Minute)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.mu.Lock()
			_ = f.flush()
			_ = f.file.Close()
			f.mu.Unlock()

			return

		case <-flushTicker.C:
			if err := f.flush(); err != nil {
				slog.Error("Failed to flush file cache", "error", err)
			}
		}
	}
}

func (f *FileCache) flush() error {
	items := make([]string, 0, len(f.buffer))

	f.mu.RLock()
	for _, res := range f.buffer {
		text, err := res.MarshalText()
		if err != nil {
			continue // Skip this resource if it cannot be marshaled
		}
		items = append(items, string(text))
	}
	f.mu.RUnlock()

	// Clear the file and write new content
	if err := f.file.Truncate(0); err != nil {
		return err
	}

	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := f.file.WriteString(items[0]); err != nil {
		return err
	}
	return nil
}
