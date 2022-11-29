package main

import (
	"log"

	"github.com/Samar2170/Blackbox/services"
)

type ThumbnailJob struct {
	FmdId uint
}

func (tj *ThumbnailJob) Do() {
	err := services.CreateThumbnail(tj.FmdId)
	if err != nil {
		log.Println(err)
	}
}
