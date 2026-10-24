package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	Mongo             *mongo.Client
	Permissions       *helper.PermissionSet
	UserService       *service.UserService
	CourseService     *service.CourseService
	SupervisorService *service.SupervisorService
	ReviewService     *service.ReviewService
	PortfolioService  *service.PortfolioService
	JWT               *helper.JWTManager
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	// --- users ---
	users := api.Group("/users")
	perms := deps.Permissions

	// Hak dapat diputuskan tanpa melihat data -> middleware
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List,
	)

	users.Post("/",
		middleware.RequirePermission(perms, "user:update:any"),
		deps.UserService.Create,
	)

	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete,
	)

	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole,
	)

	// Hak bergantung pada kepemilikan data -> diperiksa di service
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// --- kelas dan KRS ---
	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", deps.CourseService.List)
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

	if deps.ReviewService != nil {
		courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
		courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
		courses.Post("/:id/reviews", deps.ReviewService.Create)

		// --- ulasan kelas (MongoDB) ---
		reviews := api.Group("/reviews",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		reviews.Get("/:id", deps.ReviewService.Get)
		reviews.Patch("/:id", deps.ReviewService.Update)
		reviews.Delete("/:id", deps.ReviewService.Delete)
		reviews.Post("/:id/replies", deps.ReviewService.Reply)
		reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)
	}

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT))
	enrollments.Post("/swap", deps.CourseService.Swap)

	// --- pembimbing tugas akhir (D.2) ---
	if deps.SupervisorService != nil {
		supervisors := api.Group("/supervisors",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		supervisors.Get("/", deps.SupervisorService.List)
		supervisors.Get("/:id", deps.SupervisorService.Get)
		supervisors.Post("/:id/assign",
			middleware.RequirePermission(perms, "supervision:manage"),
			deps.SupervisorService.Assign)
		supervisors.Patch("/:id",
			middleware.RequirePermission(perms, "supervision:manage"),
			deps.SupervisorService.UpdateQuota)

		supervisions := api.Group("/supervisions",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		supervisions.Post("/switch",
			middleware.RequirePermission(perms, "supervision:manage"),
			deps.SupervisorService.Switch)
	}

	// --- portofolio mahasiswa (Tugas Mandiri Modul 10) ---
	if deps.PortfolioService != nil {
		students := api.Group("/students",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		students.Get("/:id/portfolio", deps.PortfolioService.Get)
		students.Put("/:id/portfolio", deps.PortfolioService.Upsert)
		students.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
		students.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)

		api.Get("/portfolios/stats/technologies", deps.PortfolioService.TopTechnologies)
	}
}

// healthCheck kini memeriksa DUA database. Aplikasi dianggap sehat hanya
// bila keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
			}
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
