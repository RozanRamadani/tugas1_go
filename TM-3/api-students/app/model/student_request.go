package model

// CreateStudentRequest adalah data yang diterima
// ketika client membuat student baru.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest digunakan untuk PUT.
// PUT mengganti seluruh data student.
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest digunakan untuk PATCH.
// Pointer digunakan supaya kita bisa membedakan:
//
// field tidak dikirim  → nil
// field dikirim false  → pointer ke false
// field dikirim 0      → pointer ke 0
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
