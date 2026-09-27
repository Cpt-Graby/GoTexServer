package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

const baseDir = "./serving_files"

var root *os.Root

func fileExists(filePath string) (bool, error) {
	_, err := root.Stat(filePath)
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

	exists, err := fileExists(fileName)
	if err != nil {
		fmt.Println("cannot check file:", fileName, err)
		http.NotFound(w, r)
		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !exists {
		fmt.Println("file does not exist:", fileName)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filefileName=%q", fileName))
	http.ServeFileFS(w, r, root.FS(), fileName)
}

func main() {
	var err error
	root, err = os.OpenRoot(baseDir)
	if err != nil {
		log.Fatal(err)
	}
	defer root.Close()

	http.HandleFunc("/dl/", dlHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
