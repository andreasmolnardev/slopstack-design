package migrations

import "github.com/pocketbase/pocketbase/core"

func init() {
	core.AppMigrations.Register(func(app core.App) error {
		queries := []string{
			`CREATE TABLE IF NOT EXISTS design_projects (id TEXT PRIMARY KEY, owner TEXT NOT NULL, name TEXT NOT NULL, repository_url TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created TEXT NOT NULL, updated TEXT NOT NULL)`,
			`CREATE INDEX IF NOT EXISTS idx_design_projects_owner ON design_projects(owner)`,
			`CREATE TABLE IF NOT EXISTS design_sessions (id TEXT PRIMARY KEY, project TEXT NOT NULL REFERENCES design_projects(id) ON DELETE CASCADE, owner TEXT NOT NULL, title TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'draft', context TEXT NOT NULL DEFAULT '{}', created TEXT NOT NULL, updated TEXT NOT NULL)`,
			`CREATE INDEX IF NOT EXISTS idx_design_sessions_project ON design_sessions(project, updated)`,
			`CREATE TABLE IF NOT EXISTS design_proposals (id TEXT PRIMARY KEY, session TEXT NOT NULL REFERENCES design_sessions(id) ON DELETE CASCADE, owner TEXT NOT NULL, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', preview_url TEXT NOT NULL DEFAULT '', screenshot TEXT NOT NULL DEFAULT '', revision INTEGER NOT NULL DEFAULT 1, revision_of TEXT NOT NULL DEFAULT '', created TEXT NOT NULL, updated TEXT NOT NULL)`,
			`CREATE INDEX IF NOT EXISTS idx_design_proposals_session ON design_proposals(session, revision)`,
			`CREATE TABLE IF NOT EXISTS design_reviews (id TEXT PRIMARY KEY, proposal TEXT NOT NULL REFERENCES design_proposals(id) ON DELETE CASCADE, reviewer TEXT NOT NULL, decision TEXT NOT NULL CHECK(decision IN ('feedback','approved','rejected')), feedback TEXT NOT NULL DEFAULT '', created TEXT NOT NULL)`,
			`CREATE INDEX IF NOT EXISTS idx_design_reviews_proposal ON design_reviews(proposal, created)`,
		}
		for _, query := range queries {
			if _, err := app.DB().NewQuery(query).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, table := range []string{"design_reviews", "design_proposals", "design_sessions", "design_projects"} {
			if _, err := app.DB().NewQuery("DROP TABLE IF EXISTS " + table).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, "20261118000000_create_design_workspace.go")
}
