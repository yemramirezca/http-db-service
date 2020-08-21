FROM golang:1.12 as builder
ENV GO111MODULE=on
ARG DOCK_PKG_DIR=/go/src/github.com/yemramirezca/http-db-service/
ADD . $DOCK_PKG_DIR
WORKDIR $DOCK_PKG_DIR
RUN CGO_ENABLED=0 GOOS=linux go build  -o main .
#RUN go test ./...


FROM alpine:3.12.0
WORKDIR /app/
COPY --from=builder /go/src/github.com/yemramirezca/http-db-service/main /app/
COPY --from=builder /go/src/github.com/yemramirezca/http-db-service/docs/api/api.yaml /app/
CMD ["./main"]

EXPOSE 8017:8017