package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	Id      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type Empty struct{}

func handleNextDate(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Next date handler running")
	if req.Method == http.MethodGet {
		now := req.FormValue("now")
		if now == "" {
			sendError(res, "required parameter absent: 'now'", http.StatusBadRequest)
			return
		}

		date := req.FormValue("date")
		if date == "" {
			date = now
		}

		repeat := req.FormValue("repeat")
		if repeat == "" {
			sendError(res, "required parameter absent: 'repeat'", http.StatusBadRequest)
			return
		}

		nowDate, err := time.Parse(dateFormat, now)
		if err != nil {
			sendError(res, err.Error(), http.StatusBadRequest)
		}

		nextDate, err := NextDate(nowDate, date, repeat)
		if err != nil {
			sendError(res, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprint(res, nextDate)
	}

}

func handleTask(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Task handler running")

	// add tasks
	if req.Method == http.MethodPost {
		postTaskHandler(res, req)
	}

	// edit tasks
	if req.Method == http.MethodGet {
		getTaskHandler(res, req)
	}

	// edit tasks
	if req.Method == http.MethodPut {
		putTaskHandler(res, req)
	}

	// edit tasks
	if req.Method == http.MethodDelete {
		deleteTaskHandler(res, req)
	}

}

func postTaskHandler(res http.ResponseWriter, req *http.Request) {
	var task Task
	var buf bytes.Buffer
	var nextDate string
	var err error

	_, err = buf.ReadFrom(req.Body)
	if err != nil {
		sendError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		sendError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(res, "task title must be specified", http.StatusBadRequest)
		return
	}

	today := TodayDate(time.Now())
	todayStr := today.Format(DATE_SCHEDULE_FORMAT)

	if task.Date == "" {
		nextDate = todayStr
	} else {
		taskDate, err := time.Parse(DATE_SCHEDULE_FORMAT, task.Date)
		if err != nil {
			sendError(res, "wrong format for date", http.StatusBadRequest)
			return
		}
		if task.Repeat != "" {
			if taskDate.Before(today) {
				nextDate, err = NextDate(today, task.Date, task.Repeat)
				if err != nil {
					sendError(res, "wrong format for date", http.StatusBadRequest)
					return
				}
			} else {
				nextDate = task.Date
			}
		} else {
			if taskDate.Before(today) {
				nextDate = todayStr
			} else {
				nextDate = task.Date
			}
		}
	}

	storeService := GetStoreService()
	if storeService == nil {
		sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
		return
	}

	id, err := storeService.AddTask(nextDate, task.Title, task.Comment, task.Repeat)
	if err != nil {
		sendError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	sendOk(res, map[string]any{"id": id})
}

func getTaskHandler(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		sendError(res, "required parameter id must be specified", http.StatusBadRequest)
		return
	}
	storeService := GetStoreService()
	if storeService == nil {
		sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
		return
	}
	task, err := storeService.GetTaskById(id)
	if err != nil {
		sendError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if task == nil {
		sendOk(res, Task{})
		return
	}
	sendOk(res, task)
}

func putTaskHandler(res http.ResponseWriter, req *http.Request) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		sendError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		sendError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendError(res, "task title must be specified", http.StatusBadRequest)
		return
	}

	today := TodayDate(time.Now())

	nextDate, err := NextDate(today, task.Date, task.Repeat)
	if err != nil {
		sendError(res, err.Error(), http.StatusBadRequest)
		return
	}

	storeService := GetStoreService()
	if storeService == nil {
		sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
		return
	}

	task.Date = nextDate
	err = storeService.UpdateTask(&task)
	if err != nil {
		sendError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	sendOk(res, Task{})
}

func deleteTaskHandler(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		sendError(res, "parameter 'id' must be specified", http.StatusBadRequest)
		return
	}
	fmt.Println("Getting store service", id)
	storService := GetStoreService()
	if storService == nil {
		sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
		return
	}
	fmt.Println("Delete task running for id", id)
	err := storService.DeleteTask(id)
	if err != nil {
		sendError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	sendOk(res, Empty{})
}

func doneTaskHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Done task handler running")
	if req.Method == http.MethodPost {
		id := req.FormValue("id")
		if id == "" {
			sendError(res, "parameter 'id' must be specified", http.StatusBadRequest)
			return
		}
		storeService := GetStoreService()
		if storeService == nil {
			sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
			return
		}
		err := storeService.RescheduleTask(id)
		if err != nil {
			sendError(res, err.Error(), http.StatusInternalServerError)
			return
		}
		sendOk(res, Empty{})
	}
}

func handleTasks(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Tasks handler running")
	if req.Method == http.MethodGet {
		search := req.FormValue("search")
		storeService := GetStoreService()
		if storeService == nil {
			sendError(res, "cannot get instance of store service", http.StatusInternalServerError)
			return
		}
		tasks, err := storeService.GetTasks(search, LIMIT)
		if err != nil {
			sendError(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if tasks == nil || len(*tasks) == 0 {
			sendOk(res, map[string]any{"tasks": []Task{}})
			return
		}
		sendOk(res, map[string]any{"tasks": tasks})
	}
}

func ListenApi(port string) {
	webDir := "./web"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	mux.HandleFunc("/api/nextdate", handleNextDate)
	mux.HandleFunc("/api/task", handleTask)
	mux.HandleFunc("/api/tasks", handleTasks)
	mux.HandleFunc("/api/task/done", doneTaskHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil {
		panic(err)
	}
}

func sendError(w http.ResponseWriter, message string, status int) {
	sendResponse(w, map[string]string{"error": message}, status)
}

func sendOk(w http.ResponseWriter, data interface{}) {
	sendResponse(w, data, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data interface{}, status int) {
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(resp)
}
