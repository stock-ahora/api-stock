package main

import (
    "log"
    "net/http"

    "github.com/stock-ahora/api-stock/internal/http"
)

func main() {
    r := httpserver.NewRouter() // ver internal/http/router.go

    addr := ":8080"
    log.Printf("API listening on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        log.Fatal(err)
    }
}