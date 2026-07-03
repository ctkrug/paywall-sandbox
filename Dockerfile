FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/paywall-sandbox ./cmd/paywall-sandbox

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/paywall-sandbox /usr/local/bin/paywall-sandbox
EXPOSE 8402
ENTRYPOINT ["/usr/local/bin/paywall-sandbox", "serve", "--addr", ":8402"]
