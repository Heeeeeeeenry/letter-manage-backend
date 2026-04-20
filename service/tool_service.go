package service

import (
	"errors"
	"fmt"
	"time"
)

// TimeDiff calculates the difference between two times
func TimeDiff(args map[string]interface{}) (map[string]interface{}, error) {
	startStr, ok := args["start"].(string)
	if !ok {
		return nil, errors.New("start required")
	}
	endStr, ok := args["end"].(string)
	if !ok {
		return nil, errors.New("end required")
	}
	layout := "2006-01-02 15:04:05"
	start, err := time.ParseInLocation(layout, startStr, time.Local)
	if err != nil {
		layout = "2006-01-02"
		start, err = time.ParseInLocation(layout, startStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid start time: %s", startStr)
		}
	}
	end, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, time.Local)
	if err != nil {
		end, err = time.ParseInLocation("2006-01-02", endStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid end time: %s", endStr)
		}
	}
	dur := end.Sub(start)
	days := int(dur.Hours() / 24)
	hours := int(dur.Hours()) % 24
	minutes := int(dur.Minutes()) % 60
	return map[string]interface{}{
		"days":         days,
		"hours":        hours,
		"minutes":      minutes,
		"total_hours":  dur.Hours(),
		"total_minutes": dur.Minutes(),
	}, nil
}

// TimeAdd adds duration to a time
func TimeAdd(args map[string]interface{}) (map[string]interface{}, error) {
	startStr, ok := args["start"].(string)
	if !ok {
		return nil, errors.New("start required")
	}
	layout := "2006-01-02 15:04:05"
	start, err := time.ParseInLocation(layout, startStr, time.Local)
	if err != nil {
		start, err = time.ParseInLocation("2006-01-02", startStr, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid start time")
		}
	}
	days := 0
	hours := 0
	if v, ok := args["days"].(float64); ok {
		days = int(v)
	}
	if v, ok := args["hours"].(float64); ok {
		hours = int(v)
	}
	result := start.AddDate(0, 0, days).Add(time.Duration(hours) * time.Hour)
	return map[string]interface{}{
		"result": result.Format("2006-01-02 15:04:05"),
	}, nil
}

// HolidayCheck checks if a date is a holiday/weekend
func HolidayCheck(args map[string]interface{}) (map[string]interface{}, error) {
	dateStr, ok := args["date"].(string)
	if !ok {
		return nil, errors.New("date required")
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid date")
	}
	weekday := t.Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday
	return map[string]interface{}{
		"date":       dateStr,
		"is_holiday": isWeekend,
		"is_weekend": isWeekend,
		"weekday":    weekday.String(),
	}, nil
}

// WorkdaysCalculate calculates number of working days between two dates
func WorkdaysCalculate(args map[string]interface{}) (map[string]interface{}, error) {
	startStr, ok := args["start"].(string)
	if !ok {
		return nil, errors.New("start required")
	}
	endStr, ok := args["end"].(string)
	if !ok {
		return nil, errors.New("end required")
	}
	start, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid start date")
	}
	end, err := time.ParseInLocation("2006-01-02", endStr, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid end date")
	}
	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		w := d.Weekday()
		if w != time.Saturday && w != time.Sunday {
			count++
		}
	}
	return map[string]interface{}{
		"workdays": count,
	}, nil
}

// WorkdaysAdd adds working days to a date
func WorkdaysAdd(args map[string]interface{}) (map[string]interface{}, error) {
	startStr, ok := args["start"].(string)
	if !ok {
		return nil, errors.New("start required")
	}
	daysF, ok := args["days"].(float64)
	if !ok {
		return nil, errors.New("days required")
	}
	days := int(daysF)
	start, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid start date")
	}
	current := start
	added := 0
	for added < days {
		current = current.AddDate(0, 0, 1)
		w := current.Weekday()
		if w != time.Saturday && w != time.Sunday {
			added++
		}
	}
	return map[string]interface{}{
		"result": current.Format("2006-01-02"),
	}, nil
}

// MonthCalendar returns calendar info for a month
func MonthCalendar(args map[string]interface{}) (map[string]interface{}, error) {
	year := time.Now().Year()
	month := int(time.Now().Month())
	if v, ok := args["year"].(float64); ok {
		year = int(v)
	}
	if v, ok := args["month"].(float64); ok {
		month = int(v)
	}
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	var days []map[string]interface{}
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		w := d.Weekday()
		isWeekend := w == time.Saturday || w == time.Sunday
		days = append(days, map[string]interface{}{
			"date":       d.Format("2006-01-02"),
			"day":        d.Day(),
			"weekday":    int(w),
			"is_weekend": isWeekend,
			"is_holiday": isWeekend,
		})
	}
	return map[string]interface{}{
		"year":  year,
		"month": month,
		"days":  days,
	}, nil
}
