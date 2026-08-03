# Stage 1: Build the frontend (static site, no backend required)
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source (this includes the static Bible data under public/data/)
COPY frontend/ ./

# Build the frontend (copies public/ -> dist/ and bundles the Svelte app)
RUN npm run build

# Stage 2: Serve the static build with nginx
FROM nginx:alpine

# Copy the built frontend (SPA + static JSON data) into the web root
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html

# Minimal nginx config: serve static files with an SPA fallback.
# There is no API — all data is static JSON served directly from /data/.
COPY <<EOF /etc/nginx/conf.d/default.conf
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # SPA fallback - serve index.html for all other routes
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # Cache static assets and immutable JSON data aggressively
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|json)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
