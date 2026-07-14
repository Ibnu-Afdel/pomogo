package store

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"
)

// ExportSessions returns session data in JSON or CSV format.
func (s *Store) ExportSessions(format string, start, end time.Time) (string, error) {
	sessions, err := s.GetSessions(start, end)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal sessions to JSON: %w", err)
		}
		return string(data), nil

	case "csv":
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)

		// Header
		header := []string{"id", "type", "task", "note", "started_at", "ended_at", "completed", "duration_secs", "project_name"}
		if err := writer.Write(header); err != nil {
			return "", fmt.Errorf("failed to write CSV header: %w", err)
		}

		// Rows
		for _, sess := range sessions {
			completedStr := "false"
			if sess.Completed {
				completedStr = "true"
			}

			row := []string{
				fmt.Sprintf("%d", sess.ID),
				sess.Type,
				sess.Task,
				sess.Note,
				sess.StartedAt.Format(time.RFC3339),
				sess.EndedAt.Format(time.RFC3339),
				completedStr,
				fmt.Sprintf("%d", sess.DurationSecs),
				sess.ProjectName,
			}
			if err := writer.Write(row); err != nil {
				return "", fmt.Errorf("failed to write CSV row: %w", err)
			}
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return "", fmt.Errorf("CSV writer error: %w", err)
		}

		return buf.String(), nil

	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// GenerateMarkdownReport returns a formatted markdown text of session summaries.
func (s *Store) GenerateMarkdownReport(start, end time.Time) (string, error) {
	sessions, err := s.GetSessions(start, end)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	var totalWorkSecs int
	var completedCount int
	var skippedCount int
	projectDurations := make(map[string]int)

	for _, sess := range sessions {
		if sess.Type == "work" {
			if sess.Completed {
				completedCount++
				totalWorkSecs += sess.DurationSecs
				projName := sess.ProjectName
				if projName == "" {
					projName = "(no project)"
				}
				projectDurations[projName] += sess.DurationSecs
			} else {
				skippedCount++
			}
		}
	}

	var buf bytes.Buffer
	buf.WriteString("# PomoGo Focus Report\n\n")
	buf.WriteString(fmt.Sprintf("Report Period: %s to %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02")))

	buf.WriteString("## Summary\n")
	totalHours := float64(totalWorkSecs) / 3600.0
	buf.WriteString(fmt.Sprintf("- **Total Focus Time**: %.2f hours\n", totalHours))
	buf.WriteString(fmt.Sprintf("- **Completed Sessions**: %d\n", completedCount))
	buf.WriteString(fmt.Sprintf("- **Skipped Sessions**: %d\n\n", skippedCount))

	buf.WriteString("## Project Breakdown\n")
	if len(projectDurations) == 0 {
		buf.WriteString("No completed sessions in this period.\n")
	} else {
		buf.WriteString("| Project | Hours Focused |\n")
		buf.WriteString("|---|---|\n")
		for proj, secs := range projectDurations {
			hours := float64(secs) / 3600.0
			buf.WriteString(fmt.Sprintf("| %s | %.2f |\n", proj, hours))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
