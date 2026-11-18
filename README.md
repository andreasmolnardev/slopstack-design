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

## Self-host

Build and run with Docker Compose:

```sh
docker compose up --build
```

Persistent application data is stored in the `design-data` volume. Back it up while the service is stopped. Configure a reverse proxy with HTTPS before allowing remote access. The first account is created through PocketBase's admin setup flow.

## Agent workflow

Connect an MCP client to `/mcp`. Register a project, create a session, attach design context, submit one or more proposals with screenshots or preview URLs, then request review. The MCP response includes structured decisions and revision identifiers. An approval records intent only; an agent must make and report repository changes separately.

## Review principles

See [the style guide](docs/STYLE_GUIDE.md) for reusable visual conventions and the anti-slop review checklist.
