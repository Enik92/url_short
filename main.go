package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"time"
)

type Linker struct {
	ID int `json:"id"`
	Link string `json:"link"`
	Shortlink string `json:"shortlink"`
	Lifetime  time.Duration `json:"lifetime"`
	Createdate time.Time `json:"create_date"`
	Expiresdate time.Time `json:"expires_date"`
}

type Result struct {
	Link string
	Short string
	Create_date time.Time
	Expires_date time.Time
	Lifetime time.Duration
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func shortener() string {
	rand.Seed(time.Now().UnixNano())
	bsl := make([]byte, 5)
	for i := range bsl {
		bsl[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(bsl)
}

func save (w http.ResponseWriter, r *http.Request) {
	link := new(Linker)
	result := Result{}
	if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&link)
		result.Link = link.Link
		result.Lifetime = link.Lifetime
		result.Short = shortener()
		result.Create_date = time.Now()
		result.Expires_date = result.Create_date.Add(time.Second * result.Lifetime)
		db, err := sql.Open("sqlite3", "linker_go.db")
		if err != nil {
			panic(err)
		}
		defer db.Close()
		db.Exec(
			"insert into links (link, shortlink, lifetime, create_date, expires_date) values ($1, $2, $3, $4, $5)",
			result.Link,
			result.Short,
			result.Lifetime,
			result.Create_date,
			result.Expires_date,
		)

		fmt.Fprintf(w, "your URL:%s, your shortlink:%s, lifetime:%d , create_date: %v, expires_date: %v", link.Link,result.Short, link.Lifetime, result.Create_date, result.Expires_date)
	}
}

func redirect (w http.ResponseWriter, r *http.Request){
	var url string
	vars := mux.Vars(r)
	db, err := sql.Open("sqlite3", "linker_go.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	rows := db.QueryRow("select link from links where shortlink =$1 AND expires_date < date('now') limit 1 ", vars["key"])
	rows.Scan(& url)
	http.Redirect(w, r, url, 302)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/save", save) // получаем в пост запросе урл, укорачиваем и сохраняем его в базу
	router.HandleFunc("/{key}", redirect) // редиректим по короткому урлу
	http.ListenAndServe(":8000", router)
}