package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}
type TaskS struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks := make([]Task, 0)

	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("select * from tasks")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Name, &task.Completed)
		tasks = append(tasks, task)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(tasks)
}

func getUncompletedTasks(w http.ResponseWriter, r *http.Request) {
	tasks := make([]TaskS, 0)
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("select id, name from tasks where completed = 0")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var task TaskS
		err := rows.Scan(&task.ID, &task.Name)

		tasks = append(tasks, task)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(tasks)
}
func getCompletedTasks(w http.ResponseWriter, r *http.Request) {
	tasks := make([]TaskS, 0)
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query("select id, name from tasks where completed = 1")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var task TaskS
		err := rows.Scan(&task.ID, &task.Name)
		tasks = append(tasks, task)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(tasks)
}
func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	w.Header().Set("Content-Type", "application/json")
	json.NewDecoder(r.Body).Decode(&task)
	result, err := db.Exec("insert into tasks (name, completed) values($1, 0)", task.Name)
	if err != nil {
		panic(err.Error())
	}

	temp, _ := result.LastInsertId()
	task.ID = int(temp)
	task.Completed = false
	json.NewEncoder(w).Encode(task)

}
func editTask(w http.ResponseWriter, r *http.Request) {
	var task TaskS
	var temp int
	w.Header().Set("Content-Type", "application/json")
	data := mux.Vars(r)
	json.NewDecoder(r.Body).Decode(&task)
	ind, _ := strconv.Atoi(data["id"])
	rows, err := db.Query("select id from tasks")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	//check if id exists

	found := false
	for rows.Next() {
		err := rows.Scan(&temp)
		if err != nil {
			panic(err.Error())
		}
		if temp == ind {
			found = true
		}
	}
	if !found {
		w.WriteHeader(404)
		return
	}
	_, err = db.Exec("update tasks set name = $1 where id = $2", task.Name, ind)
	if err != nil {
		panic(err.Error())
	}
	task.ID = ind
	json.NewEncoder(w).Encode(task)

}
func completeTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	w.Header().Set("Content-Type", "application/json")
	data := mux.Vars(r)
	ind, _ := strconv.Atoi(data["id"])

	_, err = db.Exec("update tasks set completed = 1 where id = $1", ind)
	if err != nil {
		panic(err.Error())
	}

	row := db.QueryRow("select * from tasks where id = $1", ind)
	err := row.Scan(&task.ID, &task.Name, &task.Completed)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(task)
}
func uncompleteTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	w.Header().Set("Content-Type", "application/json")
	data := mux.Vars(r)
	ind, _ := strconv.Atoi(data["id"])

	_, err = db.Exec("update tasks set completed = 0 where id = $1", ind)
	if err != nil {
		panic(err.Error())
	}

	row := db.QueryRow("select * from tasks where id = $1", ind)
	err := row.Scan(&task.ID, &task.Name, &task.Completed)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(task)
}
func deleteTask(w http.ResponseWriter, r *http.Request) {
	data := mux.Vars(r)
	ind, _ := strconv.Atoi(data["id"])

	_, err := db.Exec("delete from tasks where id = $1", ind)
	if err != nil {
		panic(err.Error())
	}
	w.WriteHeader(204)
}

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("sqlite3", "todolistdb.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/tasks", getAllTasks).Methods("GET")
	r.HandleFunc("/api/tasks/uncompleted", getUncompletedTasks).Methods("GET")
	r.HandleFunc("/api/tasks/completed", getCompletedTasks).Methods("GET")
	r.HandleFunc("/api/tasks", createTask).Methods("POST")
	r.HandleFunc("/api/tasks/{id}", editTask).Methods("PUT")
	r.HandleFunc("/api/tasks/complete/{id}", completeTask).Methods("POST")
	r.HandleFunc("/api/tasks/uncomplete/{id}", uncompleteTask).Methods("POST")
	r.HandleFunc("/api/tasks/{id}", deleteTask).Methods("DELETE")
	http.ListenAndServe(":8080", r)
}
