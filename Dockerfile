FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/activity-consumer ./cmd/activity-consumer

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/activity-consumer /app/activity-consumer
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/api"]
