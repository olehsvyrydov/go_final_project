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

	if date.Before(today) {
		date = today
	}

	if len(days) == 0 {
		return "", fmt.Errorf("days should be specified")
	}

	if len(rules) == 2 {
		months, err := stringToIntArray(rules[1])
		if err != nil {
			return "", err
		}

		sort.Ints(months)
		if months[0] < 0 {
			return "", fmt.Errorf("month cannot have negative value")
		}
		num := find(int(date.Month()), months)

		nextMonthDate := getFirstDayOfYear(date).AddDate(0, num-1, 0)
		d, err := findNextDay(nextMonthDate, days)
		if err != nil {
			return "", err
		}
		nd := d.Format(DATE_SCHEDULE_FORMAT)
		return nd, nil
	}

	d, err := findNextDay(date, days)
	if err != nil {
		return "", err
	}
	nd := d.Format(DATE_SCHEDULE_FORMAT)
	return nd, nil
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

func findNextDay(date time.Time, days []int) (time.Time, error) {

	newDays, err := transformForDate(date, &days)
	if err != nil {
		return date, err
	}

	nextDay := adjustingFind(date.Day(), newDays)

	if nextDay < 0 {
		return getFirstdayOfNextMonth(date).AddDate(0, 0, newDays[0]-1), nil
	}

	for {
		nextDate := getFirstdayOfMonth(date).AddDate(0, 0, nextDay-1)
		if nextDate.Day() == nextDay {
			return nextDate, nil
		}
		date = date.AddDate(0, 1, 0)
	}
}

func TodayDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func transformForDate(date time.Time, days *[]int) ([]int, error) {
	nextDay := date.AddDate(0, 0, 1)
	var newDays []int
	for _, d := range *days {
		if d < -2 || d > 31 {
			return nil, fmt.Errorf("incorrect date format")
		}
		if d < 0 {
			firstDayOfNextMonth := getFirstdayOfNextMonth(nextDay)
			d = firstDayOfNextMonth.AddDate(0, 0, d).Day()
			newDays = append(newDays, d)
		} else {
			newDays = append(newDays, d)
		}
	}
	sort.Ints(newDays)
	return newDays, nil
}

func getFirstdayOfNextMonth(date time.Time) time.Time {
	nextDate := date.AddDate(0, 1, 0)
	return getFirstdayOfMonth(nextDate)
}

func getFirstdayOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func getFirstDayOfYear(date time.Time) time.Time {
	return time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
}

func adjustingFind(num int, nums []int) int {
	for _, v := range nums {
		if num < v {
			return v
		}
	}
	return -1
}

func find(num int, nums []int) int {
	for _, v := range nums {
		if num <= v {
			return v
		}
	}
	return -1
}
