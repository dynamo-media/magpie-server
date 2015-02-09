package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		content, err := ioutil.ReadFile("./static/upload.html")
		if err != nil {
			log.Fatal("Error reading file: ", err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)

		return
	}

	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20)

		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println("File upload error - acessing `file` input.", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		f, err := os.OpenFile("./artifacts/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println("Error opening file in `./artifacts/`.", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		io.Copy(f, file)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func main() {
	log.Println("Starting magpie-repository server...")

	http.HandleFunc("/upload", uploadHandler)

	fs := http.FileServer(http.Dir("artifacts"))
	http.Handle("/artifacts/", http.StripPrefix("/artifacts/", fs))

	log.Println("Listening on port 3000!")

	// TODO: Serve HTTPS as well
	log.Fatal(http.ListenAndServe(":3000", nil))
}
