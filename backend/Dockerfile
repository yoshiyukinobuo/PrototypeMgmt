FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/server .

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/server /server
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
