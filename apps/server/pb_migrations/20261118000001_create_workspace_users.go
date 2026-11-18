package migrations

import "github.com/pocketbase/pocketbase/core"

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		if _, err := app.FindCollectionByNameOrId("users"); err == nil {
			return nil
		}
		return app.Save(core.NewAuthCollection("users"))
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	}, "20261118000001_create_workspace_users.go")
}
