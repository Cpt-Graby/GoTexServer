package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func checkFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func dlHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/dl/"):]
	if !checkFileExists("./" + name) {
		fmt.Println("file does not exists!", name)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	http.ServeFile(w, r, name)
}

func main() {
	http.HandleFunc("/dl/", dlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
