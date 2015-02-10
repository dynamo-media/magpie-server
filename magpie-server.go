package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	port   = flag.Int("port", 3000, "Listening port")
	apiKey = flag.String("apiKey", "", "API key for uploading")
)

func isRequestAuthenticated(r *http.Request) bool {
	requestApiKey := r.Header.Get("X-Api-Key")
	return *apiKey == "" || requestApiKey == *apiKey
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		uploadHtml := `<html>
		<head><title>Upload file</title></head>
		<body>
			<form enctype="multipart/form-data" action="/upload" method="post">
      			<input type="file" name="file" />
      			<input type="submit" value="upload" />
			</form>
		</body></html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(uploadHtml))
		return
	}

	if r.Method == "POST" {
		if !isRequestAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

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

		_, err = io.Copy(f, file)
		if err != nil {
			log.Println("Error copying data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println("Uploaded file:", handler.Filename)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func main() {
	flag.Parse()

	*apiKey = strings.TrimSpace(*apiKey)
	if *apiKey == "" {
		log.Println("Warning: No `apiKey` flag was given! Anyone can upload files.")
	}

	http.HandleFunc("/upload", uploadHandler)
	fs := http.FileServer(http.Dir("artifacts"))
	http.Handle("/artifacts/", http.StripPrefix("/artifacts/", fs))

	log.Println("Listening on port", *port)
	// TODO: Serve HTTPS as well
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(*port), nil))
}
