package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DbName = "scheduler.db"
	DbFile = "./" + DbName
	Schema = `CREATE TABLE IF NOT EXISTS "scheduler" (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		date CHAR(8) NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat VARCHAR(128)
	);`
)

type Store struct {
	db *sqlx.DB
}

func openDB() *sqlx.DB {
	dbfile := DbFile
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
			db.MustExec(Schema)
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

func (s *Store) AddTask(nextDate string, title string, comment string, repeat string) (int64, error) {
	res, err := s.db.NamedExec("INSERT INTO scheduler (date, title, comment, repeat) values (:date, :title, :comment, :repeat)",
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

func (s *Store) GetTaskById(id string) (*Task, error) {
	task := Task{}
	err := s.db.Get(&task, "SELECT * FROM scheduler WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *Store) UpdateTask(task *Task) error {
	res, err := s.db.NamedExec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
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

func (s *Store) GetById(id string) (*Task, error) {
	task := Task{}
	err := s.db.Get(&task, "SELECT * FROM scheduler WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) UpdateTaskDate(date string, id string) error {
	res := s.db.MustExec("UPDATE scheduler SET date = $1 WHERE id = $2", date, id)
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if num == 0 {
		return fmt.Errorf("no row was affected for id = %s", id)
	}
	return nil
}

func (s *Store) DeleteTask(id string) error {
	res := s.db.MustExec("DELETE FROM scheduler WHERE id = $1", id)
	num, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if num == 0 {
		return fmt.Errorf("no row was deleted for id %s", id)
	}
	return nil
}

func (s *Store) GetAllTasks(limit int) (*[]Task, error) {
	tasks := []Task{}
	err := s.db.Select(&tasks, "SELECT * FROM scheduler LIMIT $1", limit)
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (s *Store) SearchByDate(date string, limit int) (*[]Task, error) {
	tasks := []Task{}
	err := s.db.Select(&tasks, "SELECT * FROM scheduler WHERE date = $1 LIMIT $2", date, limit)
	if err != nil {
		return nil, fmt.Errorf("tasks by date error: %s", err.Error())
	}

	return &tasks, nil
}

func (s *Store) SearchByString(search string, limit int) (*[]Task, error) {
	tasks := []Task{}
	err := s.db.Select(&tasks, "SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE '$1' ORDER BY date LIMIT $2",
		"%"+search+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("tasks by string error: %s", err.Error())
	}

	return &tasks, nil
}
