package model

import "time"

// ListQuery menyimpan parameter untuk
// filtering, search, sorting, dan pagination.
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Cursor menampung nilai unik dari baris terakhir
// yang diterima oleh client.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// CursorQuery menyimpan parameter untuk pagination berbasis cursor.
type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

// Offset menentukan berapa data yang dilewati.
//
// Contoh:
// page=1, limit=10 → offset 0
// page=2, limit=10 → offset 10
// page=3, limit=10 → offset 20
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

// Meta berisi informasi pagination.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CursorMeta berisi informasi pagination berbasis cursor.
type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
