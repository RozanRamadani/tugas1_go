package main

import (
	"fmt"
	"siakad-mini/config"
	"siakad-mini/domain"
)

func main() {
	config.LoadConfig()
	config.ConnectDB()

	var enrollments []domain.Enrollment
	config.DB.Find(&enrollments)
	for _, e := range enrollments {
		fmt.Printf("ID: %d, StudentID: %d, CourseID: %d\n", e.ID, e.StudentID, e.CourseID)
	}
}
