package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const baseDir = "serving_files"

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

	if !fs.ValidPath(fileName) {
		fmt.Println("invalid path:", fileName)
		http.NotFound(w, r)
		return
	}
	fullPath := filepath.Join(baseDir, fileName)

	exists, err := fileExists(fullPath)
	if err != nil {
		fmt.Println("cannot check file:", fullPath, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("file does not exist:", fullPath)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filefileName=%q", fileName))
	http.ServeFile(w, r, fullPath)
}

func main() {
	http.HandleFunc("/dl/", dlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
