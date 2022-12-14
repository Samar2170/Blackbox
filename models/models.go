package models

import (
	"errors"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func Connect() {
	var err error
	Db, err = gorm.Open(postgres.Open(DBURI), &gorm.Config{})
	if err != nil {
		err2 := errors.New("could not connect to the database")
		log.Println(err2)

	}
	Db.AutoMigrate(User{})
	Db.AutoMigrate(FileMetaData{})

}

type User struct {
	*gorm.Model
	ID           uint   `gorm:"PrimaryIndex"`
	Username     string `gorm:"Unique"`
	Password     string
	bucketStatus bool
}

type FileMetaData struct {
	*gorm.Model
	ID            uint `gorm:"PrimaryIndex"`
	UserId        uint
	User          User `gorm:"foreignKey:UserId;references:ID;"`
	OgName        string
	NewName       string
	Extension     string
	Size          uint // store in megabytes
	Path          string
	SignedUrl     string `gorm:"uniqueIndex"`
	ThumbnailPath string
}

func (u *User) Create() error {
	err := Db.Create(&u).Error
	return err
}
func (u *User) CreateBucket() error {
	err := CreateDirectoryForUser(u.Username)
	if err != nil {
		return err
	}
	u.bucketStatus = true
	err = Db.Save(&u).Error
	return err
}
func (fmd *FileMetaData) SaveThumbnailPath(thumbnailPath string) error {
	fmd.ThumbnailPath = thumbnailPath
	err := Db.Save(&fmd).Error
	return err
}

func GetUserByUsername(username string) (User, error) {
	var u User
	err := Db.Where("username = ?", username).First(&u).Error
	return u, err
}
func GetAllUsernames() ([]string, error) {
	var usernames []string
	err := Db.Model(&User{}).Pluck("username", &usernames).Error
	return usernames, err
}

func GetUserById(id uint) (User, error) {
	var u User
	err := Db.Where("id = ?", id).First(&u).Error
	return u, err
}

func (f *FileMetaData) Create() error {
	err := Db.Create(&f).Error
	return err
}

func GetFileBySignedUrl(signedUrl string) (FileMetaData, error) {
	var fm FileMetaData
	err := Db.Where("signed_url = ?", signedUrl).First(&fm).Error
	return fm, err
}

func GetFileBySignedUrlUser(signedUrl string, userId uint) (FileMetaData, error) {
	var fm FileMetaData
	err := Db.Where("signed_url = ? user_id = ?", signedUrl, userId).First(&fm).Error
	return fm, err
}

func GetFileMetaDataById(id uint) (FileMetaData, error) {
	var fm FileMetaData
	err := Db.Where("id = ?", id).First(&fm).Error
	return fm, err
}

func GetFileMetaDataWOThumbnails() ([]FileMetaData, error) {
	var fmd []FileMetaData
	err := Db.Where("thumbnail_path IS NULL").Find(&fmd).Error
	return fmd, err

}
func GetFileMetaDataByUserID(userId uint) ([]FileMetaData, error) {
	var fmds []FileMetaData
	err := Db.Where("user_id = ?", userId).Find(&fmds).Error
	return fmds, err
}
