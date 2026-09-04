FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
COPY --from=build /server /usr/local/bin/server
EXPOSE 8080
CMD ["server"]
