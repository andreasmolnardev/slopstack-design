# Slopstack Design

A self-hosted review workspace for coding agents and the people who approve their UI work. Application repositories stay the source of truth: this service stores design context, proposals, previews, decisions, and revision history, but never edits a target repository.

## Stack

- React, Vite, TypeScript
- Go and PocketBase with persistent SQLite storage
- MCP over Streamable HTTP for agent workflows

## Development

Requirements: Go 1.26+, Node.js 22+, and pnpm 10+.

```sh
pnpm install
pnpm dev
```

The server listens on `:8090` by default; the web client runs on Vite's local port. PocketBase admin UI is available at `/_/`. Set `SLOPSTACK_DESIGN_DATA` to choose the persistent data directory and `SLOPSTACK_DESIGN_ADDR` to change the listen address.

Create user accounts from the PocketBase admin UI. The web app signs in with a PocketBase superuser account; its token is kept in browser local storage. Projects and all associated records are scoped to that account.

## Self-host

Build and run with Docker Compose:

```sh
docker compose up --build
```

Persistent application data is stored in the `design-data` volume. Back it up while the service is stopped. Configure a reverse proxy with HTTPS before allowing remote access. The first account is created through PocketBase's admin setup flow.

## Agent workflow

Connect an MCP client to `/mcp` using a PocketBase superuser bearer token. The endpoint accepts stateless JSON-RPC over HTTP. Implemented methods are `initialize`, `ping`, `tools/list`, and `tools/call`; tools include `list_projects`, `register_project`, `create_design_session`, `submit_proposal`, and `read_reviews`. Human review actions are recorded by the authenticated web API and appear in `read_reviews`. Upload screenshots as multipart field `file` to `/api/design/sessions/{sessionId}/previews`, then submit its returned URL as `screenshot` on `submit_proposal`. External preview URLs can be submitted as `preview_url`.

The web UI can be opened on a review session with `/?session={sessionId}`. It lists proposals, records approval or requested changes, and exposes session creation. An approval records intent only; an agent must make and report repository changes separately.

## Review principles

See [the style guide](docs/STYLE_GUIDE.md) for reusable visual conventions and the anti-slop review checklist.
