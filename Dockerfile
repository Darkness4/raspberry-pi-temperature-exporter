FROM docker.io/arm64v8/golang:1.17.7 as builder
WORKDIR /work
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o app .

# ---
FROM docker.io/arm64v8/busybox:1.35.0

ENV HOST=0.0.0.0
ENV PORT=3000

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-static-arm64 /tini
RUN chmod +x /tini

RUN mkdir /app
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app

COPY --from=builder /work/app .

RUN chown -R app:app .
USER app

EXPOSE 3000

ENTRYPOINT [ "/tini", "--", "./app" ]
