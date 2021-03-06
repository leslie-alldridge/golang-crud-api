package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Todos []Todo

var mainDB *sql.DB

func main() {

	db, errOpenDB := sql.Open("sqlite3", "todo.db")
	checkErr(errOpenDB)
	mainDB = db

	r := pat.New()
	r.Get("/", http.HandlerFunc(rootRoute))
	r.Del("/todos/:id", http.HandlerFunc(deleteByID))
	r.Get("/todos/:id", http.HandlerFunc(getByID))
	r.Put("/todos/:id", http.HandlerFunc(updateByID))
	r.Get("/todos", http.HandlerFunc(getAll))
	r.Post("/todos", http.HandlerFunc(insert))

	http.Handle("/", r)

	var port string

	port = ":" + os.Getenv("PORT")
	
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("ListenAndServeError:", err)
	}
}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todos", http.StatusFound)
}

func getAll(w http.ResponseWriter, r *http.Request) {
	rows, err := mainDB.Query("SELECT * FROM todos")
	checkErr(err)
	var todos Todos
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.Name)
		checkErr(err)
		todos = append(todos, todo)
	}
	jsonB, errMarshal := json.Marshal(todos)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func getByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	stmt, err := mainDB.Prepare(" SELECT * FROM todos where id = ?")
	checkErr(err)
	rows, errQuery := stmt.Query(id)
	checkErr(errQuery)
	var todo Todo
	for rows.Next() {
		err = rows.Scan(&todo.ID, &todo.Name)
		checkErr(err)
	}
	jsonB, errMarshal := json.Marshal(todo)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func insert(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	var todo Todo
	todo.Name = name
	stmt, err := mainDB.Prepare("INSERT INTO todos(name) values (?)")
	checkErr(err)
	result, errExec := stmt.Exec(todo.Name)
	checkErr(errExec)
	newID, errLast := result.LastInsertId()
	checkErr(errLast)
	todo.ID = newID
	jsonB, errMarshal := json.Marshal(todo)
	checkErr(errMarshal)
	fmt.Fprintf(w, "%s", string(jsonB))
}

func updateByID(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	id := r.URL.Query().Get(":id")
	var todo Todo
	ID, _ := strconv.ParseInt(id, 10, 0)
	todo.ID = ID
	todo.Name = name
	stmt, err := mainDB.Prepare("UPDATE todos SET name = ? WHERE id = ?")
	checkErr(err)
	result, errExec := stmt.Exec(todo.Name, todo.ID)
	checkErr(errExec)
	rowAffected, errLast := result.RowsAffected()
	checkErr(errLast)
	if rowAffected > 0 {
		jsonB, errMarshal := json.Marshal(todo)
		checkErr(errMarshal)
		fmt.Fprintf(w, "%s", string(jsonB))
	} else {
		fmt.Fprintf(w, "{row_affected=%d}", rowAffected)
	}

}

func deleteByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	stmt, err := mainDB.Prepare("DELETE FROM todos WHERE id = ?")
	checkErr(err)
	result, errExec := stmt.Exec(id)
	checkErr(errExec)
	rowAffected, errRow := result.RowsAffected()
	checkErr(errRow)
	fmt.Fprintf(w, "{row_affected=%d}", rowAffected)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
