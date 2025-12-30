# Deployment Guide

## Pre-Deployment Checklist

### ✅ Configuration Files Enhanced

I've improved the configuration with:

1. **Production Validation** - Automatically validates required settings
2. **Environment Detection** - Distinguishes dev/staging/production
3. **Security Checks** - Ensures strong JWT secrets and SSL in production
4. **Better CORS Parsing** - Handles whitespace in origin lists

### 🔒 Security Improvements

- Production mode requires strong JWT secret
- Production mode requires database SSL
- Warns if MinIO SSL is disabled in production
- Validates all required environment variables

---

## Deployment Steps

### 1. **Prepare Production Environment File**

```bash
cd c:\workspace\pocs\starving\v0-hungry-food-sharing-app-backend
cp .env.production.example .env
```

Edit `.env` with your production values:

```env
APP_ENV=production

# Database - Use your production PostgreSQL
DB_HOST=your-db-host.example.com
DB_PORT=5432
DB_USER=food_sharing_user
DB_PASSWORD=your-strong-password
DB_NAME=food_sharing
DB_SSLMODE=require  # IMPORTANT: Use SSL in production

# MinIO - Use your production MinIO/S3
MINIO_ENDPOINT=your-minio.example.com:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_BUCKET=food-images
MINIO_USE_SSL=true  # IMPORTANT: Use SSL in production
MINIO_PUBLIC_URL=https://your-minio.example.com

# JWT - Generate a strong random secret
JWT_SECRET=your-very-strong-random-secret-at-least-32-characters
JWT_EXPIRATION=24h

# Server
PORT=8000
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 2. **Generate Strong JWT Secret**

Use one of these methods:

**Option A - OpenSSL:**
```bash
openssl rand -base64 32
```

**Option B - Go:**
```bash
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

**Option C - Online (use with caution):**
```
https://www.random.org/strings/
```

### 3. **Set Up Production Database**

**Create database and user:**
```sql
CREATE DATABASE food_sharing;
CREATE USER food_sharing_user WITH ENCRYPTED PASSWORD 'your-strong-password';
GRANT ALL PRIVILEGES ON DATABASE food_sharing TO food_sharing_user;
```

**Run migrations:**
```bash
migrate -path migrations \
  -database "postgres://food_sharing_user:your-password@your-host:5432/food_sharing?sslmode=require" \
  up
```

### 4. **Set Up MinIO/S3**

**Option A - MinIO:**
- Install MinIO server
- Create bucket: `food-images`
- Set bucket policy to public read
- Enable SSL with valid certificate

**Option B - AWS S3:**
- Create S3 bucket
- Configure public read access
- Use S3 endpoint in `MINIO_ENDPOINT`
- Use AWS credentials for access/secret keys

### 5. **Build for Production**

```bash
# Build binary
go build -o food-sharing-backend main.go

# Or use Makefile
make build
```

### 6. **Deploy Options**

#### **Option A: Docker (Recommended)**

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8000
CMD ["./server"]
```

Create `docker-compose.yml`:
```yaml
version: '3.8'
services:
  backend:
    build: .
    ports:
      - "8000:8000"
    env_file:
      - .env
    depends_on:
      - postgres
      - minio
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: food_sharing
      POSTGRES_USER: food_sharing_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    restart: unless-stopped

volumes:
  postgres_data:
  minio_data:
```

Deploy:
```bash
docker-compose up -d
```

#### **Option B: Systemd Service**

Create `/etc/systemd/system/food-sharing-backend.service`:
```ini
[Unit]
Description=Food Sharing Backend API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/food-sharing-backend
EnvironmentFile=/opt/food-sharing-backend/.env
ExecStart=/opt/food-sharing-backend/server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable food-sharing-backend
sudo systemctl start food-sharing-backend
sudo systemctl status food-sharing-backend
```

#### **Option C: Cloud Platforms**

**Heroku:**
```bash
heroku create your-app-name
heroku addons:create heroku-postgresql:hobby-dev
heroku config:set APP_ENV=production
heroku config:set JWT_SECRET=your-secret
# ... set other env vars
git push heroku main
```

**Google Cloud Run:**
```bash
gcloud builds submit --tag gcr.io/PROJECT-ID/food-sharing-backend
gcloud run deploy --image gcr.io/PROJECT-ID/food-sharing-backend --platform managed
```

**AWS ECS/Fargate:**
- Build Docker image
- Push to ECR
- Create ECS task definition
- Deploy to Fargate

### 7. **Set Up Reverse Proxy (Nginx)**

Create `/etc/nginx/sites-available/food-sharing-api`:
```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Enable and restart:
```bash
sudo ln -s /etc/nginx/sites-available/food-sharing-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 8. **Enable SSL with Let's Encrypt**

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
```

### 9. **Verify Deployment**

**Health check:**
```bash
curl https://api.yourdomain.com/health
```

**Test registration:**
```bash
curl -X POST https://api.yourdomain.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"password123"}'
```

---

## Environment-Specific Configuration

### Development
```env
APP_ENV=development
DB_SSLMODE=disable
MINIO_USE_SSL=false
ALLOWED_ORIGINS=http://localhost:3000
```

### Staging
```env
APP_ENV=staging
DB_SSLMODE=require
MINIO_USE_SSL=true
ALLOWED_ORIGINS=https://staging.yourdomain.com
```

### Production
```env
APP_ENV=production
DB_SSLMODE=require
MINIO_USE_SSL=true
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

---

## Monitoring & Maintenance

### Logs
```bash
# Systemd
sudo journalctl -u food-sharing-backend -f

# Docker
docker-compose logs -f backend
```

### Database Backups
```bash
# Automated daily backup
pg_dump -h your-host -U food_sharing_user food_sharing > backup_$(date +%Y%m%d).sql
```

### Health Monitoring
Set up monitoring for:
- `/health` endpoint (should return 200)
- Database connection
- MinIO availability
- API response times

---

## Troubleshooting

### Config Validation Errors

If you see validation errors on startup:

```
Error: JWT_SECRET must be set to a strong secret in production
```

**Solution:** Set a strong JWT secret in your `.env` file

```
Error: DB_SSLMODE should not be 'disable' in production
```

**Solution:** Change `DB_SSLMODE=require` in production

### Database Connection Issues

```
Error: failed to ping database
```

**Check:**
1. Database host/port are correct
2. Firewall allows connections
3. SSL mode matches database configuration
4. User has proper permissions

### MinIO Connection Issues

```
Error: failed to create MinIO client
```

**Check:**
1. MinIO endpoint is accessible
2. Access key and secret key are correct
3. SSL certificate is valid (if using SSL)
4. Bucket exists or app has permission to create it

---

## Security Best Practices

1. ✅ **Never commit `.env` files** - Already in `.gitignore`
2. ✅ **Use strong JWT secrets** - At least 32 random characters
3. ✅ **Enable SSL** - For database and MinIO in production
4. ✅ **Restrict CORS** - Only allow your frontend domain
5. ✅ **Use environment variables** - Never hardcode secrets
6. ✅ **Regular updates** - Keep dependencies updated
7. ✅ **Database backups** - Automated daily backups
8. ✅ **Rate limiting** - Consider adding rate limiting middleware
9. ✅ **Monitoring** - Set up alerts for errors and downtime

---

## Configuration Validation

The enhanced config now automatically validates:

- ✅ Required environment variables are set
- ✅ JWT secret is strong in production
- ✅ Database SSL is enabled in production
- ✅ CORS origins are configured
- ✅ All connection strings are valid

**Test validation:**
```bash
# This will fail if config is invalid
go run main.go
```

You're now ready to deploy! 🚀
