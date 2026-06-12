package graph

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheVersion = 1

const cacheFileName = "graph-v1.gob"

// Full graph rebuild runs when any cached source file mtime is newer than the cached value.
type cacheSnapshot struct {
	Nodes    map[string]*Node
	Forward  map[string][]Edge
	Backward map[string][]Edge
}

type cacheEntry struct {
	Version    int
	BuiltAt    time.Time
	FileMTimes map[string]int64
	Graph      cacheSnapshot
}

func cachePath(moduleRoot, cacheDir string) (string, error) {
	dir := cacheDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(moduleRoot, cacheDir)
	}
	return filepath.Join(dir, cacheFileName), nil
}

func Save(moduleRoot, cacheDir string, g *Graph, mtimes map[string]int64) error {
	path, err := cachePath(moduleRoot, cacheDir)
	if err != nil {
		return err
	}

	entry := cacheEntry{
		Version:    cacheVersion,
		BuiltAt:    time.Now().UTC(),
		FileMTimes: cloneMTimes(mtimes),
		Graph:      g.snapshot(),
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := gob.NewEncoder(file).Encode(entry); err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	return nil
}

func Load(ctx context.Context, moduleRoot, cacheDir string) (*Graph, map[string]int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	path, err := cachePath(moduleRoot, cacheDir)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("cache miss: %w", err)
		}
		return nil, nil, fmt.Errorf("open cache file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var entry cacheEntry
	if err := gob.NewDecoder(file).Decode(&entry); err != nil {
		return nil, nil, fmt.Errorf("decode cache: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if entry.Version != cacheVersion {
		return nil, nil, fmt.Errorf("cache miss: unsupported version %d", entry.Version)
	}

	return graphFromSnapshot(entry.Graph), cloneMTimes(entry.FileMTimes), nil
}

func IsValid(cachedMTimes, currentMTimes map[string]int64) bool {
	if len(cachedMTimes) == 0 {
		return false
	}
	if len(cachedMTimes) != len(currentMTimes) {
		return false
	}
	for path, cached := range cachedMTimes {
		current, ok := currentMTimes[path]
		if !ok || current > cached {
			return false
		}
	}
	return true
}

func CollectMTimes(ctx context.Context, moduleRoot string, files []string) (map[string]int64, error) {
	out := make(map[string]int64, len(files))
	for _, rel := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := rel
		if !filepath.IsAbs(path) {
			path = filepath.Join(moduleRoot, rel)
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", rel, err)
		}
		out[rel] = info.ModTime().UnixNano()
	}
	return out, nil
}

func cloneMTimes(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for path, value := range in {
		out[path] = value
	}
	return out
}
