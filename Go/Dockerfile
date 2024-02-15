# syntax=docker/dockerfile:1

FROM golang:1.19

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . /app
COPY templates /app/templates

# Build
RUN go build -o /docker-gs-ping

EXPOSE 8080

ENV NAME World

# Run
CMD ["/docker-gs-ping"]
