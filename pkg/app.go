package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type App struct {
	*gin.Engine
}

func NewApp() *App {
	engine := gin.Default()
	return &App{Engine: engine}
}

// Run - starts the gin server
func (app *App) Run(addr string) error {
	err := app.Engine.Run(addr)
	if err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	return nil
}
