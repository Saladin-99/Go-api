package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"Go-api/pkg/controllers"
	database "Go-api/pkg/database/mongodb"
	"Go-api/pkg/database/mongodb/repository"
	"Go-api/pkg/utils"
)

func main() {
	// Initialize the logger
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	// Initialize the database
	db, err := database.NewDB(logger, "config/database-config.yaml")
	if err != nil {
		logger.Fatalf("Error initializing database: %v", err)
	}

	// Initialize repositories
	userRepository := repository.NewUserRepository(db.DB)
	orgRepository := repository.NewOrganizationRepository(db.DB)

	// Initialize controllers
	userController := controllers.NewUserController(logger, userRepository)
	orgController := controllers.NewOrganizationController(logger, orgRepository, userRepository)
	// Set up HTTP server
	router := gin.Default()

	// Create a router group for user-related routes
	userRoutes := router.Group("/user")
	{
		userRoutes.POST("/signup", userController.SignUp)
		userRoutes.POST("/signin", userController.SignIn)
		userRoutes.POST("/refresh", userController.RefreshToken)
	}
	// Apply JWT authentication middleware to all routes in the "/organization" group
	orgRoutes := router.Group("/organization")
	orgRoutes.Use(utils.AuthMiddleware()) // Apply the middleware here

	// Define your routes
	orgRoutes.POST("/", orgController.CreateOrg)
	orgRoutes.GET("/:organization_id", orgController.GetOrgByID)
	orgRoutes.PUT("/:organization_id", orgController.UpdateOrg)
	orgRoutes.DELETE("/:organization_id", orgController.DeleteOrg)
	orgRoutes.POST("/:organization_id/invite", orgController.InviteUser)

	// Start the server
	logger.Println("Starting server on :8080")
	err = router.Run(":8080")
	if err != nil {
		logger.Fatalf("Error starting server: %v", err)
	}
}
