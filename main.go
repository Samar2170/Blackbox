package main

import (
	"net/http"
)

func init() {
	connect()
	CreateUploadsDirectory()
	CheckBadDirectories()

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/signup", signup)

	testHandler := http.HandlerFunc(test)
	mux.Handle("/test", checkAuth(testHandler))
	uploadHandler := http.HandlerFunc(upload)
	mux.Handle("/upload", checkAuth(uploadHandler))
	uploadsHandler := http.HandlerFunc(uploads)
	mux.Handle("/uploads", checkAuth(uploadsHandler))
	downloadFileHandler := http.HandlerFunc(downloadFile)
	mux.Handle("/download", checkAuth(downloadFileHandler))
	http.ListenAndServe(":8080", mux)
}
