package main

import (
	"fmt"

	"certhub-backend/internal/config"
	"certhub-backend/internal/database"
	"certhub-backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()
	database.Init()

	if config.C.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := routes.SetupRouter()
	addr := fmt.Sprintf(":%s", config.C.Server.Port)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
