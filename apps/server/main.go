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
	if flag := pb.RootCmd.PersistentFlags().Lookup("http"); flag != nil {
		_ = pb.RootCmd.PersistentFlags().Set("http", addr)
	}
	pb.OnServe().BindFunc(func(e *core.ServeEvent) error {
		app.RegisterDesignMCPRoute(e)
		app.RegisterDesignReviewRoutes(e)
		static := filepath.Join("public")
		if _, err := os.Stat(static); err == nil {
			e.Router.GET("/{path...}", func(event *core.RequestEvent) error {
				name := event.Request.PathValue("path")
				if name == "" {
					name = "index.html"
				}
				if _, err := os.Stat(filepath.Join(static, filepath.Clean(name))); err != nil {
					name = "index.html"
				}
				return event.FileFS(os.DirFS(static), name)
			})
		}
		return e.Next()
	})
	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
}
