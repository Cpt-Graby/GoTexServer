package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
)

const baseDir = "./serving_files"

type downloadHandler struct {
	root *os.Root
}

func (dlH *downloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("path")

	if !fs.ValidPath(fileName) {
		fmt.Println("invalid path:", fileName)
		http.NotFound(w, r)
		return
	}

	f, err := dlH.root.Open(fileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			fmt.Println("cannot open file:", fileName, err)
		}
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		fmt.Println("cannot stat file:", fileName, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !info.Mode().IsRegular() {
		fmt.Println("not a regular file:", fileName)
		http.NotFound(w, r)
		return
	}

	params := map[string]string{"filename": path.Base(fileName)}
	if cd := mime.FormatMediaType("attachment", params); cd != "" {
		w.Header().Set("Content-Disposition", cd)
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")

	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func createMux(root *os.Root) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /dl/{path...}", &downloadHandler{root: root})
	return mux
}

func main() {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		log.Fatal(err)
	}
	defer root.Close()
	log.Fatal(http.ListenAndServe(":8080", createMux(root)))
}
