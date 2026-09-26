package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"errors"
)

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func dlHandler(w http.ResponseWriter, r *http.Request){
	title := r.URL.Path[len("/dl/"):]
	if !checkFileExists("./" + title) {
		fmt.Println("file not exists ", title)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", title))
	http.ServeFile(w, r, title)
}


func main() {
	http.HandleFunc("/dl/", dlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
