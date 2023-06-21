FROM golang:1.19.9-alpine3.18  AS build-env
RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/central-api
ADD . /go/src/github.com/devtron-labs/central-api
RUN GOOS=linux make

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build-env  /go/src/github.com/devtron-labs/central-api/central-api .
COPY ./DockerfileTemplateData.json /DockerfileTemplateData.json
COPY ./BuildpackMetadata.json /BuildpackMetadata.json
CMD ["./central-api"]