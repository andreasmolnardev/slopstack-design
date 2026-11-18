package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/slopstack/design/apps/server/internal/app"
	_ "github.com/slopstack/design/apps/server/pb_migrations"
)

func main() {
	dataDir := os.Getenv("SLOPSTACK_DESIGN_DATA")
	if dataDir == "" {
		dataDir = "./pb_data"
	}
	addr := os.Getenv("SLOPSTACK_DESIGN_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8090"
	}
	pb := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: dataDir})
	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		app.RegisterDesignMCPRoute(e)
		app.RegisterDesignReviewRoutes(e)
		static := filepath.Join("public")
		if _, err := os.Stat(static); err == nil {
			e.Router.GET("/{path...}", func(event *core.RequestEvent) error {
				return event.FileFS(os.DirFS(static), event.Request.PathValue("path"))
			})
		}
		return e.Next()
	})
	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
}
