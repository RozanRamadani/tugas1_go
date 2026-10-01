package main

import (
	"flag"
	"log"

	"siakad-mini/config"
	"siakad-mini/domain"
)

func main() {
	// Menyiapkan flag untuk membedakan antara migrate (default) dan rollback
	rollback := flag.Bool("rollback", false, "Rollback database tables")
	flag.Parse()

	// Inisialisasi Environment & Koneksi DB (tanpa menyalakan server web)
	config.LoadConfig()
	config.ConnectDB()
	db := config.DB

	if *rollback {
		log.Println("Rolling back migrations...")
		// Drop tabel secara berurutan sesuai relasi (child dulu, lalu parent)
		err := db.Migrator().DropTable(
			&domain.Enrollment{},
			&domain.Course{},
			&domain.Student{},
			&domain.User{},
		)
		if err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")
		return
	}

	log.Println("Running migrations...")
	// AutoMigrate membuat tabel dari model, menambah kolom yang hilang, dan menyesuaikan indeks/foreign key
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Student{},
		&domain.Course{},
		&domain.Enrollment{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration completed successfully")
}
