FROM golang AS build-env

MAINTAINER shadero
COPY ./src /work/
ENV GO111MODULE=on \
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64
WORKDIR /work
RUN go mod download && \
	go build \
		-ldflags "-s -w" \
		-o /work/app \
		/work/main.go

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=build-env /work/app /root/app

ENTRYPOINT ["/root/app"]
