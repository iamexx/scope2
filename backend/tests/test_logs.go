package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamexx/scope2-dayz-api/internal/services"
)

func TestLogService(t *testing.T) {
	// Create temp directory for test logs
	tempDir := "/tmp/dayz-test"
	profilesDir := filepath.Join(tempDir, "profiles")
	os.MkdirAll(profilesDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create test log files
	consoleLog := `12:34:56 INFO: Server started
12:34:57 ERROR: Failed to load mods
12:34:58 WARN: Unknown command
12:34:59 INFO: Player connected
`
	os.WriteFile(filepath.Join(profilesDir, "DayZServer.RPT"), []byte(consoleLog), 0644)

	adminLog := `2024/01/15, 12:34:56 "PlayerOne" (id=76561198000000000) connected
2024/01/15, 12:35:00 "PlayerOne" (id=76561198000000000) pos=<1234.5, 678.9, 0.0> (chat) Hello world
2024/01/15, 12:36:00 "PlayerTwo" (id=76561198000000001) connected
2024/01/15, 12:37:00 "PlayerOne" (id=76561198000000000) died
`
	os.WriteFile(filepath.Join(profilesDir, "DayZServer.ADM"), []byte(adminLog), 0644)

	logService := services.NewLogService()

	// Test GetLogPath
	t.Run("GetLogPath", func(t *testing.T) {
		consolePath := logService.GetLogPath(tempDir, "console")
		expected := filepath.Join(profilesDir, "DayZServer.RPT")
		if consolePath != expected {
			t.Errorf("GetLogPath(console) = %s, expected %s", consolePath, expected)
		}

		adminPath := logService.GetLogPath(tempDir, "admin")
		expected = filepath.Join(profilesDir, "DayZServer.ADM")
		if adminPath != expected {
			t.Errorf("GetLogPath(admin) = %s, expected %s", adminPath, expected)
		}
	})

	// Test ReadLogs
	t.Run("ReadLogs", func(t *testing.T) {
		logs, total, err := logService.ReadLogs(tempDir, "console", 10, 0)
		if err != nil {
			t.Fatalf("ReadLogs failed: %v", err)
		}

		if total != 4 {
			t.Errorf("ReadLogs total = %d, expected 4", total)
		}

		if len(logs) > 4 {
			t.Errorf("ReadLogs returned %d logs, expected <= 4", len(logs))
		}

		if len(logs) > 0 && logs[0].Timestamp != "12:34:56" {
			t.Errorf("First log timestamp = %s, expected 12:34:56", logs[0].Timestamp)
		}
	})

	// Test TailLogs
	t.Run("TailLogs", func(t *testing.T) {
		logs, err := logService.TailLogs(tempDir, "console", 2)
		if err != nil {
			t.Fatalf("TailLogs failed: %v", err)
		}

		if len(logs) != 2 {
			t.Errorf("TailLogs returned %d logs, expected 2", len(logs))
		}
	})

	// Test ParseAdminLog
	t.Run("ParseAdminLog", func(t *testing.T) {
		events, err := logService.ParseAdminLog(tempDir, 10)
		if err != nil {
			t.Fatalf("ParseAdminLog failed: %v", err)
		}

		if len(events) != 4 {
			t.Errorf("ParseAdminLog returned %d events, expected 4", len(events))
		}

		// Check connection event
		if len(events) > 0 && events[0].Type != "connection" {
			t.Errorf("First event type = %s, expected connection", events[0].Type)
		}
		if len(events) > 0 && events[0].Player != "PlayerOne" {
			t.Errorf("First event player = %s, expected PlayerOne", events[0].Player)
		}

		// Check chat event
		if len(events) > 1 && events[1].Type != "chat" {
			t.Errorf("Second event type = %s, expected chat", events[1].Type)
		}
		if len(events) > 1 && events[1].Message != "Hello world" {
			t.Errorf("Second event message = %s, expected 'Hello world'", events[1].Message)
		}

		// Check death event
		if len(events) > 3 && events[3].Type != "death" {
			t.Errorf("Fourth event type = %s, expected death", events[3].Type)
		}
	})
}

func TestLogService_Errors(t *testing.T) {
	logService := services.NewLogService()

	t.Run("ReadLogs_NonExistent", func(t *testing.T) {
		_, _, err := logService.ReadLogs("/non/existent/path", "console", 10, 0)
		if err == nil {
			t.Error("Expected error for non-existent path, got nil")
		}
	})

	t.Run("TailLogs_NonExistent", func(t *testing.T) {
		_, err := logService.TailLogs("/non/existent/path", "console", 10)
		if err == nil {
			t.Error("Expected error for non-existent path, got nil")
		}
	})
}

func BenchmarkLogParsing(b *testing.B) {
	// Create temp directory with test data
	tempDir := fmt.Sprintf("/tmp/dayz-bench-%d", time.Now().Unix())
	profilesDir := filepath.Join(tempDir, "profiles")
	os.MkdirAll(profilesDir, 0755)
	defer os.RemoveAll(tempDir)

	// Generate large console log
	var consoleLog string
	for i := 0; i < 1000; i++ {
		timestamp := fmt.Sprintf("%02d:%02d:%02d", i/3600, (i/60)%60, i%60)
		level := []string{"INFO", "WARN", "ERROR"}[i%3]
		consoleLog += fmt.Sprintf("%s %s: Log message %d\n", timestamp, level, i)
	}
	os.WriteFile(filepath.Join(profilesDir, "DayZServer.RPT"), []byte(consoleLog), 0644)

	// Generate large admin log
	var adminLog string
	for i := 0; i < 500; i++ {
		timestamp := fmt.Sprintf("2024/01/15, %02d:%02d:%02d", i/3600, (i/60)%60, i%60)
		adminLog += fmt.Sprintf("%s \"Player%d\" (id=%d) connected\n", timestamp, i%10, 76561198000000000+i)
	}
	os.WriteFile(filepath.Join(profilesDir, "DayZServer.ADM"), []byte(adminLog), 0644)

	logService := services.NewLogService()

	b.Run("ReadLogs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := logService.ReadLogs(tempDir, "console", 100, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ParseAdminLog", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := logService.ParseAdminLog(tempDir, 100)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}