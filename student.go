package main

import "fmt"

// Struct untuk menyimpan data student
type Student struct {
	ID       string
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo hanya membaca data student
// Karena tidak mengubah data, digunakan value receiver
func (s Student) GetInfo() string {
	return fmt.Sprintf(
		"ID: %s, Name: %s, Grade: %.2f, Active: %t",
		s.ID,
		s.Name,
		s.Grade,
		s.IsActive,
	)
}

// UpdateGrade mengubah nilai student
// Karena mengubah data asli, digunakan pointer receiver
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate mengubah status student menjadi aktif
// Karena mengubah data asli, digunakan pointer receiver
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate mengubah status student menjadi tidak aktif
// Karena mengubah data asli, digunakan pointer receiver
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	// Membuat data Student
	student := Student{
		ID:       "001",
		Name:     "Rozan",
		Grade:    90.5,
		IsActive: true,
	}

	// Menampilkan informasi awal
	fmt.Println("Data awal:")
	fmt.Println(student.GetInfo())

	// Mengubah nilai
	student.UpdateGrade(95.0)

	// Menonaktifkan student
	student.Deactivate()

	fmt.Println("\nSetelah update:")
	fmt.Println(student.GetInfo())

	// Mengaktifkan kembali student
	student.Activate()

	fmt.Println("\nSetelah activate:")
	fmt.Println(student.GetInfo())
}