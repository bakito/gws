FROM ubuntu AS gcloud

RUN apt-get update && \
    apt-get install apt-transport-https ca-certificates gnupg curl -y && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    apt-get update && \
    apt-get install google-cloud-cli -y

FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY main.go /build
COPY auth_config.go.tpl /build
COPY --from=gcloud /usr/lib/google-cloud-sdk/lib/googlecloudsdk/core/config.py /build

RUN go run main.go > auth_config.go

FROM scratch
COPY --from=builder /build/auth_config.go /