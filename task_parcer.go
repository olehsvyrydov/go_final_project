package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func NextDate(today time.Time, date string, repeat string) (string, error) {
	var startDate time.Time
	var err error

	if date == "" {
		startDate = today
	} else if startDate, err = time.Parse(DATE_SCHEDULE_FORMAT, date); err != nil {
		return "", err
	}

	rule := strings.SplitN(repeat, " ", 2)

	switch rule[0] {
	case "d":
		if len(rule) != 2 {
			return "", fmt.Errorf("rule for day should be defined")
		}
		return ruleForDay(today, startDate, rule[1])
	case "w":
		if len(rule) != 2 {
			return "", fmt.Errorf("rule for week should be defined")
		}
		return ruleForWeek(today, startDate, rule[1])
	case "m":
		if len(rule) != 2 {
			return "", fmt.Errorf("rule for month should be defined")
		}
		return ruleForMonth(today, startDate, rule[1])
	case "y":
		return ruleForYear(today, startDate)
	default:
		return "", fmt.Errorf("usupported format")
	}
}

func ruleForDay(today time.Time, date time.Time, rule string) (string, error) {

	countDays, err := strconv.Atoi(rule)
	if err != nil {
		return "", fmt.Errorf("rule for day cannot be converted to integer")
	}
	if countDays < 1 || countDays > 400 {
		return "", fmt.Errorf("invalid rule. You should use range from 1 till including 400")
	}

	for {
		date = date.AddDate(0, 0, countDays)
		if date.After(today) && date != today {
			break
		}
	}

	fmt.Println("today =", today.Format(DATE_SCHEDULE_FORMAT), "date =", date.Format(DATE_SCHEDULE_FORMAT), "rule =", rule)
	return date.Format(DATE_SCHEDULE_FORMAT), nil
}

func ruleForYear(today, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(today) {
			break
		}
	}
	return date.Format(DATE_SCHEDULE_FORMAT), nil
}

func ruleForWeek(today time.Time, date time.Time, rule string) (string, error) {
	days, err := stringToIntArray(rule)
	if err != nil {
		return "", err
	}
	sort.Ints(days)
	if date.Before(today) {
		date = today
	}
	weekday := int(date.Weekday())
	for _, day := range days {
		if day < 1 || day > 7 {
			return "", fmt.Errorf("wrong format day for week. It shouls be from 1 till 7")
		}

		if day <= weekday {
			continue
		}

		date := date.AddDate(0, 0, day-weekday)
		return date.Format(dateFormat), nil
	}

	date = date.AddDate(0, 0, 7-weekday+days[0])
	return date.Format(DATE_SCHEDULE_FORMAT), nil
}

func ruleForMonth(today time.Time, date time.Time, rule string) (string, error) {
	rules := strings.Split(rule, " ")

	days, err := stringToIntArray(rules[0])
	if err != nil {
		return "", err
	}

	var startDate time.Time
	if today.After(date) {
		startDate = today
	} else {
		startDate = date
	}

	if len(rules) == 2 {
		months, err := stringToIntArray(rules[1])
		if err != nil {
			return "", err
		}
		for contains(months, int(startDate.Month())) {
			startDate = startDate.AddDate(0, 1, 0)
		}
	}

	nextDate, err := findNextDate(days, startDate)
	if err != nil {
		return "", err
	}

	return nextDate.Format(DATE_SCHEDULE_FORMAT), nil
}

func stringToIntArray(input string) ([]int, error) {
	strArray := strings.Split(input, ",")

	intArray := make([]int, len(strArray))

	for i, str := range strArray {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("error converting string to integer: %s", err)
		}
		intArray[i] = num
	}

	return intArray, nil
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func findNextDate(days []int, date time.Time) (time.Time, error) {
	if len(days) == 0 {
		return date, nil
	}
	monthDays := getLastDayOfMonth(date).Day()
	fmt.Println("last Day of Month is ", monthDays, " days is ", days, " month is ", date.Month().String())
	var d int
	for {
		date = date.AddDate(0, 0, 1)

		for _, v := range days {
			fmt.Println("v = ", v)
			if v > monthDays && v <= 31 {
				dateNext := date
				for {
					dateNext = dateNext.AddDate(0, 1, 0)
					monthDaysNext := getLastDayOfMonth(dateNext).Day()
					if v == monthDaysNext {
						return TodayDate(dateNext), nil
					}
				}
			}
			if v > monthDays || v < -monthDays {
				return date, fmt.Errorf("incorrect date number. v is %d, monthDays is %d", v, monthDays)
			}
			d = v
			if d < 0 {
				d = monthDays + 1 + v
			}
			if d == date.Day() {
				return date, nil
			}
		}
	}
}

func getLastDayOfMonth(date time.Time) time.Time {
	nextMonth := date.AddDate(0, 1, 0)
	firstDayOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDayOfNextMonth.Add(-time.Hour)
	return lastDay
}

func TodayDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}
