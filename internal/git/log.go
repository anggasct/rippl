package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const commitRecordSep = "\x00COMMIT\x00"

// gitSince normalizes cfg.Risk.Since for git's --since flag.
// Values without a trailing " ago" get one appended (e.g. "12 months" → "12 months ago").
func gitSince(since string) string {
	since = strings.TrimSpace(since)
	if since == "" {
		return "12 months ago"
	}
	if strings.HasSuffix(since, " ago") {
		return since
	}
	return since + " ago"
}

func logArgs(since, path string) []string {
	return []string{
		"log", "--follow",
		"--format=" + commitRecordSep + "%H%x00%an%x00%at%x00%s%x00%b",
		"--since=" + gitSince(since),
		"--", path,
	}
}

func numstatArgs(since, path string) []string {
	return []string{
		"log", "--follow", "--numstat",
		"--pretty=format:" + commitRecordSep + "%H",
		"--since=" + gitSince(since),
		"--", path,
	}
}

type commitRecord struct {
	hash    string
	author  string
	when    time.Time
	subject string
	body    string
}

func parseCommitLog(out []byte) []commitRecord {
	text := string(out)
	if text == "" {
		return nil
	}
	parts := strings.Split(text, commitRecordSep)
	var commits []commitRecord
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, "\x00")
		if len(fields) < 4 {
			continue
		}
		ts, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}
		subject := fields[3]
		body := ""
		if len(fields) > 4 {
			body = strings.Join(fields[4:], "\x00")
		}
		commits = append(commits, commitRecord{
			hash:    fields[0],
			author:  fields[1],
			when:    time.Unix(ts, 0).UTC(),
			subject: subject,
			body:    body,
		})
	}
	return commits
}

func parseNumstatChurn(out []byte) int {
	churn := 0
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, commitRecordSep) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		added, errA := strconv.Atoi(fields[0])
		if errA != nil {
			continue
		}
		deleted, errD := strconv.Atoi(fields[1])
		if errD != nil {
			continue
		}
		churn += added + deleted
	}
	return churn
}

func isBugFix(subject, body string, patterns []*regexp.Regexp) bool {
	text := subject + "\n" + body
	for _, p := range patterns {
		if p.MatchString(text) {
			return true
		}
	}
	return false
}

func compileBugFixPatterns(raw []string) ([]*regexp.Regexp, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]*regexp.Regexp, 0, len(raw))
	for _, s := range raw {
		re, err := regexp.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("compile bug-fix pattern %q: %w", s, err)
		}
		out = append(out, re)
	}
	return out, nil
}

func collectFileGitStats(
	ctx context.Context,
	runner Runner,
	repoRoot, relPath, since string,
	patterns []*regexp.Regexp,
) (FileGitStats, error) {
	if err := ctx.Err(); err != nil {
		return FileGitStats{}, err
	}

	logOut, err := runner.Run(ctx, repoRoot, logArgs(since, relPath)...)
	if err != nil {
		return FileGitStats{}, err
	}
	numOut, err := runner.Run(ctx, repoRoot, numstatArgs(since, relPath)...)
	if err != nil {
		return FileGitStats{}, err
	}

	commits := parseCommitLog(logOut)
	stats := FileGitStats{
		Path:        relPath,
		CommitCount: len(commits),
		Churn:       parseNumstatChurn(numOut),
	}

	authors := make(map[string]struct{})
	for _, c := range commits {
		authors[c.author] = struct{}{}
		if stats.LastModified.IsZero() || c.when.After(stats.LastModified) {
			stats.LastModified = c.when
		}
		if isBugFix(c.subject, c.body, patterns) {
			stats.BugFixCount++
		}
	}
	stats.AuthorCount = len(authors)
	return stats, nil
}
