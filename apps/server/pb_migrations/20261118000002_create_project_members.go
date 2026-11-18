package migrations

import "github.com/pocketbase/pocketbase/core"

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		_, err := app.DB().NewQuery(`CREATE TABLE IF NOT EXISTS design_project_members (id TEXT PRIMARY KEY, project TEXT NOT NULL REFERENCES design_projects(id) ON DELETE CASCADE, user TEXT NOT NULL, role TEXT NOT NULL CHECK(role IN ('editor','reviewer')), created TEXT NOT NULL, UNIQUE(project,user))`).Execute()
		if err != nil {
			return err
		}
		_, err = app.DB().NewQuery(`CREATE INDEX IF NOT EXISTS idx_design_project_members_user ON design_project_members(user,project)`).Execute()
		return err
	}, func(app core.App) error {
		_, err := app.DB().NewQuery(`DROP TABLE IF EXISTS design_project_members`).Execute()
		return err
	}, "20261118000002_create_project_members.go")
}
