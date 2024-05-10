package main

import (
	"fmt"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DatePattern      = "[0-9]{2}\\.[0-9]{2}\\.[12]{1}[0-9]{3}"
	DateSearchFormat = "02.01.2006"
)

var storeService *StoreService

type StoreService struct {
	store *Store
}

func (ss *StoreService) Init() error {
	if ss.store == nil {
		db, err := DbConnection()
		if err != nil {
			return err
		}
		ss.store = &Store{db: db}
	}
	return nil
}

func GetStoreService() *StoreService {
	if storeService == nil {
		storeService = &StoreService{}
		err := storeService.Init()
		if err != nil {
			fmt.Println("Impossible to create db service instance.")
			return nil
		}
	}
	return storeService
}

func (ss *StoreService) AddTask(nextDate string, title string, comment string, repeat string) (int64, error) {
	return ss.store.AddTask(nextDate, title, comment, repeat)
}

func (ss *StoreService) GetTaskById(id string) (*Task, error) {
	return ss.store.GetTaskById(id)
}

func (ss *StoreService) UpdateTask(task *Task) error {
	return ss.store.UpdateTask(task)
}

func (ss *StoreService) GetTasks(search string, limit int) (*[]Task, error) {
	if search == "" {
		return ss.GetAllTasks(limit)
	}

	dateMatched, err := regexp.MatchString(DatePattern, search)
	if err != nil {
		return nil, err
	}

	if dateMatched {
		dateForSearch, err := time.Parse(DateSearchFormat, search)
		if err != nil {
			return nil, err
		}
		return ss.SearchByDate(dateForSearch, limit)
	}

	return ss.SearchByString(search, limit)
}

func (ss *StoreService) RescheduleTask(id string) error {
	task, err := ss.store.GetById(id)
	if err != nil {
		return err
	}
	if task.Repeat == "" {
		return ss.DeleteTask(id)
	}
	nextDate, err := NextDate(TodayDate(time.Now()), task.Date, task.Repeat)
	if err != nil {
		return err
	}
	return ss.store.UpdateTaskDate(nextDate, id)
}

func (ss *StoreService) DeleteTask(id string) error {
	return ss.store.DeleteTask(id)
}

func (ss *StoreService) GetAllTasks(limit int) (*[]Task, error) {
	return ss.store.GetAllTasks(limit)
}

func (ss *StoreService) SearchByDate(date time.Time, limit int) (*[]Task, error) {
	dateForSearch := date.Format(DateScheduleFormat)
	return ss.store.SearchByDate(dateForSearch, limit)
}

func (ss *StoreService) SearchByString(search string, limit int) (*[]Task, error) {
	return ss.store.SearchByString(search, limit)
}
