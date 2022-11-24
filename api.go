package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	"crypto/sha256"

	"github.com/Samar2170/Blackbox/models"
)

type jwtCustomClaims struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := models.GetUserByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not found"))
		return
	}
	if user.Password != password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect password"))
		return
	}
	claims := &jwtCustomClaims{
		user.ID,
		user.Username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while signing token"))
	}
	w.Write([]byte(t))
}

func signup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	_, err := models.GetUserByUsername(username)
	if err == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User already exists"))
		return
	}
	user := models.User{
		Username: username,
		Password: password,
	}
	err = user.Create()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while creating user"))
		return
	}
	err2 := user.CreateBucket()
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while creating bucket"))
		return
	}
	w.Write([]byte("User created successfully"))
}

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
}

func upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	userId := r.Header.Get("userId")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while reading file"))
		return
	}

	defer file.Close()
	fmt.Println(header.Filename)
	fmt.Println(header.Size)
	fmt.Println(header.Header)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while parsing userId"))
		return
	}
	user, err := models.GetUserById(uint(userIdInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while getting user"))
		return
	}
	fileSizeinMB := header.Size / (1024 * 1024)
	ogFileName, fileExtension := strings.Split(header.Filename, ".")[0], header.Filename[strings.LastIndex(header.Filename, ".")+1:]
	fileName := fmt.Sprintf("%s_%s.%s", user.Username, ogFileName, fileExtension)
	filePath := MAIN_DIR + "/" + user.Username + "/" + fileName
	fileHash := sha256.New()
	fileHash.Write([]byte(filePath))
	fileHashString := fmt.Sprintf("%x", fileHash.Sum(nil))

	fm := models.FileMetaData{
		NewName:   fileName,
		UserId:    uint(userIdInt),
		OgName:    header.Filename,
		Size:      uint(fileSizeinMB),
		Path:      filePath,
		SignedUrl: fileHashString,
		Extension: fileExtension,
	}
	// fmt.Println(fm)

	err = models.SaveFile(file, filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while saving file"))
		return
	}
	err = fm.Create()
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("File already exists and overwritten."))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while saving file metadata"))
		return
	}

	w.Write([]byte("File uploaded successfully. Can be accessed at " + fileHashString))

}

func uploads(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while parsing userId"))
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while parsing form"))
		return
	}
	files := r.MultipartForm.File["files"]

	models.SaveFiles(files, userIdInt)

	w.Write([]byte("Files uploaded successfully"))

}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	urlSplit := strings.Split(r.URL.Path, "/")
	fileHash := urlSplit[len(urlSplit)-1]
	fmd, err := models.GetFileBySignedUrl(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldnt find File"))
		return
	}
	file, err := os.Open(fmd.Path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("File Not Found"))
		return
	}
	defer file.Close()
	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		io.Copy(writer, file)
	}()

	w.Header().Set("Content-Disposition", "attachment; filename="+fmd.NewName)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	io.Copy(w, reader)

}
