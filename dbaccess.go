package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var storeService *StoreService

type StoreService struct {
	store *Store
}

type Store struct {
	db *sqlx.DB
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

func openDB() *sqlx.DB {
	dbfile := DB_FILE
	envFile := os.Getenv("TODO_DBFILE")
	if len(envFile) > 0 {
		dbfile = envFile
		path := filepath.Dir(dbfile)
		pathInfo, err := os.Stat(path)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if !pathInfo.IsDir() {
			err := os.MkdirAll(path, 0644)
			if err != nil {
				fmt.Println(err)
				return nil
			}
		}
	}
	fmt.Println("Database file", dbfile)
	db, err := sqlx.Open("sqlite3", dbfile)
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return nil
	}
	fmt.Println("Database opened")
	return db
}

func DbConnection() (*sqlx.DB, error) {
	db := openDB()
	if db != nil {
		if !isTableExists(db) {
			fmt.Println("Table not exists")
			db.MustExec(SCHEMA)
			fmt.Println("Table created")
		}
		fmt.Println("Database returned")
		return db, nil
	}

	return nil, fmt.Errorf("database error")
}

func isTableExists(db *sqlx.DB) bool {
	var count int
	err := db.Get(&count, `SELECT count(id) FROM scheduler`)
	if err != nil {
		fmt.Println("Table is absent")
		return false
	}
	fmt.Println("Table exists. Count", count)
	return true
}

func (ss *StoreService) AddTask(nextDate string, title string, comment string, repeat string) (int64, error) {
	res, err := ss.store.db.NamedExec("INSERT INTO scheduler (date, title, comment, repeat) values (:date, :title, :comment, :repeat)",
		map[string]interface{}{
			"date":    nextDate,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		})

	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (ss *StoreService) GetTaskById(id string) (*Task, error) {
	task := Task{}
	err := ss.store.db.Get(&task, "SELECT * FROM scheduler WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (ss *StoreService) UpdateTask(task *Task) error {
	res, err := ss.store.db.NamedExec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		map[string]interface{}{
			"id":      task.Id,
			"date":    task.Date,
			"comment": task.Comment,
			"title":   task.Title,
			"repeat":  task.Repeat})
	if err != nil {
		return err
	}
	fmt.Println("Update Task db query result error", err, "for task", task)
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if num == 0 {
		return fmt.Errorf("no rows were affected after update task with id %s and params %s", task.Id, task)
	}
	return err
}

func (ss *StoreService) GetTasks(search string, limit int) (*[]Task, error) {
	if search == "" {
		return ss.GetAllTasks(limit)
	}

	dateMatched, err := regexp.MatchString(DATE_PATTERN, search)
	if err != nil {
		return nil, err
	}

	if dateMatched {
		dateForSearch, err := time.Parse(DATE_SEARCH_FORMAT, search)
		if err != nil {
			return nil, err
		}
		return ss.SearchByDate(dateForSearch, limit)
	}

	return ss.SearchByString(search, limit)
}

func (ss *StoreService) RescheduleTask(id string) error {
	task := Task{}
	err := ss.store.db.Get(&task, "SELECT * FROM scheduler WHERE id = $1", id)
	if err != nil {
		return err
	}
	if task.Repeat == "" {
		ss.DeleteTask(id)
	} else {
		nextDate, err := NextDate(TodayDate(time.Now()), task.Date, task.Repeat)
		if err != nil {
			return err
		}
		res := ss.store.db.MustExec("UPDATE scheduler SET date = $1 WHERE id = $2", nextDate, id)
		num, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if num == 0 {
			return fmt.Errorf("no row was affected for id = %s", id)
		}
		return nil
	}

	return nil
}

func (ss *StoreService) DeleteTask(id string) error {
	res := ss.store.db.MustExec("DELETE FROM scheduler WHERE id = $1", id)
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Println("rows affected for id", id, "=", num)
	if num == 0 {
		return fmt.Errorf("no row was deleted for id %s", id)
	}
	return nil
}

func (ss *StoreService) GetAllTasks(limit int) (*[]Task, error) {
	tasks := []Task{}
	err := ss.store.db.Select(&tasks, "SELECT * FROM scheduler LIMIT $1", limit)
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (ss *StoreService) SearchByDate(date time.Time, limit int) (*[]Task, error) {
	dateForsearch := date.Format(dateFormat)
	tasks := []Task{}
	fmt.Println("Tasks by date for date:", dateForsearch)
	err := ss.store.db.Select(&tasks, "SELECT * FROM scheduler WHERE date = $1 LIMIT $2", dateForsearch, limit)
	if err != nil {
		fmt.Println("Tasks by date error:", err.Error())
		return nil, err
	}

	return &tasks, nil
}

func (ss *StoreService) SearchByString(search string, limit int) (*[]Task, error) {
	tasks := []Task{}
	err := ss.store.db.Select(&tasks, "SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE '$1' ORDER BY date LIMIT $2",
		"%"+search+"%", limit)
	if err != nil {
		fmt.Println("Tasks by string error:", err.Error())
		return nil, err
	}

	return &tasks, nil
}
