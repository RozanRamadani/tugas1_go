//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

// drop mengirim DELETE /courses/:id/enroll atas nama student.
func (a api) drop(t *testing.T, student model.User, courseID int) int {
	t.Helper()

	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Fatalf("membuat token: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request drop: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

// swap mengirim POST /enrollments/swap atas nama student.
func (a api) swap(t *testing.T, student model.User, fromID, toID int) (int, string) {
	t.Helper()

	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Fatalf("membuat token: %v", err)
	}

	body, _ := json.Marshal(model.SwapRequest{
		FromCourseID: fromID,
		ToCourseID:   toID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments/swap", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request swap: %v", err)
	}
	defer resp.Body.Close()

	var res struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	return resp.StatusCode, res.Code
}

// D.2 Butir 1: POST /enrollments/swap saat kelas tujuan sudah penuh.
// Response harus 409 COURSE_FULL, mahasiswa TETAP terdaftar di kelas asal,
// enrolled_count kedua kelas tidak berubah, dan tidak ada event enrollment.swapped di outbox.
func TestEnrollEndpoint_SwapCourseFull(t *testing.T) {
	a := newAPI(t)
	ctx := context.Background()
	courseRepo := repository.NewCourseRepository(pool)

	fromCourse := seedCourse(t, 3, 40)
	toCourse := seedCourse(t, 3, 2) // kuota hanya 2
	student := seedStudent(t)

	// Mahasiswa mendaftar di kelas asal
	if status, _ := a.enroll(t, student, fromCourse.ID); status != http.StatusCreated {
		t.Fatalf("enroll kelas asal: status = %d, seharusnya 201", status)
	}

	// Penuhi kelas tujuan dengan 2 mahasiswa lain
	s1 := seedStudent(t)
	s2 := seedStudent(t)
	if status, _ := a.enroll(t, s1, toCourse.ID); status != http.StatusCreated {
		t.Fatalf("enroll siswa 1 ke kelas tujuan: status = %d, seharusnya 201", status)
	}
	if status, _ := a.enroll(t, s2, toCourse.ID); status != http.StatusCreated {
		t.Fatalf("enroll siswa 2 ke kelas tujuan: status = %d, seharusnya 201", status)
	}

	// Coba swap ke kelas tujuan yang sudah penuh
	status, code := a.swap(t, student, fromCourse.ID, toCourse.ID)
	if status != http.StatusConflict || code != "COURSE_FULL" {
		t.Fatalf("swap kelas penuh: dapat %d %s, seharusnya 409 COURSE_FULL", status, code)
	}

	// Verifikasi mahasiswa TETAP terdaftar di kelas asal
	enrolledFrom := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, fromCourse.ID)
	if enrolledFrom != 1 {
		t.Errorf("mahasiswa tidak lagi terdaftar di kelas asal: count = %d, seharusnya 1", enrolledFrom)
	}

	// Verifikasi mahasiswa TIDAK terdaftar di kelas tujuan
	enrolledTo := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, toCourse.ID)
	if enrolledTo != 0 {
		t.Errorf("mahasiswa terdaftar di kelas tujuan: count = %d, seharusnya 0", enrolledTo)
	}

	// Verifikasi enrolled_count kedua kelas tidak berubah
	gotFrom, err := courseRepo.FindByID(ctx, fromCourse.ID)
	if err != nil || gotFrom.EnrolledCount != 1 {
		t.Errorf("enrolled_count kelas asal = %d, seharusnya 1", gotFrom.EnrolledCount)
	}

	gotTo, err := courseRepo.FindByID(ctx, toCourse.ID)
	if err != nil || gotTo.EnrolledCount != 2 {
		t.Errorf("enrolled_count kelas tujuan = %d, seharusnya 2", gotTo.EnrolledCount)
	}

	// Verifikasi TIDAK ADA event enrollment.swapped di outbox untuk mahasiswa ini
	swapEvents := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentSwapped)
	if swapEvents != 0 {
		t.Errorf("event outbox enrollment.swapped = %d, seharusnya 0", swapEvents)
	}
}

// D.2 Butir 2: DELETE /courses/:id/enroll: berhasil keluar menghasilkan 204,
// baris enrollment hilang, enrolled_count berkurang satu, dan outbox berisi tepat satu event enrollment.dropped.
func TestEnrollEndpoint_DropSuccess(t *testing.T) {
	a := newAPI(t)
	ctx := context.Background()
	courseRepo := repository.NewCourseRepository(pool)

	course := seedCourse(t, 3, 10)
	student := seedStudent(t)

	// Mahasiswa mendaftar kelas terlebih dahulu
	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("enroll kelas: status = %d, seharusnya 201", status)
	}

	// Mahasiswa keluar dari kelas (Drop)
	status := a.drop(t, student, course.ID)
	if status != http.StatusNoContent {
		t.Fatalf("drop kelas: status = %d, seharusnya 204", status)
	}

	// Verifikasi baris enrollment hilang
	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, course.ID)
	if enrolled != 0 {
		t.Errorf("baris enrollments = %d, seharusnya 0 setelah drop", enrolled)
	}

	// Verifikasi enrolled_count berkurang satu (menjadi 0)
	got, err := courseRepo.FindByID(ctx, course.ID)
	if err != nil || got.EnrolledCount != 0 {
		t.Errorf("enrolled_count = %d, seharusnya 0 setelah drop", got.EnrolledCount)
	}

	// Verifikasi outbox berisi tepat satu event enrollment.dropped
	dropEvents := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentDropped)
	if dropEvents != 1 {
		t.Errorf("event outbox enrollment.dropped = %d, seharusnya 1", dropEvents)
	}
}
