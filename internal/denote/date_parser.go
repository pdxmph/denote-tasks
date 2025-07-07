package denote

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseNaturalDate parses natural language dates into YYYY-MM-DD format
func ParseNaturalDate(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return "", nil
	}
	
	now := time.Now()
	var targetDate time.Time
	
	// Try standard YYYY-MM-DD format first
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, input); matched {
		_, err := time.Parse("2006-01-02", input)
		if err != nil {
			return "", fmt.Errorf("invalid date format: %s", input)
		}
		return input, nil
	}
	
	// Relative date patterns: 1d, 5d, 2w, 1m, etc.
	if matched := regexp.MustCompile(`^(\d+)([dwmy])$`).FindStringSubmatch(input); matched != nil {
		num, _ := strconv.Atoi(matched[1])
		unit := matched[2]
		
		switch unit {
		case "d": // days
			targetDate = now.AddDate(0, 0, num)
		case "w": // weeks
			targetDate = now.AddDate(0, 0, num*7)
		case "m": // months
			targetDate = now.AddDate(0, num, 0)
		case "y": // years
			targetDate = now.AddDate(num, 0, 0)
		}
		
		return targetDate.Format("2006-01-02"), nil
	}
	
	// Special keywords
	switch input {
	case "today", "tod":
		return now.Format("2006-01-02"), nil
	case "tomorrow", "tom":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	}
	
	// Weekday names
	weekday, err := parseWeekday(input)
	if err == nil {
		// Find next occurrence of this weekday
		daysUntil := int(weekday - now.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7 // Next week
		}
		targetDate = now.AddDate(0, 0, daysUntil)
		return targetDate.Format("2006-01-02"), nil
	}
	
	// Month day format: "jan 15", "15 jan", "january 15"
	if date, err := parseMonthDay(input, now.Year()); err == nil {
		return date, nil
	}
	
	return "", fmt.Errorf("unrecognized date format: %s", input)
}

// parseWeekday parses weekday names and abbreviations
func parseWeekday(input string) (time.Weekday, error) {
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"sun":       time.Sunday,
		"monday":    time.Monday,
		"mon":       time.Monday,
		"tuesday":   time.Tuesday,
		"tue":       time.Tuesday,
		"tues":      time.Tuesday,
		"wednesday": time.Wednesday,
		"wed":       time.Wednesday,
		"thursday":  time.Thursday,
		"thu":       time.Thursday,
		"thur":      time.Thursday,
		"thurs":     time.Thursday,
		"friday":    time.Friday,
		"fri":       time.Friday,
		"saturday":  time.Saturday,
		"sat":       time.Saturday,
	}
	
	if wd, ok := weekdays[input]; ok {
		return wd, nil
	}
	
	return time.Sunday, fmt.Errorf("not a weekday: %s", input)
}

// parseMonthDay parses formats like "jan 15" or "15 jan"
func parseMonthDay(input string, year int) (string, error) {
	months := map[string]int{
		"january":   1, "jan": 1,
		"february":  2, "feb": 2,
		"march":     3, "mar": 3,
		"april":     4, "apr": 4,
		"may":       5,
		"june":      6, "jun": 6,
		"july":      7, "jul": 7,
		"august":    8, "aug": 8,
		"september": 9, "sep": 9, "sept": 9,
		"october":   10, "oct": 10,
		"november":  11, "nov": 11,
		"december":  12, "dec": 12,
	}
	
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return "", fmt.Errorf("not a month day format")
	}
	
	var month, day int
	var err error
	
	// Try "month day" format
	if m, ok := months[parts[0]]; ok {
		month = m
		day, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}
	} else if m, ok := months[parts[1]]; ok {
		// Try "day month" format
		month = m
		day, err = strconv.Atoi(parts[0])
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("invalid month name")
	}
	
	// Validate day
	if day < 1 || day > 31 {
		return "", fmt.Errorf("invalid day: %d", day)
	}
	
	// Create date and check if it's in the past
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if date.Before(time.Now()) {
		// Use next year
		date = time.Date(year+1, time.Month(month), day, 0, 0, 0, 0, time.Local)
	}
	
	return date.Format("2006-01-02"), nil
}