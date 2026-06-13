# Flowork Router — standalone, OpenAI/Anthropic-compatible LLM gateway.
# Point ANY agent (Hermes, OpenClaw, Cursor, Claude Code, …) at it: base_url = http://host:2402/v1.
#   docker build -f router/Dockerfile -t flowork-router router
#   docker run -p 2402:2402 -v flowork-router-data:/data flowork-router
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY . .
RUN GOWORK=off CGO_ENABLED=0 go build -tags netgo -ldflags '-s -w' -o /flowork-router .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 flowork
COPY --from=build /flowork-router /usr/local/bin/flowork-router
ENV HOME=/data XDG_DATA_HOME=/data
RUN mkdir -p /data && chown flowork /data
USER flowork
VOLUME /data
EXPOSE 2402
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://127.0.0.1:2402/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/flowork-router", "-addr", "0.0.0.0:2402"]
