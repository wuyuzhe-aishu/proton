package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/proton/ecms-go/internal/handler"
	"github.com/kweaver-ai/proton/ecms-go/internal/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9202"
	}

	r := gin.Default()

	g := r.Group("/api/ecms/v1alpha1")
	g.Use(middleware.SimpleAuth())

	execHandler := handler.NewExecHandler()
	execHandler.Register(g)

	fileHandler := handler.NewFileHandler()
	fileHandler.Register(g)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fmt.Printf("Starting server on :%s\n", port)
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
