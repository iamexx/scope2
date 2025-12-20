package services

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type ConfigValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type ServerConfig struct {
    Hostname              string `json:"hostname"`
    Password              string `json:"password"`
    MaxPlayers            int    `json:"maxPlayers"`
    Port                  int    `json:"port"`
    AdminPassword         string `json:"adminPassword"`
    ServerTime            string `json:"serverTime"`
    DisableVoN            int    `json:"disableVoN"`
    Disable3rdPerson      int    `json:"disable3rdPerson"`
    DisableCrosshair      int    `json:"disableCrosshair"`
    Persistent            int    `json:"persistent,omitempty"`
    TimeAcceleration      int    `json:"timeAcceleration,omitempty"`
    NightTimeAcceleration int    `json:"nightTimeAcceleration,omitempty"`
}

type ConfigChangeLog struct {
    Timestamp time.Time              `json:"timestamp"`
    User      string                 `json:"user"`
    Changes   map[string]interface{} `json:"changes"`
}

type ConfigEditor struct {
    configChangeLogger *os.File
}

func NewConfigEditor() *ConfigEditor {
    return &ConfigEditor{}
}

func (e *ConfigEditor) ReadConfig(configPath string) (*ServerConfig, error) {
    if !strings.HasSuffix(configPath, "serverDZ.cfg") {
        configPath = filepath.Join(configPath, "serverDZ.cfg")
    }

    file, err := os.Open(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("config file not found: %s", configPath)
        }
        return nil, fmt.Errorf("failed to open config file: %w", err)
    }
    defer file.Close()

    config := &ServerConfig{
        Hostname:           "",
        Password:           "",
        MaxPlayers:         0,
        Port:               0,
        AdminPassword:      "",
        ServerTime:         "SystemTime",
        DisableVoN:         0,
        Disable3rdPerson:   0,
        DisableCrosshair:   0,
        Persistent:         0,
        TimeAcceleration:   1,
        NightTimeAcceleration: 1,
    }

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
            continue
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        value = strings.Trim(value, "\";")

        if err := e.setConfigValue(config, key, value); err != nil {
            fmt.Printf("Warning: failed to parse config value %s=%s: %v\n", key, value, err)
        }
    }

    return config, scanner.Err()
}

func (e *ConfigEditor) setConfigValue(config *ServerConfig, key string, value string) error {
    switch key {
    case "hostname":
        config.Hostname = value
    case "password":
        config.Password = value
    case "maxPlayers":
        maxPlayers, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("invalid maxPlayers value: %s", value)
        }
        config.MaxPlayers = maxPlayers
    case "port":
        port, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("invalid port value: %s", value)
        }
        config.Port = port
    case "adminPassword":
        config.AdminPassword = value
    case "serverTime":
        config.ServerTime = value
    case "disableVoN":
        disableVoN, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("invalid disableVoN value: %s", value)
        }
        config.DisableVoN = disableVoN
    case "disable3rdPerson":
        disable3rdPerson, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("invalid disable3rdPerson value: %s", value)
        }
        config.Disable3rdPerson = disable3rdPerson
    case "disableCrosshair":
        disableCrosshair, err := strconv.Atoi(value)
        if err != nil {
            return fmt.Errorf("invalid disableCrosshair value: %s", value)
        }
        config.DisableCrosshair = disableCrosshair
    case "persistent":
        if value != "" {
            persistent, err := strconv.Atoi(value)
            if err != nil {
                return fmt.Errorf("invalid persistent value: %s", value)
            }
            config.Persistent = persistent
        }
    case "timeAcceleration":
        if value != "" {
            timeAccel, err := strconv.Atoi(value)
            if err != nil {
                return fmt.Errorf("invalid timeAcceleration value: %s", value)
            }
            config.TimeAcceleration = timeAccel
        }
    case "nightTimeAcceleration":
        if value != "" {
            nightTimeAccel, err := strconv.Atoi(value)
            if err != nil {
                return fmt.Errorf("invalid nightTimeAcceleration value: %s", value)
            }
            config.NightTimeAcceleration = nightTimeAccel
        }
    }

    return nil
}

func (e *ConfigEditor) ValidateConfig(config *ServerConfig) []ConfigValidationError {
    var errors []ConfigValidationError

    if config.Hostname == "" {
        errors = append(errors, ConfigValidationError{
            Field:   "hostname",
            Message: "hostname cannot be empty",
        })
    } else if len(config.Hostname) > 64 {
        errors = append(errors, ConfigValidationError{
            Field:   "hostname",
            Message: "hostname cannot exceed 64 characters",
        })
    }

    if config.MaxPlayers < 1 || config.MaxPlayers > 300 {
        errors = append(errors, ConfigValidationError{
            Field:   "maxPlayers",
            Message: "maxPlayers must be between 1 and 300",
        })
    }

    if config.Port < 1 || config.Port > 65535 {
        errors = append(errors, ConfigValidationError{
            Field:   "port",
            Message: "port must be between 1 and 65535",
        })
    }

    if config.DisableVoN != 0 && config.DisableVoN != 1 {
        errors = append(errors, ConfigValidationError{
            Field:   "disableVoN",
            Message: "disableVoN must be 0 or 1",
        })
    }

    if config.Disable3rdPerson != 0 && config.Disable3rdPerson != 1 {
        errors = append(errors, ConfigValidationError{
            Field:   "disable3rdPerson",
            Message: "disable3rdPerson must be 0 or 1",
        })
    }

    if config.DisableCrosshair != 0 && config.DisableCrosshair != 1 {
        errors = append(errors, ConfigValidationError{
            Field:   "disableCrosshair",
            Message: "disableCrosshair must be 0 or 1",
        })
    }

    if !isValidServerTime(config.ServerTime) {
        errors = append(errors, ConfigValidationError{
            Field:   "serverTime",
            Message: "serverTime must be 'SystemTime' or in format 'YYYY/MM/DD/HH/MM'",
        })
    }

    if config.TimeAcceleration < 0 || config.TimeAcceleration > 100 {
        errors = append(errors, ConfigValidationError{
            Field:   "timeAcceleration",
            Message: "timeAcceleration must be between 0 and 100",
        })
    } else if config.TimeAcceleration == 0 && config.TimeAcceleration != 0 {
        config.TimeAcceleration = 1
    }

    if config.NightTimeAcceleration < 0 || config.NightTimeAcceleration > 100 {
        errors = append(errors, ConfigValidationError{
            Field:   "nightTimeAcceleration",
            Message: "nightTimeAcceleration must be between 0 and 100",
        })
    }

    if config.Persistent != 0 && config.Persistent != 1 {
        errors = append(errors, ConfigValidationError{
            Field:   "persistent",
            Message: "persistent must be 0 or 1",
        })
    }

    return errors
}

func isValidServerTime(serverTime string) bool {
    if serverTime == "SystemTime" {
        return true
    }

    timeRegex := regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})/(\d{2})/(\d{2})$`)
    matches := timeRegex.FindStringSubmatch(serverTime)
    if len(matches) != 6 {
        return false
    }

    year, _ := strconv.Atoi(matches[1])
    month, _ := strconv.Atoi(matches[2])
    day, _ := strconv.Atoi(matches[3])
    hour, _ := strconv.Atoi(matches[4])
    minute, _ := strconv.Atoi(matches[5])

    if year < 2000 || year > 2100 {
        return false
    }
    if month < 1 || month > 12 {
        return false
    }
    if day < 1 || day > 31 {
        return false
    }
    if hour < 0 || hour > 23 {
        return false
    }
    if minute < 0 || minute > 59 {
        return false
    }

    return true
}

func (e *ConfigEditor) BackupConfig(configPath string) error {
    if !strings.HasSuffix(configPath, "serverDZ.cfg") {
        configPath = filepath.Join(configPath, "serverDZ.cfg")
    }

    backupPath := configPath + fmt.Sprintf(".bak.%d", time.Now().Unix())

    source, err := os.Open(configPath)
    if err != nil {
        return fmt.Errorf("failed to open config file: %w", err)
    }
    defer source.Close()

    dest, err := os.Create(backupPath)
    if err != nil {
        return fmt.Errorf("failed to create backup file: %w", err)
    }
    defer dest.Close()

    _, err = io.Copy(dest, source)
    if err != nil {
        return fmt.Errorf("failed to copy config to backup: %w", err)
    }

    return nil
}

func (e *ConfigEditor) UpdateConfig(configPath string, updates map[string]interface{}) (*ServerConfig, []ConfigValidationError, error) {
    config, err := e.ReadConfig(configPath)
    if err != nil {
        return nil, nil, err
    }

    for key, value := range updates {
        if err := e.updateConfigField(config, key, value); err != nil {
            return nil, []ConfigValidationError{{
                Field:   key,
                Message: err.Error(),
            }}, nil
        }
    }

    validationErrors := e.ValidateConfig(config)
    if len(validationErrors) > 0 {
        return nil, validationErrors, nil
    }

    return config, nil, nil
}

func (e *ConfigEditor) updateConfigField(config *ServerConfig, key string, value interface{}) error {
    switch key {
    case "hostname":
        if strValue, ok := value.(string); ok {
            config.Hostname = strValue
        } else {
            return fmt.Errorf("hostname must be a string")
        }
    case "password":
        if strValue, ok := value.(string); ok {
            config.Password = strValue
        } else {
            return fmt.Errorf("password must be a string")
        }
    case "maxPlayers":
        if floatValue, ok := value.(float64); ok {
            config.MaxPlayers = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.MaxPlayers = intValue
        } else {
            return fmt.Errorf("maxPlayers must be an integer")
        }
    case "port":
        if floatValue, ok := value.(float64); ok {
            config.Port = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.Port = intValue
        } else {
            return fmt.Errorf("port must be an integer")
        }
    case "adminPassword":
        if strValue, ok := value.(string); ok {
            config.AdminPassword = strValue
        } else {
            return fmt.Errorf("adminPassword must be a string")
        }
    case "serverTime":
        if strValue, ok := value.(string); ok {
            config.ServerTime = strValue
        } else {
            return fmt.Errorf("serverTime must be a string")
        }
    case "disableVoN":
        if floatValue, ok := value.(float64); ok {
            config.DisableVoN = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.DisableVoN = intValue
        } else {
            return fmt.Errorf("disableVoN must be 0 or 1")
        }
    case "disable3rdPerson":
        if floatValue, ok := value.(float64); ok {
            config.Disable3rdPerson = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.Disable3rdPerson = intValue
        } else {
            return fmt.Errorf("disable3rdPerson must be 0 or 1")
        }
    case "disableCrosshair":
        if floatValue, ok := value.(float64); ok {
            config.DisableCrosshair = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.DisableCrosshair = intValue
        } else {
            return fmt.Errorf("disableCrosshair must be 0 or 1")
        }
    case "persistent":
        if value == nil {
            config.Persistent = 0
        } else if floatValue, ok := value.(float64); ok {
            config.Persistent = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.Persistent = intValue
        } else if boolValue, ok := value.(bool); ok {
            if boolValue {
                config.Persistent = 1
            } else {
                config.Persistent = 0
            }
        } else {
            return fmt.Errorf("persistent must be 0 or 1")
        }
    case "timeAcceleration":
        if value == nil {
            config.TimeAcceleration = 1
        } else if floatValue, ok := value.(float64); ok {
            config.TimeAcceleration = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.TimeAcceleration = intValue
        } else {
            return fmt.Errorf("timeAcceleration must be an integer")
        }
    case "nightTimeAcceleration":
        if value == nil {
            config.NightTimeAcceleration = 1
        } else if floatValue, ok := value.(float64); ok {
            config.NightTimeAcceleration = int(floatValue)
        } else if intValue, ok := value.(int); ok {
            config.NightTimeAcceleration = intValue
        } else {
            return fmt.Errorf("nightTimeAcceleration must be an integer")
        }
    default:
        return fmt.Errorf("unknown configuration field: %s", key)
    }

    return nil
}

func (e *ConfigEditor) WriteConfig(configPath string, config *ServerConfig) error {
    if !strings.HasSuffix(configPath, "serverDZ.cfg") {
        configPath = filepath.Join(configPath, "serverDZ.cfg")
    }

    var content strings.Builder

    content.WriteString("// DayZ Server Configuration File\n")
    content.WriteString("// Generated by Scope2 DayZ API\n")
    content.WriteString(fmt.Sprintf("// Timestamp: %s\n\n", time.Now().Format(time.RFC3339)))

    if config.Hostname != "" {
        content.WriteString(fmt.Sprintf("hostname = \"%s\";\n\n", config.Hostname))
    }

    if config.Password != "" {
        content.WriteString(fmt.Sprintf("password = \"%s\";\n\n", config.Password))
    } else {
        content.WriteString("password = \"\";\n\n")
    }

    content.WriteString(fmt.Sprintf("maxPlayers = %d;\n\n", config.MaxPlayers))
    
    if config.Port > 0 {
        content.WriteString(fmt.Sprintf("port = %d;\n\n", config.Port))
    }

    if config.AdminPassword != "" {
        content.WriteString(fmt.Sprintf("adminPassword = \"%s\";\n\n", config.AdminPassword))
    }

    content.WriteString(fmt.Sprintf("serverTime = \"%s\";\n\n", config.ServerTime))

    if config.Persistent != 0 {
        content.WriteString(fmt.Sprintf("persistent = %d;\n\n", config.Persistent))
    }

    if config.TimeAcceleration > 0 && config.TimeAcceleration != 1 {
        content.WriteString(fmt.Sprintf("timeAcceleration = %d;\n\n", config.TimeAcceleration))
    }

    if config.NightTimeAcceleration > 0 && config.NightTimeAcceleration != 1 {
        content.WriteString(fmt.Sprintf("nightTimeAcceleration = %d;\n\n", config.NightTimeAcceleration))
    }

    content.WriteString(fmt.Sprintf("disableVoN = %d;\n\n", config.DisableVoN))
    content.WriteString(fmt.Sprintf("disable3rdPerson = %d;\n\n", config.Disable3rdPerson))
    content.WriteString(fmt.Sprintf("disableCrosshair = %d;\n", config.DisableCrosshair))

    err := os.WriteFile(configPath, []byte(content.String()), 0644)
    if err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }

    return nil
}

func (e *ConfigEditor) LogConfigChange(user string, changes map[string]interface{}) error {
    changeLog := ConfigChangeLog{
        Timestamp: time.Now(),
        User:      user,
        Changes:   changes,
    }

    if e.configChangeLogger == nil {
        logPath := "/var/log/dayz-server/config-changes.log"
        var err error
        e.configChangeLogger, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            return fmt.Errorf("failed to open config change log: %w", err)
        }
    }

    logEntry, err := json.Marshal(changeLog)
    if err != nil {
        return fmt.Errorf("failed to marshal change log: %w", err)
    }

    _, err = e.configChangeLogger.WriteString(string(logEntry) + "\n")
    return err
}

func (e *ConfigEditor) GetBackupPath(configPath string) string {
    if !strings.HasSuffix(configPath, "serverDZ.cfg") {
        configPath = filepath.Join(configPath, "serverDZ.cfg")
    }

    dir := filepath.Dir(configPath)
    backupFiles, err := filepath.Glob(filepath.Join(dir, "serverDZ.cfg.bak.*"))
    if err != nil || len(backupFiles) == 0 {
        return ""
    }

    var latestBackup string
    var latestTime int64

    for _, backup := range backupFiles {
        parts := strings.Split(backup, ".")
        if len(parts) >= 3 {
            timestamp := parts[len(parts)-1]
            if t, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
                if t > latestTime {
                    latestTime = t
                    latestBackup = backup
                }
            }
        }
    }

    return latestBackup
}

func (e *ConfigEditor) RestoreFromBackup(configPath string) error {
    backupPath := e.GetBackupPath(configPath)
    if backupPath == "" {
        return fmt.Errorf("no backup found")
    }

    source, err := os.Open(backupPath)
    if err != nil {
        return fmt.Errorf("failed to open backup file: %w", err)
    }
    defer source.Close()

    if !strings.HasSuffix(configPath, "serverDZ.cfg") {
        configPath = filepath.Join(configPath, "serverDZ.cfg")
    }

    dest, err := os.Create(configPath)
    if err != nil {
        return fmt.Errorf("failed to create config file: %w", err)
    }
    defer dest.Close()

    _, err = io.Copy(dest, source)
    if err != nil {
        return fmt.Errorf("failed to restore from backup: %w", err)
    }

    return nil
}
