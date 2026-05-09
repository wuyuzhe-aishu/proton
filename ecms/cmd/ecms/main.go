package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/proton/ecms/internal/handler"
	"github.com/kweaver-ai/proton/ecms/internal/middleware"
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

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
