package model

import "time"

type Supervisor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Quota           int       `json:"quota"`
	SupervisedCount int       `json:"supervised_count"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
}

type Supervision struct {
	ID           int64     `json:"id"`
	StudentID    string    `json:"student_id"`
	SupervisorID int       `json:"supervisor_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// AssignSupervisorRequest dipakai endpoint POST /supervisors/:id/assign.
type AssignSupervisorRequest struct {
	StudentID string `json:"student_id" validate:"required"`
}

// SwitchSupervisorRequest dipakai endpoint POST /supervisions/switch.
type SwitchSupervisorRequest struct {
	StudentID      string `json:"student_id" validate:"required"`
	ToSupervisorID int    `json:"to_supervisor_id" validate:"required,min=1"`
}

// UpdateSupervisorQuotaRequest dipakai endpoint PATCH /supervisors/:id.
type UpdateSupervisorQuotaRequest struct {
	Quota   int `json:"quota" validate:"required,min=1,max=100"`
	Version int `json:"version" validate:"required,min=1"`
}
