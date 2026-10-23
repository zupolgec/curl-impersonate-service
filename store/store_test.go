package store

import (
	"path/filepath"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestTokenLifecycle(t *testing.T) {
	s := openTestStore(t)

	tok, err := s.CreateToken("ci")
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if len(tok.Token) < 32 {
		t.Fatalf("token too short: %q", tok.Token)
	}

	name, ok := s.ValidateToken(tok.Token)
	if !ok || name != "ci" {
		t.Fatalf("ValidateToken = %q,%v want ci,true", name, ok)
	}

	if _, ok := s.ValidateToken("bogus"); ok {
		t.Fatal("bogus token validated")
	}

	if err := s.SetTokenEnabled(tok.ID, false); err != nil {
		t.Fatalf("SetTokenEnabled: %v", err)
	}
	if _, ok := s.ValidateToken(tok.Token); ok {
		t.Fatal("disabled token validated")
	}

	if err := s.DeleteToken(tok.ID); err != nil {
		t.Fatalf("DeleteToken: %v", err)
	}
	toks, err := s.ListTokens()
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(toks) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(toks))
	}
}

func TestSeedTokenIdempotent(t *testing.T) {
	s := openTestStore(t)
	if err := s.SeedToken("legacy", "fixed-value"); err != nil {
		t.Fatalf("SeedToken: %v", err)
	}
	if err := s.SeedToken("legacy", "fixed-value"); err != nil {
		t.Fatalf("SeedToken (2nd): %v", err)
	}
	toks, _ := s.ListTokens()
	if len(toks) != 1 {
		t.Fatalf("expected 1 token after duplicate seed, got %d", len(toks))
	}
}

func TestSettings(t *testing.T) {
	s := openTestStore(t)
	if got := s.GetSetting("cors", "default"); got != "default" {
		t.Fatalf("GetSetting default = %q", got)
	}
	if err := s.SetSetting("cors", "https://a.com"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := s.SetSetting("cors", "https://b.com"); err != nil {
		t.Fatalf("SetSetting update: %v", err)
	}
	if got := s.GetSetting("cors", "default"); got != "https://b.com" {
		t.Fatalf("GetSetting = %q want https://b.com", got)
	}
}

func TestLogsAndPurge(t *testing.T) {
	s := openTestStore(t)
	for range 3 {
		if err := s.AddLog(LogEntry{TokenName: "ci", Browser: "chrome116", Method: "GET", TargetHost: "example.com", StatusCode: 200, Success: true, DurationMs: 12}); err != nil {
			t.Fatalf("AddLog: %v", err)
		}
	}
	logs, err := s.ListLogs(10)
	if err != nil {
		t.Fatalf("ListLogs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}

	// Nothing older than 1h yet.
	if n, _ := s.PurgeLogsOlderThan(time.Hour); n != 0 {
		t.Fatalf("expected 0 purged, got %d", n)
	}
	// Everything older than 0 → all purged.
	if n, _ := s.PurgeLogsOlderThan(-time.Second); n != 3 {
		t.Fatalf("expected 3 purged, got %d", n)
	}
}

func TestDailyStatsSurvivePurge(t *testing.T) {
	s := openTestStore(t)
	add := func(tok string, ok bool, ms int64) {
		t.Helper()
		if err := s.AddLog(LogEntry{TokenName: tok, Browser: "chrome152", Method: "GET", TargetHost: "example.com", Success: ok, DurationMs: ms}); err != nil {
			t.Fatalf("AddLog: %v", err)
		}
	}
	add("a", true, 100)
	add("a", false, 300)
	add("b", true, 200)

	// Purging the logs must not touch the history.
	if _, err := s.PurgeLogsOlderThan(-time.Second); err != nil {
		t.Fatal(err)
	}

	tot, err := s.TotalStats()
	if err != nil {
		t.Fatal(err)
	}
	if tot.Requests != 3 || tot.Success != 2 || tot.Failed != 1 || tot.AvgMs() != 200 {
		t.Fatalf("totals = %+v", tot)
	}
	if tot.Since.IsZero() {
		t.Fatal("Since should be the first tracked day")
	}

	days, err := s.DailyStats(30)
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().UTC().Format("2006-01-02")
	if len(days) != 1 || days[0].Day != today || days[0].Requests != 3 {
		t.Fatalf("days = %+v", days)
	}
}

func TestDailyStatsBackfilledFromLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a database from before the history existed: logs only.
	two := time.Now().Add(-48 * time.Hour).Unix()
	for _, ts := range []int64{two, two, time.Now().Unix()} {
		if _, err := s.db.Exec(`INSERT INTO usage_logs (ts, token_name, browser, method, target_host, status_code, success, duration_ms, error_type) VALUES (?, 'a', 'chrome150', 'GET', 'x.com', 200, 1, 50, '')`, ts); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.db.Exec(`DELETE FROM daily_stats`); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()

	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	days, _ := s.DailyStats(30)
	if len(days) != 2 || days[0].Requests != 1 || days[1].Requests != 2 {
		t.Fatalf("backfill days = %+v", days)
	}

	// Reopening must not double count.
	_ = s.Close()
	s, _ = Open(path)
	if tot, _ := s.TotalStats(); tot.Requests != 3 {
		t.Fatalf("after reopen requests = %d, want 3", tot.Requests)
	}
}
