# Build stage
FROM debian:bullseye-slim AS builder

# Build argument to specify which service to build
ARG BUILD_TARGET=gateway

# Configure Aliyun Debian mirrors for build stage
RUN echo "deb http://mirrors.aliyun.com/debian/ bullseye main" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian/ bullseye-updates main" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian-security bullseye-security main" >> /etc/apt/sources.list

# Install Go and build tools
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    tar \
    git \
    curl \
    vim \
    make \
    unzip \
    bash \
    && rm -rf /var/lib/apt/lists/* \
    && wget -O go.tar.gz "https://mirrors.aliyun.com/golang/go1.24.3.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

# Set Go environment
ENV PATH="/usr/local/go/bin:${PATH}" \
    GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code for building
COPY . .

# Build the specified service
RUN case "${BUILD_TARGET}" in \
      gateway) \
        echo "Building Gateway service..." && \
        go build -ldflags "-s -w" -o gateway ./cmd/gateway && \
        echo "gateway" > /app/service_name && \
        ln -sf gateway /app/service ;; \
      billing) \
        echo "Building Billing service..." && \
        go build -ldflags "-s -w" -o billing ./cmd/billing && \
        echo "billing" > /app/service_name && \
        ln -sf billing /app/service ;; \
      controller) \
        echo "Building Controller service..." && \
        go build -ldflags "-s -w" -o controller ./cmd/controller && \
        echo "controller" > /app/service_name && \
        ln -sf controller /app/service ;; \
      patrol) \
        echo "Building Patrol service..." && \
        go build -ldflags "-s -w" -o patrol ./cmd/patrol && \
        echo "patrol" > /app/service_name && \
        ln -sf patrol /app/service ;; \
      *) \
        echo "Unknown BUILD_TARGET: ${BUILD_TARGET}. Available: gateway, billing, controller, patrol" && \
        exit 1 ;; \
    esac

# Runtime stage
FROM debian:bullseye-slim

# Configure Aliyun Debian mirrors for runtime stage (using HTTP to avoid SSL issues)
RUN echo "deb http://mirrors.aliyun.com/debian/ bullseye main" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian/ bullseye-updates main" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian-security bullseye-security main" >> /etc/apt/sources.list

# Install common tools and clean up
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    wget \
    vim \
    bash \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

WORKDIR /app

# Copy the built service binary and service name
COPY --from=builder /app/service_name .
COPY --from=builder /app/service .

# Copy and setup entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Expose ports
EXPOSE 3000 9090

ENTRYPOINT ["/app/entrypoint.sh"] 