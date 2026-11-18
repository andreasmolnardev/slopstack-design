package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func designID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}

// RegisterDesignMCPRoute installs an authenticated, stateless MCP endpoint.
// Call it from the application's OnServe hook. PocketBase's normal auth
// middleware validates the bearer token before this handler runs.
func RegisterDesignMCPRoute(e *core.ServeEvent) {
	e.Router.POST("/mcp", func(event *core.RequestEvent) error {
		if event.Auth == nil || event.Auth.Collection().Name != "users" {
			return event.UnauthorizedError("valid PocketBase user token required", nil)
		}
		var request struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Method  string          `json:"method"`
			Params  struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(io.LimitReader(event.Request.Body, 1<<20)).Decode(&request); err != nil {
			return event.JSON(http.StatusBadRequest, map[string]any{"jsonrpc": "2.0", "id": nil, "error": map[string]any{"code": -32700, "message": "invalid JSON"}})
		}
		response := map[string]any{"jsonrpc": "2.0", "id": request.ID}
		switch request.Method {
		case "initialize":
			response["result"] = map[string]any{"protocolVersion": "2025-03-26", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "slopstack-design", "version": "0.1.0"}}
		case "ping":
			response["result"] = map[string]any{}
		case "tools/list":
			response["result"] = map[string]any{"tools": designMCPTools()}
		case "tools/call":
			value, err := callDesignMCPTool(event.App, event.Auth.Id, request.Params.Name, request.Params.Arguments)
			if err != nil {
				response["result"] = map[string]any{"isError": true, "content": []any{map[string]string{"type": "text", "text": err.Error()}}}
			} else {
				data, _ := json.Marshal(value)
				response["result"] = map[string]any{"content": []any{map[string]string{"type": "text", "text": string(data)}}}
			}
		default:
			response["error"] = map[string]any{"code": -32601, "message": "method not found"}
		}
		return event.JSON(http.StatusOK, response)
	}).Bind(apis.RequireAuth("users"))
}

// RegisterDesignReviewRoutes adds the human review actions used by the web UI.
// All routes require a PocketBase user session and proposal ownership is
// checked through the project/session chain before a decision is recorded.
func RegisterDesignReviewRoutes(e *core.ServeEvent) {
	e.Router.GET("/api/design/projects", func(event *core.RequestEvent) error {
		var rows []map[string]any
		err := event.App.DB().NewQuery("SELECT id,name,repository_url,metadata,created,updated FROM design_projects WHERE owner={:owner} ORDER BY name").Bind(map[string]any{"owner": event.Auth.Id}).All(&rows)
		if err != nil {
			return event.InternalServerError("could not load projects", err)
		}
		return event.JSON(http.StatusOK, rows)
	}).Bind(apis.RequireAuth("users"))
	e.Router.POST("/api/design/projects", func(event *core.RequestEvent) error {
		var input struct {
			Name          string         `json:"name"`
			RepositoryURL string         `json:"repository_url"`
			Metadata      map[string]any `json:"metadata"`
		}
		if err := json.NewDecoder(io.LimitReader(event.Request.Body, 256<<10)).Decode(&input); err != nil {
			return event.BadRequestError("invalid project", err)
		}
		input.Name = strings.TrimSpace(input.Name)
		input.RepositoryURL = strings.TrimSpace(input.RepositoryURL)
		if input.Name == "" {
			return event.BadRequestError("project name is required", nil)
		}
		metadata, _ := json.Marshal(input.Metadata)
		if len(metadata) == 0 {
			metadata = []byte("{}")
		}
		id, now := designID(), time.Now().UTC().Format(time.RFC3339Nano)
		_, err := event.App.DB().NewQuery("INSERT INTO design_projects (id,owner,name,repository_url,metadata,created,updated) VALUES ({:id},{:owner},{:name},{:url},{:metadata},{:now},{:now})").Bind(map[string]any{"id": id, "owner": event.Auth.Id, "name": input.Name, "url": input.RepositoryURL, "metadata": string(metadata), "now": now}).Execute()
		if err != nil {
			return event.InternalServerError("could not create project", err)
		}
		return event.JSON(http.StatusCreated, map[string]string{"id": id})
	}).Bind(apis.RequireAuth("users"))
	e.Router.GET("/api/design/projects/{id}/sessions", func(event *core.RequestEvent) error {
		var rows []map[string]any
		err := event.App.DB().NewQuery("SELECT id,project,title,status,context,created,updated FROM design_sessions WHERE project={:project} AND owner={:owner} ORDER BY updated DESC").Bind(map[string]any{"project": event.Request.PathValue("id"), "owner": event.Auth.Id}).All(&rows)
		if err != nil {
			return event.InternalServerError("could not load sessions", err)
		}
		return event.JSON(http.StatusOK, rows)
	}).Bind(apis.RequireAuth("users"))
	e.Router.POST("/api/design/projects/{id}/sessions", func(event *core.RequestEvent) error {
		var input struct {
			Title   string         `json:"title"`
			Context map[string]any `json:"context"`
		}
		if err := json.NewDecoder(io.LimitReader(event.Request.Body, 256<<10)).Decode(&input); err != nil {
			return event.BadRequestError("invalid session", err)
		}
		input.Title = strings.TrimSpace(input.Title)
		if input.Title == "" {
			return event.BadRequestError("session title is required", nil)
		}
		context, _ := json.Marshal(input.Context)
		if len(context) == 0 {
			context = []byte("{}")
		}
		id, now := designID(), time.Now().UTC().Format(time.RFC3339Nano)
		_, err := event.App.DB().NewQuery("INSERT INTO design_sessions(id,project,owner,title,status,context,created,updated) SELECT {:id},id,{:owner},{:title},'draft',{:context},{:now},{:now} FROM design_projects WHERE id={:project} AND owner={:owner}").Bind(map[string]any{"id": id, "project": event.Request.PathValue("id"), "owner": event.Auth.Id, "title": input.Title, "context": string(context), "now": now}).Execute()
		if err != nil {
			return event.InternalServerError("could not create session", err)
		}
		var count int
		_ = event.App.DB().NewQuery("SELECT count(*) FROM design_sessions WHERE id={:id}").Bind(map[string]any{"id": id}).Row(&count)
		if count == 0 {
			return event.NotFoundError("project not found", nil)
		}
		return event.JSON(http.StatusCreated, map[string]string{"id": id})
	}).Bind(apis.RequireAuth("users"))
	e.Router.GET("/api/design/proposals/{id}/reviews", func(event *core.RequestEvent) error {
		var rows []map[string]any
		err := event.App.DB().NewQuery("SELECT r.id,r.decision,r.feedback,r.created FROM design_reviews r JOIN design_proposals p ON p.id=r.proposal WHERE p.id={:proposal} AND p.owner={:owner} ORDER BY r.created").Bind(map[string]any{"proposal": event.Request.PathValue("id"), "owner": event.Auth.Id}).All(&rows)
		if err != nil {
			return event.InternalServerError("could not load history", err)
		}
		return event.JSON(http.StatusOK, rows)
	}).Bind(apis.RequireAuth("users"))
	e.Router.POST("/api/design/sessions/{id}/previews", func(event *core.RequestEvent) error {
		event.Request.Body = http.MaxBytesReader(event.Response, event.Request.Body, 16<<20)
		if err := event.Request.ParseMultipartForm(16 << 20); err != nil {
			return event.BadRequestError("upload must be a screenshot under 16 MB", err)
		}
		file, header, err := event.Request.FormFile("file")
		if err != nil {
			return event.BadRequestError("a screenshot file is required", err)
		}
		defer file.Close()
		mediaType := header.Header.Get("Content-Type")
		if mediaType != "image/png" && mediaType != "image/jpeg" && mediaType != "image/webp" {
			return event.BadRequestError("screenshot must be PNG, JPEG, or WebP", nil)
		}
		var sessionID string
		err = event.App.DB().NewQuery("SELECT id FROM design_sessions WHERE id={:id} AND owner={:owner}").Bind(map[string]any{"id": event.Request.PathValue("id"), "owner": event.Auth.Id}).Row(&sessionID)
		if err != nil || sessionID == "" {
			return event.NotFoundError("session not found", err)
		}
		id := designID()
		ext, _ := mime.ExtensionsByType(mediaType)
		suffix := ".img"
		if len(ext) > 0 {
			suffix = ext[0]
		}
		name := id + suffix
		// Store uploads under the PocketBase data directory and expose only an opaque file URL.
		path := filepath.Join(event.App.DataDir(), "design_uploads", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return event.InternalServerError("could not store screenshot", err)
		}
		out, err := os.Create(path)
		if err != nil {
			return event.InternalServerError("could not store screenshot", err)
		}
		defer out.Close()
		if _, err = io.Copy(out, io.LimitReader(file, 16<<20)); err != nil {
			return event.InternalServerError("could not store screenshot", err)
		}
		return event.JSON(http.StatusCreated, map[string]string{"url": "/api/design/previews/" + name})
	}).Bind(apis.RequireAuth("users"))
	e.Router.GET("/api/design/previews/{name}", func(event *core.RequestEvent) error {
		name := event.Request.PathValue("name")
		if len(name) < 20 || strings.ContainsAny(name, "/\\") {
			return event.NotFoundError("preview not found", nil)
		}
		var count int
		err := event.App.DB().NewQuery("SELECT count(*) FROM design_proposals p JOIN design_sessions s ON s.id=p.session WHERE s.owner={:owner} AND (p.preview_url={:url} OR p.screenshot={:name})").Bind(map[string]any{"owner": event.Auth.Id, "url": "/api/design/previews/" + name, "name": name}).Row(&count)
		if err != nil || count == 0 {
			return event.NotFoundError("preview not found", err)
		}
		return event.FileFS(os.DirFS(filepath.Join(event.App.DataDir(), "design_uploads")), name)
	}).Bind(apis.RequireAuth("users"))
	e.Router.GET("/api/design/sessions/{id}/proposals", func(event *core.RequestEvent) error {
		if event.Auth == nil || event.Auth.Collection().Name != "users" {
			return event.UnauthorizedError("sign in required", nil)
		}
		var rows []map[string]any
		err := event.App.DB().NewQuery("SELECT p.id,p.title,p.description,p.preview_url,p.revision,p.created,p.updated FROM design_proposals p JOIN design_sessions s ON s.id=p.session WHERE p.session={:session} AND s.owner={:owner} ORDER BY p.created").Bind(map[string]any{"session": event.Request.PathValue("id"), "owner": event.Auth.Id}).All(&rows)
		if err != nil {
			return event.InternalServerError("could not load proposals", err)
		}
		return event.JSON(http.StatusOK, rows)
	}).Bind(apis.RequireAuth("users"))
	e.Router.POST("/api/design/proposals/{id}/reviews", func(event *core.RequestEvent) error {
		if event.Auth == nil || event.Auth.Collection().Name != "users" {
			return event.UnauthorizedError("sign in required", nil)
		}
		var input struct {
			Decision string `json:"decision"`
			Feedback string `json:"feedback"`
		}
		if err := json.NewDecoder(io.LimitReader(event.Request.Body, 64<<10)).Decode(&input); err != nil {
			return event.BadRequestError("invalid review", err)
		}
		if input.Decision != "feedback" && input.Decision != "approved" && input.Decision != "rejected" {
			return event.BadRequestError("decision must be feedback, approved, or rejected", nil)
		}
		var id string
		err := event.App.DB().NewQuery("SELECT p.id FROM design_proposals p JOIN design_sessions s ON s.id=p.session WHERE p.id={:proposal} AND s.owner={:owner}").Bind(map[string]any{"proposal": event.Request.PathValue("id"), "owner": event.Auth.Id}).Row(&id)
		if err != nil || id == "" {
			return event.NotFoundError("proposal not found", err)
		}
		_, err = event.App.DB().NewQuery("INSERT INTO design_reviews (id,proposal,reviewer,decision,feedback,created) VALUES ({:id},{:proposal},{:reviewer},{:decision},{:feedback},{:created})").Bind(map[string]any{"id": fmt.Sprintf("%d", time.Now().UnixNano()), "proposal": id, "reviewer": event.Auth.Id, "decision": input.Decision, "feedback": input.Feedback, "created": time.Now().UTC().Format("2006-01-02 15:04:05.000Z")}).Execute()
		if err != nil {
			return event.InternalServerError("could not save review", err)
		}
		return event.JSON(http.StatusCreated, map[string]string{"status": input.Decision})
	}).Bind(apis.RequireAuth("users"))
}

func designMCPTools() []map[string]any {
	str := map[string]any{"type": "string"}
	return []map[string]any{
		{"name": "list_projects", "description": "List projects registered by the authenticated user", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}}},
		{"name": "register_project", "description": "Register a code repository for design work", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"name": str, "repository_url": str, "metadata": map[string]any{"type": "object"}}, "required": []string{"name", "repository_url"}}},
		{"name": "create_design_session", "description": "Start a design review session for a registered project", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"project_id": str, "title": str}, "required": []string{"project_id", "title"}}},
		{"name": "submit_proposal", "description": "Submit a design direction or a new revision for review", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"session_id": str, "title": str, "description": str, "preview_url": str, "screenshot": str, "revision_of": str}, "required": []string{"session_id", "title"}}},
		{"name": "read_reviews", "description": "Read feedback and decisions for a proposal", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"proposal_id": str}, "required": []string{"proposal_id"}}},
	}
}

func callDesignMCPTool(app core.App, owner, name string, args map[string]any) (any, error) {
	arg := func(key string) string { value, _ := args[key].(string); return strings.TrimSpace(value) }
	id := func() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
	now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")
	switch name {
	case "list_projects":
		var rows []map[string]any
		err := app.DB().NewQuery("SELECT id,name,repository_url,metadata,created,updated FROM design_projects WHERE owner={:owner} ORDER BY name").Bind(map[string]any{"owner": owner}).All(&rows)
		return rows, err
	case "register_project":
		projectID := id()
		metadata, _ := json.Marshal(args["metadata"])
		if len(metadata) == 0 {
			metadata = []byte("{}")
		}
		_, err := app.DB().NewQuery("INSERT INTO design_projects (id,owner,name,repository_url,metadata,created,updated) VALUES ({:id},{:owner},{:name},{:url},{:metadata},{:now},{:now})").Bind(map[string]any{"id": projectID, "owner": owner, "name": arg("name"), "url": arg("repository_url"), "metadata": string(metadata), "now": now}).Execute()
		return map[string]string{"id": projectID}, err
	case "create_design_session":
		if arg("title") == "" {
			return nil, fmt.Errorf("title is required")
		}
		sessionID := id()
		_, err := app.DB().NewQuery("INSERT INTO design_sessions (id,project,owner,title,status,created,updated) SELECT {:id},id,{:owner},{:title},'draft',{:now},{:now} FROM design_projects WHERE id={:project} AND owner={:owner}").Bind(map[string]any{"id": sessionID, "owner": owner, "title": arg("title"), "now": now, "project": arg("project_id")}).Execute()
		if err != nil {
			return nil, err
		}
		var count int
		_ = app.DB().NewQuery("SELECT count(*) FROM design_sessions WHERE id={:id}").Bind(map[string]any{"id": sessionID}).Row(&count)
		if count == 0 {
			return nil, fmt.Errorf("project not found or access denied")
		}
		return map[string]string{"id": sessionID}, nil
	case "submit_proposal":
		proposalID := id()
		if arg("title") == "" {
			return nil, fmt.Errorf("title is required")
		}
		revision, revisionOf := 1, arg("revision_of")
		if revisionOf != "" {
			_ = app.DB().NewQuery("SELECT revision+1 FROM design_proposals WHERE id={:id} AND owner={:owner}").Bind(map[string]any{"id": revisionOf, "owner": owner}).Row(&revision)
			if revision < 2 {
				return nil, fmt.Errorf("revision target not found or access denied")
			}
		}
		_, err := app.DB().NewQuery("INSERT INTO design_proposals (id,session,owner,title,description,preview_url,screenshot,revision,revision_of,created,updated) SELECT {:id},s.id,{:owner},{:title},{:description},{:preview},{:screenshot},{:revision},{:revision_of},{:now},{:now} FROM design_sessions s WHERE s.id={:session} AND s.owner={:owner}").Bind(map[string]any{"id": proposalID, "owner": owner, "title": arg("title"), "description": arg("description"), "preview": arg("preview_url"), "screenshot": arg("screenshot"), "revision": revision, "revision_of": revisionOf, "now": now, "session": arg("session_id")}).Execute()
		if err != nil {
			return nil, err
		}
		var count int
		_ = app.DB().NewQuery("SELECT count(*) FROM design_proposals WHERE id={:id}").Bind(map[string]any{"id": proposalID}).Row(&count)
		if count == 0 {
			return nil, fmt.Errorf("session not found or access denied")
		}
		return map[string]string{"id": proposalID}, nil
	case "read_reviews":
		var rows []map[string]any
		err := app.DB().NewQuery("SELECT r.id,r.decision,r.feedback,r.created FROM design_reviews r JOIN design_proposals p ON p.id=r.proposal WHERE p.id={:proposal} AND p.owner={:owner} ORDER BY r.created").Bind(map[string]any{"proposal": arg("proposal_id"), "owner": owner}).All(&rows)
		return rows, err
	default:
		return nil, fmt.Errorf("unknown tool %q", name)
	}
}
