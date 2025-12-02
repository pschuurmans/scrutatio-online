# Stage 1: Build the frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build the frontend
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.25.4-alpine AS backend-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bijbel-api ./cmd/api

# Stage 3: Final image with nginx + Go backend
FROM alpine:latest

WORKDIR /app

# Install nginx and supervisor to run both processes
RUN apk --no-cache add nginx supervisor ca-certificates

# Copy the Go binary
COPY --from=backend-builder /bijbel-api .

# Copy internal data files (books.json, crossref files, etc.)
COPY --from=backend-builder /app/internal ./internal

# Copy the frontend dist to nginx html folder
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html

# Create nginx config
RUN mkdir -p /etc/nginx/http.d
COPY <<EOF /etc/nginx/http.d/default.conf
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # API proxy to Go backend
    location /api/ {
        proxy_pass http://127.0.0.1:3000/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # SPA fallback - serve index.html for all other routes
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

# Create supervisor config to run both nginx and the Go app
COPY <<EOF /etc/supervisord.conf
[supervisord]
nodaemon=true
logfile=/dev/null
logfile_maxbytes=0

[program:nginx]
command=nginx -g "daemon off;"
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:api]
command=/app/bijbel-api
directory=/app
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

# Expose port 80 (nginx)
EXPOSE 80

# Run supervisor to manage both processes
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
