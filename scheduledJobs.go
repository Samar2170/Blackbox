package main

import (
	"log"

	"github.com/Samar2170/Blackbox/models"
)

func CreateThumbnails() {
	fmd, err := models.GetFileMetaDataWOThumbnails()
	if err != nil {
		log.Println(err)
	}
	for _, fmd := range fmd {
		thumbnailJob := ThumbnailJob{FmdId: fmd.ID}
		JobQueue <- &thumbnailJob
	}
}
