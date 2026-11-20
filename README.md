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

Create user accounts from the PocketBase admin UI. The web app signs in with a PocketBase user account; its token is kept in browser local storage. Project owners can add existing accounts as reviewers or editors. Reviewers can read proposals and submit decisions; editors can also create sessions and submit proposals. Records are scoped to the owner or project membership.

## Self-host

For a local development build from the checked-out source:

```sh
docker compose -f docker-compose.dev.yaml up --build
```

For a production deployment using the published GHCR image:

```sh
docker compose -f docker-compose.yaml pull
docker compose -f docker-compose.yaml up -d
```

Persistent application data is stored in the `design-data` volume (or `design-data-dev` for the dev compose file). Set `SLOPSTACK_DESIGN_TAG` to select a release; the default is `latest`. For a private GHCR package, authenticate Docker with GitHub before pulling. Back up the volume while the service is stopped. Configure a reverse proxy with HTTPS before allowing remote access. The first account is created through PocketBase's admin setup flow.

The [container workflow](.github/workflows/publish-container.yml) builds pull requests without publishing. Pushes to `main` publish `latest`, branch, and commit tags; version tags such as `v1.2.3` publish versioned images. The workflow builds `linux/amd64` and `linux/arm64` images.

## Agent workflow

Connect an MCP client to `/mcp` using a PocketBase user bearer token. The endpoint accepts stateless JSON-RPC over HTTP. Implemented methods are `initialize`, `ping`, `tools/list`, and `tools/call`; tools include `list_projects`, `register_project`, `create_design_session`, `read_design_context`, `submit_proposal`, and `read_reviews`. Human review actions are recorded by the authenticated web API and appear in `read_reviews`. Upload screenshots as multipart field `file` to `/api/design/sessions/{sessionId}/previews`, then submit its returned URL as `screenshot` on `submit_proposal`. External preview URLs can be submitted as `preview_url`.

The web UI can be opened on a review session with `/?session={sessionId}`. It lists proposals, records approval or requested changes, and exposes session creation. An approval records intent only; an agent must make and report repository changes separately.

## Review principles

See [the style guide](docs/STYLE_GUIDE.md) for reusable visual conventions and the anti-slop review checklist.

The first integration flow is documented in the [Taskboard pilot guide](docs/TASKBOARD_PILOT.md), with its current token snapshot in `docs/templates/taskboard.json`.
