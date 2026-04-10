FROM golang:latest

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y \
    fontconfig \
    libfreetype6 \
    libjpeg62-turbo \
    libxrender1 \
    libxext6 \
    && rm -rf /var/lib/apt/lists/*

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN go build -o main .

# Run
CMD ["./main"]