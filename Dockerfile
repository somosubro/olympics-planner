# Build: docker build -t olympics-schedule-planner-api .
# Run:  docker run --rm -p 8080:8080 -e PORT=8080 olympics-schedule-planner-api
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api .
COPY data/ ./data/
ENV PORT=8080
EXPOSE 8080
USER nobody
CMD ["./api"]
