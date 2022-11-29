package services

import (
	"fmt"
	"image"
	"image/color"

	"github.com/Samar2170/Blackbox/models"
	"github.com/disintegration/imaging"
)

var validExtensions = map[string]struct{}{
	"jpg":  struct{}{},
	"jpeg": struct{}{},
	"png":  struct{}{},
}

func CreateThumbnail(fmdId uint) error {
	fmd, err := models.GetFileMetaDataById(fmdId)
	if err != nil {
		return err
	}
	if _, ok := validExtensions[fmd.Extension]; !ok {
		return nil
	}
	var thumbnail image.Image
	img, err := imaging.Open(fmd.Path)
	if err != nil {
		return err
	}
	thumbnail = imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	dst := imaging.New(100, 100, color.NRGBA{0, 0, 0, 0})

	// paste thumbnails into the new image side by side
	dst = imaging.Paste(dst, thumbnail, image.Pt(0, 0))
	user, err := models.GetUserById(fmd.UserId)
	if err != nil {
		return err
	}
	// save the resulting image as JPEG
	fmtId := fmt.Sprintf("%d", fmdId)
	thumbnailPath := MAIN_DIR + "/" + user.Username + "/thumbnails/" + fmtId + ".png"
	err = imaging.Save(dst, thumbnailPath)
	if err != nil {
		return err
	}
	err = fmd.SaveThumbnailPath(thumbnailPath)
	if err != nil {
		return err
	}

	return nil

}
