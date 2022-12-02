package models

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
)

func CreateUploadsDirectory() {
	// Create uploads directory
	if _, err := os.Stat(MAIN_DIR); os.IsNotExist(err) {
		os.Mkdir(MAIN_DIR, 0755)
	}
}

func CreateDirectoryForUser(username string) error {
	// Create directory for user
	if _, err := os.Stat(MAIN_DIR + "/" + username); os.IsNotExist(err) {
		os.Mkdir(MAIN_DIR+"/"+username, 0755)
	}
	return nil
}

func CheckBadDirectories() {
	set := make(map[string]struct{})
	// Get all files
	files, err := os.ReadDir(MAIN_DIR)
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		if file.IsDir() {
			set[file.Name()] = struct{}{}
		}
	}
	// Get all users
	usernames, err := GetAllUsernames()
	if err != nil {
		log.Println(err)
	}
	for _, username := range usernames {
		if _, ok := set[username]; !ok {
			user, err := GetUserByUsername(username)
			if err != nil {
				log.Println(err)
			}
			err = user.CreateBucket()
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func SaveFile(fileForm io.Reader, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, fileForm)
	if err != nil {
		return err
	}
	return nil
}
func SaveFiles(fhs []*multipart.FileHeader, userId int) error {
	user, err := GetUserById(uint(userId))

	for _, file := range fhs {
		go func(file *multipart.FileHeader) {
			fileSizeinMB := file.Size / (1024 * 1024)
			ogFileName, fileExtension := strings.Split(file.Filename, ".")[0], file.Filename[strings.LastIndex(file.Filename, ".")+1:]
			fileName := fmt.Sprintf("%s_%s.%s", user.Username, ogFileName, fileExtension)
			filePath := MAIN_DIR + "/" + user.Username + "/" + fileName
			fileHash := sha256.New()
			fileHash.Write([]byte(filePath))
			fileHashString := fmt.Sprintf("%x", fileHash.Sum(nil))

			r, w := io.Pipe()
			go func() {
				defer w.Close()
				file, err := file.Open()
				if err != nil {
					log.Println(err)
				}
				defer file.Close()
				_, err = io.Copy(w, file)
				if err != nil {
					log.Println(err)
				}
			}()
			err = SaveFile(r, filePath)
			if err != nil {
				log.Println(err)
			}
			fileMetaData := FileMetaData{
				UserId:    uint(userId),
				OgName:    ogFileName,
				NewName:   fileName,
				Extension: fileExtension,
				Size:      uint(fileSizeinMB),
				Path:      filePath,
				SignedUrl: fileHashString,
			}
			err = fileMetaData.Create()
			if err != nil {
				log.Println(err)
			}

		}(file)
	}
	return nil
}
