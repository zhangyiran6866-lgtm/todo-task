# 构建阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s' \
    -o server ./cmd/server/main.go

# 运行阶段
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
ENV TZ=Asia/Shanghai
EXPOSE 8080
CMD ["./server"]
