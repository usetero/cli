//go:build correctness

package extension_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/powersync/db/dbtest"
	"github.com/usetero/cli/internal/powersync/extension"
	"github.com/usetero/cli/internal/sqlite"
)

const (
	envPowerSyncFixturePath     = "TERO_POWERSYNC_FIXTURE_PATH"
	envPowerSyncFixtureMaxLines = "TERO_POWERSYNC_FIXTURE_MAX_LINES"
	defaultFixturePath          = "testdata/dev-sanitized.ndjson"
)

type replayDigest struct {
	BucketCount int64
	SumLastOp   int64
	SumTargetOp int64
	MaxLastOp   int64
	MaxTargetOp int64
	OplogCount  int64
	OplogMaxRow int64
}

func TestCorrectness_PowerSync_ControllerFixtureReplayDeterministic(t *testing.T) {
	fixturePath := os.Getenv(envPowerSyncFixturePath)
	if fixturePath == "" {
		fixturePath = defaultFixturePath
	}

	maxLines, err := fixtureMaxLinesFromEnv()
	if err != nil {
		t.Fatalf("invalid %s: %v", envPowerSyncFixtureMaxLines, err)
	}

	var baseline replayDigest
	var baselineLines int
	for i := range 2 {
		database := dbtest.OpenTestDB(t)
		lines, err := replayFixture(context.Background(), database, fixturePath, maxLines)
		if err != nil {
			t.Fatalf("replay run %d failed: %v", i+1, err)
		}
		if lines == 0 {
			t.Fatalf("replay run %d applied zero lines", i+1)
		}

		digest, err := snapshotDigest(context.Background(), database)
		if err != nil {
			t.Fatalf("snapshot digest run %d: %v", i+1, err)
		}

		if i == 0 {
			baseline = digest
			baselineLines = lines
			continue
		}

		if lines != baselineLines {
			t.Fatalf("line count mismatch: run1=%d run2=%d", baselineLines, lines)
		}
		if digest != baseline {
			t.Fatalf("digest mismatch:\nrun1=%+v\nrun2=%+v", baseline, digest)
		}
	}
}

func replayFixture(ctx context.Context, database sqlite.DB, fixturePath string, maxLines int) (int, error) {
	controller := extension.NewController(database)
	defer controller.Close()

	if _, err := controller.Start(ctx, extension.StartRequest{IncludeDefaults: true}); err != nil {
		return 0, fmt.Errorf("start: %w", err)
	}
	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEstablished); err != nil {
		return 0, fmt.Errorf("notify connection established: %w", err)
	}

	f, err := os.Open(fixturePath)
	if err != nil {
		return 0, fmt.Errorf("open fixture: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)

	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lineNo++
		if _, err := controller.SendTextLine(ctx, line); err != nil {
			return lineNo, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if maxLines > 0 && lineNo >= maxLines {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return lineNo, fmt.Errorf("read fixture: %w", err)
	}

	if _, err := controller.NotifyConnection(ctx, extension.ConnectionEnded); err != nil && !isNoActiveIterationErr(err) {
		return lineNo, fmt.Errorf("notify connection ended: %w", err)
	}
	return lineNo, nil
}

func snapshotDigest(ctx context.Context, database sqlite.DB) (replayDigest, error) {
	var d replayDigest
	if err := database.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(last_op), 0),
			COALESCE(SUM(target_op), 0),
			COALESCE(MAX(last_op), 0),
			COALESCE(MAX(target_op), 0)
		FROM ps_buckets
	`).Scan(&d.BucketCount, &d.SumLastOp, &d.SumTargetOp, &d.MaxLastOp, &d.MaxTargetOp); err != nil {
		return replayDigest{}, err
	}
	if err := database.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(MAX(rowid), 0)
		FROM ps_oplog
	`).Scan(&d.OplogCount, &d.OplogMaxRow); err != nil {
		return replayDigest{}, err
	}
	return d, nil
}

func fixtureMaxLinesFromEnv() (int, error) {
	raw := os.Getenv(envPowerSyncFixtureMaxLines)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be > 0")
	}
	return n, nil
}

func isNoActiveIterationErr(err error) bool {
	return strings.Contains(err.Error(), "No iteration is active")
}
