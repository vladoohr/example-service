### Multi-stage build
FROM golang:alpine as build

RUN go get -u -v github.com/JormungandrK/microservice-tools/gateway

COPY . /go/src/github.com/natemago/example-service
RUN go install github.com/natemago/example-service

### Main
FROM alpine:3.7

COPY --from=build /go/bin/example-service /usr/local/bin/example-service
EXPOSE 8080

ENV SERVICE_NAME="example-service"
ENV SERVICE_DOMAIN="service.consul"
ENV APIGW_ADMIN_URL="http://kong:8000"

CMD ["/usr/local/bin/example-service", "-gw", "${APIGW_ADMIN_URL}", "-name", "${SERVICE_NAME}", "-p","8080", "-domain", "${SERVICE_DOMAIN}"]