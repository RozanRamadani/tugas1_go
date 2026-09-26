package main

import (
	"context"
	"fmt"
	"log"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/database"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal("gagal connect db:", err)
	}
	defer pool.Close()

	// Apply migration 005 manually here for test
	_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx ON students (created_at DESC, id DESC);`)
	if err != nil {
		log.Fatal("gagal migrate:", err)
	}

	repo := repository.NewStudentRepository(pool)

	fmt.Println("--- MENGAMBIL HALAMAN 1 ---")
	q1 := model.CursorQuery{Limit: 2}
	students1, err := repo.FindAfterCursor(ctx, q1)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range students1 {
		fmt.Printf("- %s | %s\n", s.ID, s.Name)
	}

	if len(students1) == 0 {
		fmt.Println("Database kosong. Uji cursor dilewati.")
	} else {
		last := students1[0]
		if len(students1) > 1 {
			last = students1[1]
		}
		
		fmt.Println("\n--- MENGAMBIL HALAMAN 2 DENGAN CURSOR ---")
		cursor := model.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		q2 := model.CursorQuery{Limit: 2, After: &cursor}
		students2, err := repo.FindAfterCursor(ctx, q2)
		if err != nil {
			log.Fatal(err)
		}
		for _, s := range students2 {
			fmt.Printf("- %s | %s\n", s.ID, s.Name)
		}
	}

	fmt.Println("\n--- EXPLAIN ANALYZE HALAMAN 1 ---")
	rows, err := pool.Query(ctx, "EXPLAIN ANALYZE SELECT id, nim, name, grade, is_active, COALESCE(owner_id, 0), created_at FROM students WHERE 1 = 1 ORDER BY created_at DESC, id DESC LIMIT 3")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var plan string
		rows.Scan(&plan)
		fmt.Println(plan)
	}

	fmt.Println("\n--- EXPLAIN ANALYZE DENGAN CURSOR ---")
	rows2, err := pool.Query(ctx, "EXPLAIN ANALYZE SELECT id, nim, name, grade, is_active, COALESCE(owner_id, 0), created_at FROM students WHERE 1 = 1 AND (created_at, id) < ('2026-09-26 00:00:00', 'S001') ORDER BY created_at DESC, id DESC LIMIT 3")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var plan string
		rows2.Scan(&plan)
		fmt.Println(plan)
	}
}
