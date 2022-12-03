package main

import (
	"encoding/json"
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
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func returnJson(w http.ResponseWriter, data interface{}) {
	jData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error while encoding json"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func login(w http.ResponseWriter, r *http.Request) {
	var lr loginRequest
	data, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(data, &lr)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while parsing request"})
		return
	}
	username, password := lr.Username, lr.Password

	if username == "" || password == "" {
		returnJson(w, map[string]string{"status": "Username or password cannot be empty"})
		return
	}

	user, err := models.GetUserByUsername(username)
	if err != nil {
		returnJson(w, map[string]string{"status": "User does not exist"})
		return
	}
	if user.Password != password {
		returnJson(w, map[string]string{"status": "Incorrect password"})
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
		returnJson(w, map[string]string{"status": "Error while signing token"})
		return
	}
	respMap := map[string]string{
		"access_token": t,
		"username":     user.Username,
	}
	returnJson(w, respMap)
}

func signup(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	_, err := models.GetUserByUsername(username)
	if err == nil {
		returnJson(w, map[string]string{"status": "User already exists"})
		return
	}
	user := models.User{
		Username: username,
		Password: password,
	}
	err = user.Create()
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while creating user"})
		return
	}
	err2 := user.CreateBucket()
	if err2 != nil {
		returnJson(w, map[string]string{"status": "Error while creating bucket"})
		return
	}
	response := map[string]string{"status": "User created successfully"}
	json.NewEncoder(w).Encode(response)
}

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
}

func upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	userId := r.Header.Get("userId")
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while fetching file"})
		return
	}

	defer file.Close()
	fmt.Println(header.Filename)
	fmt.Println(header.Size)
	fmt.Println(header.Header)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while parsing userId"})
		return
	}
	user, err := models.GetUserById(uint(userIdInt))
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while fetching user"})
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
		returnJson(w, map[string]string{"status": "Error while saving file"})
		return
	}
	err = fm.Create()
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			returnJson(w, map[string]string{"status": "File already exists"})
			return
		}
		returnJson(w, map[string]string{"status": "Error while saving file metadata"})
		return
	}

	returnJson(w, map[string]string{"status": "File uploaded successfully"})

}

func uploads(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while parsing userId"})
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while parsing form"})
		return
	}
	files := r.MultipartForm.File["files"]

	models.SaveFiles(files, userIdInt)

	returnJson(w, map[string]string{"status": "Files uploaded successfully"})
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	urlSplit := strings.Split(r.URL.Path, "/")
	fileHash := urlSplit[len(urlSplit)-1]
	fmd, err := models.GetFileBySignedUrl(fileHash)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while fetching file"})
		return
	}
	file, err := os.Open(fmd.Path)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while opening file"})
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

func viewFiles(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("userId")
	userIdInt, err := strconv.ParseInt(userId, 0, 32)
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while parsing userId"})
		return
	}
	fmds, err := models.GetFileMetaDataByUserID(uint(userIdInt))
	if err != nil {
		returnJson(w, map[string]string{"status": "Error while fetching files"})
		return
	}
	returnJson(w, fmds)
}
