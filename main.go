package main

import (
	"net/http"
	"sync"

	"github.com/Samar2170/Blackbox/models"
)

func init() {
	models.Connect()
	// models.CreateUploadsDirectory()
	// models.CheckBadDirectories()
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go runServer()

	wg.Add(1)
	go runSuperVisor()

	wg.Add(1)
	go CreateThumbnails()

	wg.Wait()

}

func runServer() {
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
