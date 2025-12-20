package services

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
)

type LogEntry struct {
    Timestamp string `json:"timestamp"`
    Level     string `json:"level"`
    Message   string `json:"message"`
}

type AdminLogEvent struct {
    Timestamp string `json:"timestamp"`
    Type      string `json:"type"`
    Player    string `json:"player,omitempty"`
    Message   string `json:"message"`
    Details   string `json:"details,omitempty"`
}

type LogService struct{}

func NewLogService() *LogService {
    return &LogService{}
}

func (s *LogService) GetLogPath(serverFolder string, logType string) string {
    basePath := filepath.Join(serverFolder, "profiles")
    switch logType {
    case "console":
        return filepath.Join(basePath, "DayZServer.RPT")
    case "admin":
        return filepath.Join(basePath, "DayZServer.ADM")
    default:
        return filepath.Join(basePath, "DayZServer.RPT")
    }
}

func (s *LogService) ReadLogs(serverFolder string, logType string, lines int, offset int) ([]LogEntry, int, error) {
    logPath := s.GetLogPath(serverFolder, logType)
    
    file, err := os.Open(logPath)
    if err != nil {
        if os.IsNotExist(err) {
            return []LogEntry{}, 0, fmt.Errorf("log file not found: %s", logPath)
        }
        return []LogEntry{}, 0, fmt.Errorf("failed to open log file: %w", err)
    }
    defer file.Close()

    allLines, err := readAllLines(file)
    if err != nil {
        return []LogEntry{}, 0, fmt.Errorf("failed to read log file: %w", err)
    }

    totalLines := len(allLines)
    entries := parseLogEntries(allLines)

    start := offset
    if start > len(entries) {
        start = len(entries)
    }
    
    end := start + lines
    if end > len(entries) {
        end = len(entries)
    }

    return entries[start:end], totalLines, nil
}

func (s *LogService) TailLogs(serverFolder string, logType string, lastN int) ([]LogEntry, error) {
    logPath := s.GetLogPath(serverFolder, logType)
    
    file, err := os.Open(logPath)
    if err != nil {
        if os.IsNotExist(err) {
            return []LogEntry{}, fmt.Errorf("log file not found: %s", logPath)
        }
        return []LogEntry{}, fmt.Errorf("failed to open log file: %w", err)
    }
    defer file.Close()

    allLines, err := readAllLines(file)
    if err != nil {
        return []LogEntry{}, fmt.Errorf("failed to read log file: %w", err)
    }

    start := len(allLines) - lastN
    if start < 0 {
        start = 0
    }

    entries := parseLogEntries(allLines[start:])
    return entries, nil
}

func (s *LogService) ParseAdminLog(serverFolder string, lines int) ([]AdminLogEvent, error) {
    logPath := s.GetLogPath(serverFolder, "admin")
    
    file, err := os.Open(logPath)
    if err != nil {
        if os.IsNotExist(err) {
            return []AdminLogEvent{}, nil
        }
        return []AdminLogEvent{}, fmt.Errorf("failed to open admin log: %w", err)
    }
    defer file.Close()

    allLines, err := readAllLines(file)
    if err != nil {
        return []AdminLogEvent{}, fmt.Errorf("failed to read admin log: %w", err)
    }

    start := len(allLines) - lines
    if start < 0 {
        start = 0
    }

    events := parseAdminLogEvents(allLines[start:])
    return events, nil
}

func readAllLines(file *os.File) ([]string, error) {
    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func parseLogEntries(lines []string) []LogEntry {
    var entries []LogEntry
    dayzLogRegex := regexp.MustCompile(`^(\d{2}:\d{2}:\d{2})\s+(\w+)\s*:\s*(.+)$`)

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        matches := dayzLogRegex.FindStringSubmatch(line)
        if len(matches) == 4 {
            entries = append(entries, LogEntry{
                Timestamp: matches[1],
                Level:     matches[2],
                Message:   matches[3],
            })
        } else {
            entries = append(entries, LogEntry{
                Timestamp: "",
                Level:     "INFO",
                Message:   line,
            })
        }
    }

    return entries
}

func parseAdminLogEvents(lines []string) []AdminLogEvent {
    var events []AdminLogEvent
    
    connectionRegex := regexp.MustCompile(`(\d{4}/\d{2}/\d{2},\s*\d{2}:\d{2}:\d{2})\s+"([^"]+)"\s+\(id=(\S+)\)\s+connected`)
    disconnectionRegex := regexp.MustCompile(`(\d{4}/\d{2}/\d{2},\s*\d{2}:\d{2}:\d{2})\s+"([^"]+)"\s+\(id=(\S+)\)\s+disconnected`)
    chatRegex := regexp.MustCompile(`(\d{4}/\d{2}/\d{2},\s*\d{2}:\d{2}:\d{2})\s+"([^"]+)"\s+\(id=(\S+)\)\s+pos=<([^>]+)>\s+\(chat\)\s+(.+)`)
    deathRegex := regexp.MustCompile(`(\d{4}/\d{2}/\d{2},\s*\d{2}:\d{2}:\d{2})\s+"([^"]+)"\s+\(id=(\S+)\)\s+died`)
    damageRegex := regexp.MustCompile(`(\d{4}/\d{2}/\d{2},\s*\d{2}:\d{2}:\d{2})\s+"([^"]+)"\s+\(id=(\S+)\)\s+pos=<([^>]+)>\s+took\s+(\d+)\s+damage\s+from\s+\"([^"]+)\"`)

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        var event AdminLogEvent
        
        if matches := connectionRegex.FindStringSubmatch(line); matches != nil {
            event = AdminLogEvent{
                Timestamp: matches[1],
                Type:      "connection",
                Player:    matches[2],
                Message:   fmt.Sprintf("Connected (id=%s)", matches[3]),
            }
        } else if matches := disconnectionRegex.FindStringSubmatch(line); matches != nil {
            event = AdminLogEvent{
                Timestamp: matches[1],
                Type:      "disconnection",
                Player:    matches[2],
                Message:   fmt.Sprintf("Disconnected (id=%s)", matches[3]),
            }
        } else if matches := chatRegex.FindStringSubmatch(line); matches != nil {
            event = AdminLogEvent{
                Timestamp: matches[1],
                Type:      "chat",
                Player:    matches[2],
                Message:   matches[5],
                Details:   fmt.Sprintf("Position: %s", matches[4]),
            }
        } else if matches := deathRegex.FindStringSubmatch(line); matches != nil {
            event = AdminLogEvent{
                Timestamp: matches[1],
                Type:      "death",
                Player:    matches[2],
                Message:   fmt.Sprintf("Died (id=%s)", matches[3]),
            }
        } else if matches := damageRegex.FindStringSubmatch(line); matches != nil {
            damage, _ := strconv.Atoi(matches[5])
            event = AdminLogEvent{
                Timestamp: matches[1],
                Type:      "damage",
                Player:    matches[2],
                Message:   fmt.Sprintf("Took %d damage from %s", damage, matches[6]),
                Details:   fmt.Sprintf("Position: %s", matches[4]),
            }
        } else {
            parts := strings.SplitN(line, " ", 2)
            if len(parts) >= 2 {
                event = AdminLogEvent{
                    Timestamp: parts[0],
                    Type:      "other",
                    Message:   parts[1],
                }
            }
        }

        if event.Type != "" {
            events = append(events, event)
        }
    }

    return events
}