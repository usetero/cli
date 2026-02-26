package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/log"
)

var digitsOnly = regexp.MustCompile(`^\d+$`)

var preserveStringByKey = map[string]struct{}{
	"op":               {},
	"op_id":            {},
	"after":            {},
	"next_after":       {},
	"last_op_id":       {},
	"write_checkpoint": {},
}

func newInternalPowerSyncSanitizeFixtureCmd(scope log.Scope) *cobra.Command {
	var (
		input    string
		output   string
		maxLines int
	)

	cmd := &cobra.Command{
		Use:   "sanitize-fixture",
		Short: "Sanitize raw PowerSync NDJSON into commit-safe fixture data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input == "" {
				return fmt.Errorf("--input is required")
			}
			if output == "" {
				return fmt.Errorf("--output is required")
			}
			if maxLines < 0 {
				return fmt.Errorf("--max-lines must be >= 0")
			}

			if !filepath.IsAbs(input) {
				input = filepath.Clean(input)
			}
			if !filepath.IsAbs(output) {
				output = filepath.Clean(output)
			}

			lines, err := sanitizeFixtureFile(input, output, maxLines)
			if err != nil {
				return err
			}
			scope.Info("sanitized powersync fixture", "input", input, "output", output, "lines", lines)
			fmt.Printf("Sanitized %d lines\nInput: %s\nOutput: %s\n", lines, input, output)
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "Path to raw NDJSON fixture (required)")
	cmd.Flags().StringVar(&output, "output", "", "Path to sanitized NDJSON fixture (required)")
	cmd.Flags().IntVar(&maxLines, "max-lines", 0, "Maximum number of lines to sanitize (0 = all)")

	return cmd
}

func sanitizeFixtureFile(inputPath, outputPath string, maxLines int) (int, error) {
	in, err := os.Open(inputPath)
	if err != nil {
		return 0, fmt.Errorf("open input fixture: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, fmt.Errorf("create output directory: %w", err)
	}
	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open output fixture: %w", err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	s := newSanitizer()
	written := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		sanitized, err := s.sanitizeLine(line)
		if err != nil {
			return written, fmt.Errorf("sanitize line %d: %w", written+1, err)
		}
		if sanitized == "" {
			continue
		}
		if _, err := writer.WriteString(sanitized); err != nil {
			return written, fmt.Errorf("write output line %d: %w", written+1, err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return written, fmt.Errorf("write newline %d: %w", written+1, err)
		}

		written++
		if maxLines > 0 && written >= maxLines {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return written, fmt.Errorf("read input fixture: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return written, fmt.Errorf("flush output fixture: %w", err)
	}
	return written, nil
}

type fixtureSanitizer struct {
	cache map[string]string
}

func newSanitizer() *fixtureSanitizer {
	return &fixtureSanitizer{cache: make(map[string]string)}
}

func (s *fixtureSanitizer) sanitizeLine(line string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(line), &v); err != nil {
		return "", fmt.Errorf("invalid json: %w", err)
	}
	if shouldDropReplayLine(v) {
		return "", nil
	}
	clean := s.sanitizeValue("", v)
	out, err := json.Marshal(clean)
	if err != nil {
		return "", fmt.Errorf("marshal sanitized line: %w", err)
	}
	return string(out), nil
}

func shouldDropReplayLine(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	_, hasData := m["data"]
	_, hasCheckpoint := m["checkpoint"]
	_, hasCheckpointComplete := m["checkpoint_complete"]
	return !hasData && !hasCheckpoint && !hasCheckpointComplete
}

func (s *fixtureSanitizer) sanitizeValue(key string, value any) any {
	switch v := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(v))
		for _, k := range keys {
			out[k] = s.sanitizeValue(k, v[k])
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = s.sanitizeValue("", v[i])
		}
		return out
	case string:
		return s.sanitizeString(key, v)
	default:
		return value
	}
}

func (s *fixtureSanitizer) sanitizeString(key, raw string) string {
	if raw == "" {
		return raw
	}
	if _, ok := preserveStringByKey[key]; ok {
		return raw
	}
	if digitsOnly.MatchString(raw) {
		return raw
	}

	if maybeJSON(raw) {
		var nested any
		if err := json.Unmarshal([]byte(raw), &nested); err == nil {
			nestedClean := s.sanitizeValue("", nested)
			if out, err := json.Marshal(nestedClean); err == nil {
				return string(out)
			}
		}
	}

	if token, ok := s.cache[raw]; ok {
		return token
	}

	sum := sha256.Sum256([]byte(raw))
	token := "redacted_" + hex.EncodeToString(sum[:6])
	s.cache[raw] = token
	return token
}

func maybeJSON(s string) bool {
	if len(s) < 2 {
		return false
	}
	first := s[0]
	last := s[len(s)-1]
	return (first == '{' && last == '}') || (first == '[' && last == ']')
}
