package main

import (
	"context"
	"fmt"
	"log"

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

	// Disable seqscan to force index scan if possible
	_, err = pool.Exec(ctx, `SET enable_seqscan = OFF;`)
	
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
