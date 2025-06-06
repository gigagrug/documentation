FROM golang:alpine as dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY . .
RUN go mod download
CMD ["air", "-c", ".air.toml"]

FROM --platform=$BUILDPLATFORM golang:alpine AS build
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
	--mount=type=bind,source=go.mod,target=go.mod \
	go mod download -x
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod/ \
	--mount=type=bind,target=. \
	CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server .

FROM alpine:latest AS final
RUN --mount=type=cache,target=/var/cache/apk \
	apk --update add \
	ca-certificates \
	tzdata \
	&& \
	update-ca-certificates
ARG UID=10001
RUN adduser \
	--disabled-password \
	--gecos "" \
	--home "/nonexistent" \
	--shell "/sbin/nologin" \
	--no-create-home \
	--uid "${UID}" \
	appuser
USER appuser
COPY --from=build /bin/server /bin/
COPY ./frontend/ ./frontend/
EXPOSE 8080
ENTRYPOINT [ "/bin/server" ]
