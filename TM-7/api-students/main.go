package main

import (
	"context"
	"log"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

func main() {

	config.LoadEnv()

	if err := config.InitLogger(); err != nil {
		log.Fatal(
			"gagal menginisialisasi logger:",
			err,
		)
	}

	defer config.CloseLogger()

	pool, err := database.NewPool(
		context.Background(),
	)

	if err != nil {
		log.Fatal(
			"gagal terhubung ke database:",
			err,
		)
	}

	defer pool.Close()

	// ============================================================
	// REPOSITORY
	// ============================================================

	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)

	// ============================================================
	// PERMISSION
	// ============================================================

	rawPermissions, err := roleRepo.LoadPermissions(
		context.Background(),
	)

	if err != nil {
		log.Fatal(
			"gagal memuat permission:",
			err,
		)
	}

	permissionSet := helper.NewPermissionSet(
		rawPermissions,
	)

	log.Printf(
		"permission dimuat, roles=%v",
		permissionSet.KnownRoles(),
	)

	// ============================================================
	// SERVICE
	// ============================================================

	studentService := service.NewStudentService(
		studentRepo,
		permissionSet,
	)

	jwtSecret := config.GetEnv(
		"JWT_SECRET",
		"super-secret-key-praktikum-backend-32byte",
	)

	jwtIssuer := config.GetEnv(
		"JWT_ISSUER",
		"praktikum-backend",
	)

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		jwtIssuer,
		15*time.Minute,
		7*24*time.Hour,
	)

	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
	)

	// ============================================================
	// APPLICATION
	// ============================================================

	app := config.NewApp(
		studentService,
		authService,
		jwtManager,
		permissionSet,
	)

	log.Println(
		"Server berjalan di http://localhost:3000",
	)

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
