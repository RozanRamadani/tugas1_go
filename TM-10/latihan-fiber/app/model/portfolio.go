package model

import "time"

// Portfolio adalah entity portofolio mahasiswa yang disimpan di MongoDB.
// Package model tetap bersih dari tipe driver seperti bson.ObjectID.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    string        `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}

type Certificate struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

type UpsertPortfolioRequest struct {
	Headline     string                    `json:"headline" validate:"required,min=5,max=150"`
	Bio          string                    `json:"bio" validate:"required,min=10,max=2000"`
	Projects     []CreateProjectRequest    `json:"projects" validate:"omitempty,max=20,dive"`
	Certificates []CreateCertificateRequest `json:"certificates" validate:"omitempty,max=20,dive"`
}

type CreateProjectRequest struct {
	Title        string   `json:"title" validate:"required,min=3,max=100"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	Link         string   `json:"link" validate:"omitempty,url"`
	Year         int      `json:"year" validate:"required,min=2000,max=2100"`
}

type CreateCertificateRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=100"`
	Issuer string `json:"issuer" validate:"required,min=2,max=100"`
	Date   string `json:"date" validate:"required"` // Format: YYYY-MM-DD
}

type TechnologyStat struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
