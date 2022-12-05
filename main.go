package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/Samar2170/Blackbox/models"
	"github.com/rs/cors"
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
	// openLogFile(logPath)
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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
	viewFilesHandler := http.HandlerFunc(viewFiles)
	mux.Handle("/view-files", checkAuth(viewFilesHandler))

	loggedMux := NewLogger(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	// Insert the middleware
	// handler := cors.Default().Handler(loggedMux)
	handler := c.Handler(loggedMux)
	log.Println("Listening on port 8080")
	http.ListenAndServe(":8080", handler)
}
