---
title: Self-Hosting
description: Deploy Pulse on your own infrastructure
---

This guide covers deploying Pulse on your own infrastructure for production use.

## Architecture Overview

Pulse consists of:

- **Frontend**: Astro/Svelte static site (can be served via CDN)
- **Backend**: Go GraphQL API
- **Database**: PostgreSQL

```
                    ┌─────────────┐
                    │   Reverse   │
                    │   Proxy     │
                    │  (nginx)    │
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ Frontend │    │ Backend  │    │ OIDC     │
    │  :4321   │    │  :3000   │    │ Provider │
    └──────────┘    └────┬─────┘    └──────────┘
                         │
                         ▼
                   ┌──────────┐
                   │PostgreSQL│
                   │  :5432   │
                   └──────────┘
```

## Deployment Options

### Option 1: Docker Compose (Recommended)

Best for: Small to medium deployments, single server

```yaml
# docker-compose.prod.yaml
version: '3.8'

services:
  postgres:
    image: postgres:16
    restart: always
    environment:
      POSTGRES_USER: pulse
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: pulse
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pulse"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    image: pulse-backend:latest
    restart: always
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DBHOST: postgres
      DBNAME: pulse
      DBUSERNAME: pulse
      DBPASSWORD: ${DB_PASSWORD}
      DBSSL: disable
      ENV: production
      JWT_SECRET: ${JWT_SECRET}
      OIDC_BASE_URL: https://api.yourdomain.com
      OIDC_FRONTEND_URL: https://yourdomain.com
      OIDC_PROVIDERS: ${OIDC_PROVIDERS}
    ports:
      - "127.0.0.1:3000:3000"

  frontend:
    image: pulse-frontend:latest
    restart: always
    environment:
      PUBLIC_API_URL: https://api.yourdomain.com/graphql
    ports:
      - "127.0.0.1:4321:4321"

volumes:
  postgres_data:
```

### Option 2: Kubernetes

Best for: Large deployments, high availability

Create Kubernetes manifests for each component:

```yaml
# backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pulse-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pulse-backend
  template:
    metadata:
      labels:
        app: pulse-backend
    spec:
      containers:
      - name: backend
        image: pulse-backend:latest
        ports:
        - containerPort: 3000
        env:
        - name: DBPASSWORD
          valueFrom:
            secretKeyRef:
              name: pulse-secrets
              key: db-password
        # ... other env vars
```

### Option 3: Platform as a Service

Deploy to platforms like:
- **Railway**
- **Render**
- **Fly.io**
- **DigitalOcean App Platform**

## Infrastructure Requirements

### Minimum Requirements

- **CPU**: 1 vCPU
- **RAM**: 2 GB
- **Storage**: 10 GB

### Recommended for Production

- **CPU**: 2+ vCPU
- **RAM**: 4+ GB
- **Storage**: 50+ GB SSD

### Database

- PostgreSQL 14+ (managed recommended)
- Enable automatic backups
- Consider read replicas for high traffic

## SSL/TLS Configuration

### Using Let's Encrypt with Certbot

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d yourdomain.com -d api.yourdomain.com
```

### Nginx Configuration

```nginx
# Frontend
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:4321;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}

# Backend API
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Environment Setup

Create a `.env` file for secrets:

```bash
# .env
DB_PASSWORD=your-secure-database-password
JWT_SECRET=your-256-bit-random-secret
OIDC_PROVIDERS='[{"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."}]'
```

Generate a secure JWT secret:
```bash
openssl rand -base64 32
```

## Backup Strategy

### Database Backups

```bash
# Daily backup script
#!/bin/bash
BACKUP_DIR=/backups
DATE=$(date +%Y%m%d)
pg_dump -h localhost -U pulse -d pulse | gzip > $BACKUP_DIR/pulse_$DATE.sql.gz

# Keep last 30 days
find $BACKUP_DIR -name "pulse_*.sql.gz" -mtime +30 -delete
```

### Automated Backups

Add to crontab:
```bash
0 2 * * * /opt/scripts/backup.sh
```

## Monitoring

### Health Checks

Backend health endpoint:
```bash
curl https://api.yourdomain.com/health
```

### Logging

Configure structured logging:
```bash
ENV: production  # Enables JSON logging
```

View logs:
```bash
docker compose logs -f backend
```

### Metrics (Optional)

Pulse exposes Prometheus metrics at `/metrics`. Configure Prometheus to scrape:

```yaml
scrape_configs:
  - job_name: 'pulse'
    static_configs:
      - targets: ['localhost:3000']
```

## Updates

### Rolling Updates with Docker Compose

```bash
# Pull new images
docker compose pull

# Restart with new images
docker compose up -d

# Run migrations if needed
docker compose exec backend go run cmd/main.go migrate up
```

### Zero-Downtime Updates

1. Deploy new version alongside old
2. Run migrations
3. Switch traffic to new version
4. Remove old version

## Security Checklist

- [ ] Use HTTPS everywhere
- [ ] Set strong `JWT_SECRET`
- [ ] Use strong database password
- [ ] Enable database SSL in production
- [ ] Configure firewall rules
- [ ] Keep secrets out of version control
- [ ] Enable automatic security updates
- [ ] Set up monitoring and alerting
- [ ] Configure regular backups
- [ ] Review OIDC provider settings
