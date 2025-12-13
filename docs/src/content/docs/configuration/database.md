---
title: Database
description: Configure PostgreSQL for Kaimu
---

Kaimu uses PostgreSQL as its database. This guide covers database setup and configuration.

## Requirements

- PostgreSQL 14 or higher
- `uuid-ossp` extension (for UUID generation)

## Configuration

Configure the database connection via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DBHOST` | `localhost` | PostgreSQL host |
| `DBNAME` | `kaimu` | Database name |
| `DBUSERNAME` | `kaimu` | Database user |
| `DBPASSWORD` | *required* | Database password |
| `DBPORT` | `5432` | Database port |
| `DBSSL` | `disable` | SSL mode |

### SSL Modes

| Mode | Description |
|------|-------------|
| `disable` | No SSL (development only) |
| `require` | SSL required, no verification |
| `verify-ca` | SSL required, verify CA certificate |
| `verify-full` | SSL required, verify CA and hostname |

## Local Development

Use Docker Compose to run PostgreSQL locally:

```bash
docker compose up -d postgres
```

This creates a PostgreSQL instance with:
- User: `kaimu`
- Password: `mysecretpassword`
- Database: `kaimu`
- Port: `5432`

## Migrations

Kaimu uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Run Migrations

```bash
# Using go run
go run cmd/main.go migrate up

# Or in Docker
docker compose exec backend go run cmd/main.go migrate up
```

### Rollback Migrations

```bash
go run cmd/main.go migrate down
```

### Create New Migration

```bash
make migrate-create name=add_new_feature
```

This creates two files in `backend/db/migrations/`:
- `XXXXXX_add_new_feature.up.sql`
- `XXXXXX_add_new_feature.down.sql`

## Schema Overview

### Core Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts |
| `organizations` | Top-level organizations |
| `organization_members` | Organization membership |
| `projects` | Projects within organizations |
| `boards` | Kanban boards |
| `board_columns` | Board columns |
| `cards` | Cards on boards |
| `tags` | Tags for cards |
| `card_tags` | Card-tag relationships |

### Authentication Tables

| Table | Description |
|-------|-------------|
| `oidc_identities` | OIDC provider identities linked to users |

## Production Setup

### Managed PostgreSQL

For production, consider using a managed PostgreSQL service:

- **AWS RDS**
- **Google Cloud SQL**
- **Azure Database for PostgreSQL**
- **DigitalOcean Managed Databases**
- **Supabase**

Benefits:
- Automatic backups
- High availability
- Automatic updates
- Monitoring and alerting

### Connection String

For managed databases, you may receive a connection string. Extract the components:

```
postgresql://user:password@host:port/database?sslmode=require
```

Then set environment variables:
```bash
DBHOST=host
DBPORT=port
DBNAME=database
DBUSERNAME=user
DBPASSWORD=password
DBSSL=require
```

### Connection Pooling

For high-traffic deployments, consider using a connection pooler like PgBouncer:

```yaml
# docker-compose.prod.yaml
services:
  pgbouncer:
    image: edoburu/pgbouncer
    environment:
      DATABASE_URL: postgresql://user:pass@postgres:5432/kaimu
      POOL_MODE: transaction
      MAX_CLIENT_CONN: 100
```

## Backup & Recovery

### Manual Backup

```bash
pg_dump -h localhost -U kaimu -d kaimu > backup.sql
```

### Restore

```bash
psql -h localhost -U kaimu -d kaimu < backup.sql
```

### Automated Backups

For production, set up automated backups:

1. **Managed services** usually include automatic backups
2. **Self-hosted**: Use `pg_dump` with cron or a backup service

Example cron job (daily backup):
```bash
0 2 * * * pg_dump -h localhost -U kaimu -d kaimu | gzip > /backups/kaimu_$(date +\%Y\%m\%d).sql.gz
```

## Troubleshooting

### Connection Refused

- Verify PostgreSQL is running: `docker compose ps`
- Check the host and port configuration
- Ensure the firewall allows connections

### Authentication Failed

- Verify username and password
- Check that the user has access to the database

### SSL Required

If connecting to a managed database:
```bash
DBSSL=require
```

### Extension Not Found

If `uuid-ossp` is missing:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```
