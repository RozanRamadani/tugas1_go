package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	"latihan-fiber/route"
)

func main() {
	// ============================================================
	// 1. LOAD CONFIGURATION & DATABASE
	// ============================================================
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		log.Fatalf("gagal terhubung ke MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()
	_ = mongoClient.Database(config.GetEnv("MONGO_DATABASE", "praktikum_backend"))

	// ============================================================
	// 2. RAKIT DEPENDENCY & LOAD PERMISSION
	// ============================================================
	userRepository := repository.NewUserRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	courseRepository := repository.NewCourseRepository(pool)
	txManager := repository.NewTxManager(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		log.Fatalf("gagal memuat permission: %v", err)
	}

	permissions := helper.NewPermissionSet(rawPermissions)
	log.Printf("permission dimuat, roles=%v", permissions.KnownRoles())

	jwtSecret := config.GetEnv("JWT_SECRET", "supersecretloadtestkey123456789012")
	jwtManager := helper.NewJWTManager(jwtSecret, "latihan-fiber", 15*time.Minute, 7*24*time.Hour)

	userService := service.NewUserService(userRepository, permissions)
	courseService := service.NewCourseService(courseRepository, txManager)
	supervisorRepository := repository.NewSupervisorRepository(pool)
	supervisorService := service.NewSupervisorService(supervisorRepository, txManager)

	// ============================================================
	// 3. BUAT APLIKASI FIBER
	// ============================================================
	app := config.NewApp(route.Dependencies{
		Pool:              pool,
		Mongo:             mongoClient,
		Permissions:       permissions,
		UserService:       userService,
		CourseService:     courseService,
		SupervisorService: supervisorService,
		JWT:               jwtManager,
	})

	// ============================================================
	// 4. JALANKAN SERVER
	// ============================================================
	port := config.GetEnv("APP_PORT", "3000")
	fmt.Println("Server berjalan di http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}
