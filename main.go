package main

import (
	"log"
	"os"

	"users-api/config"
	"users-api/handlers"
	"users-api/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	client := config.NewDynamoDBClient()
	repo := repository.NewUserRepository(client)
	handler := handlers.NewUserHandler(repo)

	r := gin.Default()

	users := r.Group("/users")
	{
		users.POST("", handler.CreateUser)
		users.GET("", handler.GetUsers)
		users.GET("/:id", handler.GetUserByID)
		users.PUT("/:id", handler.UpdateUser)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
