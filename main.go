package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/subosito/gotenv"
	"log"
	"net/http"


	"os"

	"github.com/gorilla/mux"
)

type Book struct {
	ID     int	  `json:id`
	Title  string `json:title`
	Author string `json:author`
	Year   string `json:year`
}

var books []Book
var db *sql.DB

func init() {
	gotenv.Load()
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	pqUrl, err := pq.ParseURL(os.Getenv("ELEPHANTSQL_URL"))
	logFatal(err)

	db, err = sql.Open("postgres", pqUrl)
	logFatal(err)

	err = db.Ping()
	logFatal(err)

	router := mux.NewRouter()

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", addBook).Methods("POST")
	router.HandleFunc("/books", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", removeBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", router))
	fmt.Println("Run Server Port  8000")
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	var book Book
	books = []Book{}

	rows, err := db.Query("SELECT * FROM books")
	logFatal(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		logFatal(err)

		books = append(books, book)
	}
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	params := mux.Vars(r)

	rows := db.QueryRow("SELECT * FROM books WHERE id=$1", params["id"])

	err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	logFatal(err)

	json.NewEncoder(w).Encode(book)
}

func addBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	var bookID int

	json.NewDecoder(r.Body).Decode(&book)
	err := db.QueryRow("INSERT INTO books (title, author, year) VALUES ($1, $2, $3) RETURNING id;",
		book.Title, book.Author, book.Year).Scan(&bookID)
	logFatal(err)

	json.NewEncoder(w).Encode(bookID)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	result, err := db.Exec("UPDATE books SET title=$1, author=$2, year=$3 WHERE id=$4 RETURNING id",
		&book.Title, &book.Author, &book.Year, &book.ID)
	rowsUpdated, err := result.RowsAffected()
	logFatal(err)

	json.NewEncoder(w).Encode(rowsUpdated)
}

func removeBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	result, err := db.Exec("DELETE FROM books WHERE id = $1", params["id"])
	logFatal(err)

	rowsDeleted, err := result.RowsAffected()
	logFatal(err)

	json.NewEncoder(w).Encode(rowsDeleted)
}
