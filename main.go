package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func dlHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len("/dl/"):]
	exists, err := fileExists("./" + fileName)
	if err != nil {
		fmt.Println("cannot check file:", fileName, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("file does not exist:", fileName)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filefileName=%q", fileName))
	http.ServeFile(w, r, fileName)
}

func main() {
	http.HandleFunc("/dl/", dlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
