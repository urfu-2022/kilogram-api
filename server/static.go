package server

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

const (
	staticPath = "./static/"
)

var (
	fileDir     = http.Dir(staticPath)
	fileServer  = http.FileServer(fileDir)
	fileHandler = http.StripPrefix("/static/", fileServer)
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	file, handler, err := r.FormFile("file")

	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)

		http.Error(w, "error parse form", http.StatusInternalServerError)

		return
	}

	defer file.Close()

	filename := fmt.Sprintf("%d-%s", rand.Int(), handler.Filename)
	uploadPath := filepath.Join(staticPath, filename)
	dst, err := os.Create(uploadPath)

	if err != nil {
		http.Error(w, "error creating new file", http.StatusInternalServerError)

		return
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "error saving file", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"result": "%s"}`, filename)))
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	fileHandler.ServeHTTP(w, r)
}
