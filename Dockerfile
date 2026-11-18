FROM node:22-alpine AS web-build
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install
COPY apps/web apps/web
RUN pnpm --filter @slopstack/design-web build

FROM golang:1.26-alpine AS server-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-build /src/apps/web/dist ./apps/server/public
RUN CGO_ENABLED=0 go build -o /out/slopstack-design ./apps/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=server-build /out/slopstack-design ./slopstack-design
ENV SLOPSTACK_DESIGN_ADDR=:8090
ENV SLOPSTACK_DESIGN_DATA=/data
VOLUME ["/data"]
EXPOSE 8090
ENTRYPOINT ["./slopstack-design"]
