FROM golang:1.27.2-alpine AS base

RUN apk add --no-cache make

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

FROM base AS build
RUN make

FROM base AS ci
RUN apk add --no-cache gcc musl-dev

FROM gcr.io/distroless/static-debian13:nonroot AS runtime

COPY --from=build /app/bin/ovh-ddns /ovh-ddns
WORKDIR /data
VOLUME ["/data"]

ENTRYPOINT ["/ovh-ddns"]
