FROM golang:1.21 as builder

WORKDIR /workspace

COPY go.* .
RUN go mod download \
    && go mod verify

COPY pkg pkg
COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go-git-backup ./cmd/git-backup


# Use distroless as minimal base image to package the project
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /go-git-backup .
ENTRYPOINT ["/go-git-backup"]
