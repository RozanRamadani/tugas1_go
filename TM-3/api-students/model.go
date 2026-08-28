package main

import "time"

// Student adalah data utama mahasiswa.
type Student struct {
	ID       int       `json:"id"`
	NIM      string    `json:"nim"`
	Name     string    `json:"name"`
	Grade    float64   `json:"grade"`
	IsActive bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// POST /students
// Semua field utama wajib dikirim.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PUT /students/:id
// PUT mengganti seluruh data.
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PATCH /students/:id
// Pointer yang digunakan untuk membedakan:
// field tidak dikirim (nil)
// dengan field dikirim tetapi bernilai tertentu.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// Format response API.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Metadata untuk pagination.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Query parameter GET /students.
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}