package models

import (
	"log"
	"math"

	"gorm.io/gorm"
)

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"totalRows"`
	TotalPages int         `json:"totalPages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

func paginate(value interface{}, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	log.Println(totalPages)

	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

type CategoryGorm struct {
	db *gorm.DB
}

func (cg *CategoryGorm) List(pagination Pagination) (*Pagination, error) {
	var categories []*CategoryGorm

	cg.db.Scopes(paginate(categories, &pagination, cg.db)).Find(&categories)
	pagination.Rows = categories

	return &pagination, nil
}

func GetFilesDataByUserIdAdvanced(userId uint, pagination Pagination) (*Pagination, []*FileMetaData, error) {
	var files []*FileMetaData

	err := Db.Scopes(paginate(files, &pagination, Db)).Where("user_id = ?", userId).Find(&files).Error
	return &pagination, files, err
}
