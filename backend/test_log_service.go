package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iamexx/scope2-dayz-api/internal/services"
)

func main() {
	fmt.Println("=== Testing Log Service ===")
	
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

	fmt.Println("\n--- Testing ReadLogs ---")
	logs, total, err := logService.ReadLogs(tempDir, "console", 10, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total lines: %d\n", total)
	fmt.Printf("Returned entries: %d\n", len(logs))
	if len(logs) > 0 {
		fmt.Printf("First entry: %+v\n", logs[0])
	}

	fmt.Println("\n--- Testing TailLogs ---")
	tailLogs, err := logService.TailLogs(tempDir, "console", 2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Tailed entries: %d\n", len(tailLogs))
	for i, entry := range tailLogs {
		fmt.Printf("  Entry %d: %s - %s\n", i, entry.Timestamp, entry.Message)
	}

	fmt.Println("\n--- Testing ParseAdminLog ---")
	events, err := logService.ParseAdminLog(tempDir, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Admin events: %d\n", len(events))
	for i, event := range events {
		fmt.Printf("  Event %d: [%s] %s - %s (%s)\n", 
			i, event.Type, event.Timestamp, event.Player, event.Message)
		if event.Details != "" {
			fmt.Printf("    Details: %s\n", event.Details)
		}
	}

	fmt.Println("\n=== Testing Config Editor ===")
	
	// Test config file
	configContent := `// DayZ Server Configuration
hostname = "My Server";
maxPlayers = 50;
port = 2302;
password = "serverpass";
serverTime = "SystemTime";
disableVoN = 0;
disable3rdPerson = 1;
disableCrosshair = 0;
`
	configPath := filepath.Join(tempDir, "serverDZ.cfg")
	os.WriteFile(configPath, []byte(configContent), 0644)

	configEditor := services.NewConfigEditor()

	fmt.Println("\n--- Testing ReadConfig ---")
	config, err := configEditor.ReadConfig(tempDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Config loaded: %+v\n", config)

	fmt.Println("\n--- Testing ValidateConfig ---")
	validationErrors := configEditor.ValidateConfig(config)
	if len(validationErrors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range validationErrors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	} else {
		fmt.Println("Config validation passed!")
	}

	fmt.Println("\n--- Testing BackupConfig ---")
	err = configEditor.BackupConfig(tempDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Backup created successfully")

	// List backup files
	files, _ := os.ReadDir(profilesDir)
	fmt.Println("Backup files:")
	for _, file := range files {
		if strings.Contains(file.Name(), ".bak.") {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	fmt.Println("\n--- Testing UpdateConfig ---")
	updates := map[string]interface{}{
		"maxPlayers": 75,
		"hostname":   "Updated Server Name",
	}
	updatedConfig, validationErrors, err := configEditor.UpdateConfig(tempDir, updates)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(validationErrors) > 0 {
		fmt.Println("Validation errors:")
		for _, err := range validationErrors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	} else {
		fmt.Printf("Config updated successfully: %+v\n", updatedConfig)
	}

	fmt.Println("\n--- Testing WriteConfig ---")
	testConfig := &services.ServerConfig{
		Hostname:           "Test Server",
		Password:           "testpass",
		MaxPlayers:         100,
		Port:               2304,
		AdminPassword:      "adminpass",
		ServerTime:         "SystemTime",
		DisableVoN:         0,
		Disable3rdPerson:   1,
		DisableCrosshair:   0,
	}
	
	writePath := filepath.Join(profilesDir, "test_config.cfg")
	err = configEditor.WriteConfig(writePath, testConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Config written successfully")

	// Read back the written config
	readConfig, err := configEditor.ReadConfig(writePath)
	if err != nil {
		fmt.Printf("Error reading back config: %v\n", err)
		return
	}
	fmt.Printf("Read back config: %+v\n", readConfig)

	fmt.Println("\n=== All tests completed successfully! ===")
}