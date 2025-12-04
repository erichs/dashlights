package signals

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InsecureCurlPipeSignal detects insecure `curl | bash` or `curl | sh` usage
// in the most recent shell history entries.
//
// It approximates: tail -3 ~/.$(basename $SHELL)_history
// and scans only the last few lines for curl pipelines into bash/sh.
type InsecureCurlPipeSignal struct {
	reason string
}

var curlPipePattern = regexp.MustCompile(`(?i)\bcurl\b[^|]*\|\s*(sudo\s+)?(bash|sh)\b`)

func NewInsecureCurlPipeSignal() Signal {
	return &InsecureCurlPipeSignal{}
}

func (s *InsecureCurlPipeSignal) Name() string {
	return "Insecure Curl Pipe"
}

func (s *InsecureCurlPipeSignal) Emoji() string {
	return "⚠️"
}

func (s *InsecureCurlPipeSignal) Diagnostic() string {
	if s.reason != "" {
		return s.reason
	}
	return "Recent shell history contains insecure curl | bash or curl | sh execution"
}

func (s *InsecureCurlPipeSignal) Remediation() string {
	return "Avoid piping curl directly into bash/sh; use checksum.sh or download-inspect-execute instead"
}

func (s *InsecureCurlPipeSignal) Check(ctx context.Context) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return false
	}

	shellName := filepath.Base(shellPath)
	if shellName == "" {
		return false
	}

	historyFileName := "." + shellName + "_history"

	// Validate the history filename to avoid directory traversal (G304)
	if strings.ContainsAny(historyFileName, `/\\`) || strings.Contains(historyFileName, "..") {
		return false
	}

	historyFile := filepath.Join(homeDir, historyFileName)
	historyFile = filepath.Clean(historyFile)

	// Ensure the resolved path remains within the user's home directory.
	if !strings.HasPrefix(historyFile, homeDir+string(filepath.Separator)) && historyFile != homeDir {
		return false
	}

	file, err := os.Open(historyFile)
	if err != nil {
		return false
	}
	defer file.Close()

	// For very large history files, seek close to the end so we only
	// scan roughly the last 128KB, which will still contain the final
	// few commands but keeps worst-case I/O bounded.
	const tailBytes int64 = 128 * 1024
	if info, err := file.Stat(); err == nil {
		if size := info.Size(); size > tailBytes {
			if _, err := file.Seek(size-tailBytes, io.SeekStart); err != nil {
				// If seeking fails for some reason, fall back to scanning
				// from the beginning rather than failing the signal.
				if _, err := file.Seek(0, io.SeekStart); err != nil {
					return false
				}
			}
		}
	}

	scanner := bufio.NewScanner(file)
	// Allow reasonably long history lines before Scanner bails out.
	const maxLineSize = 1024 * 1024
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	const maxLines = 3
	lastLines := make([]string, 0, maxLines)

	for scanner.Scan() {
		// Respect context cancellation for large history files
		select {
		case <-ctx.Done():
			return false
		default:
		}

		line := scanner.Text()
		lastLines = append(lastLines, line)
		if len(lastLines) > maxLines {
			lastLines = lastLines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}

	var matches []string
	for _, line := range lastLines {
		if curlPipePattern.MatchString(line) {
			matches = append(matches, strings.TrimSpace(line))
		}
	}

	if len(matches) == 0 {
		return false
	}

	if len(matches) == 1 {
		s.reason = "Recent shell history contains insecure curl | bash/sh: " + truncateHistoryLine(matches[0])
	} else {
		s.reason = "Recent shell history contains multiple insecure curl | bash/sh commands"
	}

	return true
}

// truncateHistoryLine limits diagnostic length while keeping the interesting prefix.
func truncateHistoryLine(line string) string {
	const maxLen = 120
	if len(line) <= maxLen {
		return line
	}
	return line[:maxLen] + "..."
}
