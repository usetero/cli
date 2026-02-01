package powersync_test

import (
	"testing"

	"github.com/usetero/cli/internal/powersync"
)

func TestDownloadProgress_TotalProgress(t *testing.T) {
	t.Parallel()

	t.Run("returns zero for nil", func(t *testing.T) {
		t.Parallel()

		var d *powersync.DownloadProgress
		downloaded, total := d.TotalProgress()

		if downloaded != 0 || total != 0 {
			t.Errorf("TotalProgress() = (%d, %d), want (0, 0)", downloaded, total)
		}
	})

	t.Run("returns zero for empty buckets", func(t *testing.T) {
		t.Parallel()

		d := &powersync.DownloadProgress{
			Buckets: map[string]powersync.BucketProgress{},
		}
		downloaded, total := d.TotalProgress()

		if downloaded != 0 || total != 0 {
			t.Errorf("TotalProgress() = (%d, %d), want (0, 0)", downloaded, total)
		}
	})

	t.Run("sums single bucket", func(t *testing.T) {
		t.Parallel()

		d := &powersync.DownloadProgress{
			Buckets: map[string]powersync.BucketProgress{
				"bucket1": {SinceLast: 5, TargetCount: 20},
			},
		}
		downloaded, total := d.TotalProgress()

		if downloaded != 5 || total != 20 {
			t.Errorf("TotalProgress() = (%d, %d), want (5, 20)", downloaded, total)
		}
	})

	t.Run("sums multiple buckets", func(t *testing.T) {
		t.Parallel()

		d := &powersync.DownloadProgress{
			Buckets: map[string]powersync.BucketProgress{
				"bucket1": {SinceLast: 5, TargetCount: 20},
				"bucket2": {SinceLast: 10, TargetCount: 30},
				"bucket3": {SinceLast: 3, TargetCount: 10},
			},
		}
		downloaded, total := d.TotalProgress()

		if downloaded != 18 || total != 60 {
			t.Errorf("TotalProgress() = (%d, %d), want (18, 60)", downloaded, total)
		}
	})
}
