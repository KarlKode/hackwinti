package main

import (
    "backend/pkg/database"
    "backend/pkg/server"
    "database/sql"
    "log"
    "net/http"
)

type Server struct {
    db *sql.DB
}

const (
    ASSETS = "/assets/"
    PORT   = "8080"
    DBDRIVER = "postgres"
    DBHOST = "localhost"
    DBNNAME = "hz"
    DBUSER = "postgres"
    DBPASSWORD = ""
)

func main() {
    db := database.NewDB(DBDRIVER, DBHOST, DBNNAME, DBUSER, DBPASSWORD)
    s := server.NewServer(db)
    router := server.NewRouter(s, ASSETS)
    log.Printf("Starting HTTP server on :%s\n", PORT)
    err := http.ListenAndServe(":"+PORT, router)
    if err != nil {
        log.Fatal("ListenAndServe Error: ", err)
    }
}