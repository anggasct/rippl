package testmap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type profileEntry struct {
	total   int
	covered int
}

func parseCoverProfile(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("parse coverprofile: %w", err)
	}
	defer func() { _ = f.Close() }()

	stats := make(map[string]profileEntry)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		const goExt = ".go:"
		idx := strings.Index(line, goExt)
		if idx < 0 {
			continue
		}
		filePart := line[:idx+len(".go")]
		rest := strings.TrimSpace(line[idx+len(goExt):])
		fields := strings.Fields(rest)
		if len(fields) < 2 {
			continue
		}
		numStmts, err := strconv.Atoi(fields[len(fields)-2])
		if err != nil || numStmts <= 0 {
			continue
		}
		count, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			continue
		}
		key := normalizeProfilePath(filePart)
		entry := stats[key]
		entry.total += numStmts
		if count > 0 {
			entry.covered += numStmts
		}
		stats[key] = entry
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse coverprofile: %w", err)
	}

	out := make(map[string]float64, len(stats))
	for path, entry := range stats {
		if entry.total == 0 {
			continue
		}
		out[path] = float64(entry.covered) * 100 / float64(entry.total)
	}
	return out, nil
}

func normalizeProfilePath(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	if idx := strings.Index(path, "/internal/"); idx >= 0 {
		return path[idx+1:]
	}
	if idx := strings.Index(path, "/pkg/"); idx >= 0 {
		return "pkg/" + path[idx+len("/pkg/"):]
	}
	return path
}

func matchProfilePct(profiles map[string]float64, graphPath string) (float64, bool) {
	if pct, ok := profiles[graphPath]; ok {
		return pct, true
	}
	for profilePath, pct := range profiles {
		if strings.HasSuffix(profilePath, graphPath) || strings.HasSuffix(graphPath, profilePath) {
			return pct, true
		}
		if filepath.Base(profilePath) == filepath.Base(graphPath) {
			return pct, true
		}
	}
	return 0, false
}
