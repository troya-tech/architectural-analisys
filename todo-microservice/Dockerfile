# Multi-stage build for a small container image
FROM golang:1.21 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/todo-service .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/todo-service /app/todo-service
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/todo-service"]
