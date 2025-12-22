package metrics

import (
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	if collector == nil {
		t.Fatal("NewCollector() returned nil")
	}

	uptime, total, success, failed, _, browsers := collector.GetMetrics()

	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}

	if success != 0 {
		t.Errorf("success = %d, want 0", success)
	}

	if failed != 0 {
		t.Errorf("failed = %d, want 0", failed)
	}

	if uptime < 0 {
		t.Errorf("uptime = %d, want >= 0", uptime)
	}

	if len(browsers) != 0 {
		t.Errorf("browsers map should be empty, got %d items", len(browsers))
	}
}

func TestRecordRequest_Success(t *testing.T) {
	collector := NewCollector()

	collector.RecordRequest("chrome116", true, 100*time.Millisecond)

	_, total, success, failed, avgDuration, browsers := collector.GetMetrics()

	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}

	if success != 1 {
		t.Errorf("success = %d, want 1", success)
	}

	if failed != 0 {
		t.Errorf("failed = %d, want 0", failed)
	}

	if avgDuration <= 0 {
		t.Errorf("avgDuration = %f, want > 0", avgDuration)
	}

	if browsers["chrome116"] != 1 {
		t.Errorf("browsers[chrome116] = %d, want 1", browsers["chrome116"])
	}
}

func TestRecordRequest_Failed(t *testing.T) {
	collector := NewCollector()

	collector.RecordRequest("ff109", false, 50*time.Millisecond)

	_, total, success, failed, _, browsers := collector.GetMetrics()

	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}

	if success != 0 {
		t.Errorf("success = %d, want 0", success)
	}

	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}

	if browsers["ff109"] != 1 {
		t.Errorf("browsers[ff109] = %d, want 1", browsers["ff109"])
	}
}

func TestRecordRequest_Multiple(t *testing.T) {
	collector := NewCollector()

	collector.RecordRequest("chrome116", true, 100*time.Millisecond)
	collector.RecordRequest("chrome116", true, 200*time.Millisecond)
	collector.RecordRequest("ff109", false, 150*time.Millisecond)

	_, total, success, failed, avgDuration, browsers := collector.GetMetrics()

	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}

	if success != 2 {
		t.Errorf("success = %d, want 2", success)
	}

	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}

	expectedAvg := (100.0 + 200.0 + 150.0) / 3.0
	if avgDuration < expectedAvg-1 || avgDuration > expectedAvg+1 {
		t.Errorf("avgDuration = %f, want ~%f", avgDuration, expectedAvg)
	}

	if browsers["chrome116"] != 2 {
		t.Errorf("browsers[chrome116] = %d, want 2", browsers["chrome116"])
	}

	if browsers["ff109"] != 1 {
		t.Errorf("browsers[ff109] = %d, want 1", browsers["ff109"])
	}
}
