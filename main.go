package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

var POSTGRES_HOST string
var POSTGRES_PORT string
var POSTGRES_USER string
var POSTGRES_PASS string
var POSTGRES_SSL string
var POSTGRES_DB string

var LIVE_ADDRESS string
var LIVE_LIST_SYTLE string
var LIVE_STREAM_SYTLE string

var listStmt *sql.Stmt
var streamStmt *sql.Stmt

const xmlHeader = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><?xml-stylesheet type=\"text/xsl\" href=\"%s\"?>"

type Stream struct {
	Path        string
	Channel     string
	Title       string
	Description string
}

func main() {
	flag.StringVar(&POSTGRES_HOST, "pg_host", "db", "Postgres host.")
	flag.StringVar(&POSTGRES_PORT, "pg_port", "5432", "Postgres port.")
	flag.StringVar(&POSTGRES_USER, "pg_user", "postgres", "Postgres user.")
	flag.StringVar(&POSTGRES_PASS, "pg_pass", "", "Postgres password.")
	flag.StringVar(&POSTGRES_SSL, "pg_ssl", "disable", "Postgres SSL mode.")
	flag.StringVar(&POSTGRES_DB, "pg_db", "postgres", "Postgres database name.")
	flag.StringVar(&LIVE_ADDRESS, "live_addr", ":8080", "Live server address.")
	flag.StringVar(&LIVE_LIST_SYTLE, "live_list_style", "streams.xsl", "Live listing style.")
	flag.StringVar(&LIVE_STREAM_SYTLE, "live_stream_style", "stream.xsl", "Live stream sytle.")
	flag.Parse()

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASS, POSTGRES_DB, POSTGRES_SSL))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	listStmt, err = db.Prepare("SELECT path, channel, title, description FROM streams")
	if err != nil {
		panic(err)
	}
	streamStmt, err = db.Prepare("SELECT path, channel, title, description FROM streams WHERE streams.path=$1")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", list).Methods("GET")
	router.HandleFunc("/{stream}", stream).Methods("GET")
	http.ListenAndServe(LIVE_ADDRESS, router)
}

func list(w http.ResponseWriter, r *http.Request) {
	rows, err := listStmt.Query()
	if err != nil {
		panic(err)
	}

	var streams []Stream

	for rows.Next() {
		var stream Stream
		if err := rows.Scan(&stream.Path, &stream.Channel, &stream.Title, &stream.Description); err != nil {
			panic(err)
		}
		streams = append(streams, stream)
	}

	w.Write([]byte(fmt.Sprintf(xmlHeader, LIVE_LIST_SYTLE)))
	w.Write([]byte("<Streams>"))

	data, err := xml.Marshal(streams)
	if err != nil {
		panic(err)
	}
	w.Write(data)
	w.Write([]byte("</Streams>"))
}

func stream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stream Stream
	if err := streamStmt.QueryRow(vars["stream"]).Scan(&stream.Path, &stream.Channel, &stream.Title, &stream.Description); err != nil {
		w.WriteHeader(404)
		return
	}

	w.Write([]byte(fmt.Sprintf(xmlHeader, LIVE_STREAM_SYTLE)))

	data, err := xml.Marshal(stream)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}
