package model

type CreateStudentRequest struct {
	ID       string  `json:"id" validate:"required"`
	NIM      string  `json:"nim" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	NIM      *string  `json:"nim" validate:"omitnil"`
	Name     *string  `json:"name" validate:"omitnil"`
	Grade    *float64 `json:"grade" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active"`
}
