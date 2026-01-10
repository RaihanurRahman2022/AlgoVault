# Deployment Guide

This guide covers deploying AlgoVault to free hosting platforms.

## Architecture Overview

- **Frontend**: React + TypeScript + Vite
- **Backend**: Go (Golang) with SQLite/PostgreSQL
- **Authentication**: JWT tokens
- **Database**: SQLite (dev) or PostgreSQL (production)

## Free Hosting Options

### Frontend Deployment

#### Option 1: Vercel (Recommended)
- **Free Tier**: Unlimited projects, 100GB bandwidth
- **Setup**:
  1. Push code to GitHub
  2. Import project on [Vercel](https://vercel.com)
  3. Set build command: `cd frontend && npm install && npm run build`
  4. Set output directory: `frontend/dist`
  5. Add environment variable: `VITE_API_BASE_URL` (your backend URL)

#### Option 2: Netlify
- **Free Tier**: 100GB bandwidth, 300 build minutes/month
- **Setup**:
  1. Push code to GitHub
  2. Import project on [Netlify](https://netlify.com)
  3. Build command: `cd frontend && npm install && npm run build`
  4. Publish directory: `frontend/dist`
  5. Add environment variable: `VITE_API_BASE_URL`

#### Option 3: GitHub Pages
- **Free Tier**: Unlimited
- **Setup**:
  1. Update `vite.config.ts` to set `base: '/your-repo-name/'`
  2. Build: `cd frontend && npm run build`
  3. Deploy `frontend/dist` to `gh-pages` branch

### Backend Deployment

#### Option 1: Railway (Recommended)
- **Free Tier**: $5 credit/month (enough for small apps)
- **Setup**:
  1. Push code to GitHub
  2. Create new project on [Railway](https://railway.app)
  3. Add PostgreSQL service (free tier available)
  4. Set environment variables:
     - `JWT_SECRET`: Generate a strong secret
     - `DATABASE_URL`: Provided by Railway PostgreSQL
  5. Update backend to use PostgreSQL (see below)

#### Option 2: Render
- **Free Tier**: 750 hours/month
- **Setup**:
  1. Connect GitHub repo on [Render](https://render.com)
  2. Create PostgreSQL database (free tier)
  3. Set build command: `cd backend && go mod download && go build -o server`
  4. Set start command: `./server`
  5. Add environment variables:
     - `JWT_SECRET`
     - `DATABASE_URL`

#### Option 3: Fly.io
- **Free Tier**: 3 shared VMs
- **Setup**:
  1. Install Fly CLI: `curl -L https://fly.io/install.sh | sh`
  2. Run `fly launch` in backend directory
  3. Add PostgreSQL: `fly postgres create`
  4. Set secrets: `fly secrets set JWT_SECRET=your-secret`

### Database Options

#### For Production (PostgreSQL)

1. **Update `backend/models.go`** to support PostgreSQL:
   ```go
   // Replace sqlite3 with postgres driver
   import _ "github.com/lib/pq"
   
   // Update connection string
   db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
   ```

2. **Update `go.mod`**:
   ```go
   require github.com/lib/pq v1.10.9
   ```

3. **Update SQL syntax** (PostgreSQL uses different syntax):
   - Replace `TEXT` with `VARCHAR` or `TEXT`
   - Replace `DATETIME` with `TIMESTAMP`
   - Replace `INSERT OR REPLACE` with `ON CONFLICT`

#### Free PostgreSQL Hosting:
- **Supabase**: 500MB database, 2GB bandwidth (free tier)
- **Neon**: 3GB storage, unlimited projects (free tier)
- **Railway**: Included with Railway deployment
- **Render**: Included with Render deployment

## Step-by-Step Deployment

### 1. Prepare Backend for PostgreSQL

Create `backend/database.go`:
```go
package main

import (
    "database/sql"
    "os"
    _ "github.com/lib/pq"
    _ "github.com/mattn/go-sqlite3"
)

func NewDatabase() (*Database, error) {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        // Fallback to SQLite for local dev
        dbURL = "sqlite3:./algovault.db"
    }
    
    var db *sql.DB
    var err error
    
    if os.Getenv("DATABASE_URL") != "" {
        // PostgreSQL
        db, err = sql.Open("postgres", dbURL)
    } else {
        // SQLite
        db, err = sql.Open("sqlite3", "./algovault.db")
    }
    
    if err != nil {
        return nil, err
    }
    
    // ... rest of initialization
}
```

### 2. Environment Variables

**Frontend** (`.env` in frontend directory):
```
VITE_API_BASE_URL=https://your-backend-url.com
```

**Backend**:
```
JWT_SECRET=your-super-secret-key-change-this
DATABASE_URL=postgres://user:pass@host:5432/dbname
PORT=8080
```

### 3. Update API Base URL

Update `frontend/services/apiService.ts`:
```typescript
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';
```

### 4. CORS Configuration

Update `backend/main.go` CORS to allow your frontend domain:
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "https://your-frontend.vercel.app",
            "http://localhost:3000",
        }
        
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }
        // ... rest of CORS
    })
}
```

## Quick Start Commands

### Local Development

**Backend**:
```bash
cd backend
go mod download
go run .
```

**Frontend**:
```bash
cd frontend
npm install
npm run dev
```

### Production Build

**Backend**:
```bash
cd backend
go build -o server
./server
```

**Frontend**:
```bash
cd frontend
npm install
npm run build
# Deploy frontend/dist to hosting
```

## Recommended Free Stack

1. **Frontend**: Vercel
2. **Backend**: Railway
3. **Database**: Railway PostgreSQL (included)

**Total Cost**: $0/month

## Security Checklist

- [ ] Change `JWT_SECRET` to a strong random string
- [ ] Use HTTPS in production
- [ ] Set proper CORS origins
- [ ] Use environment variables for secrets
- [ ] Enable database backups (Railway/Render do this automatically)
- [ ] Use strong passwords for database

## Troubleshooting

### CORS Errors
- Ensure backend CORS allows your frontend domain
- Check that API_BASE_URL is correct

### Database Connection Errors
- Verify DATABASE_URL format
- Check database credentials
- Ensure database is accessible from backend host

### Build Errors
- Check Node.js version (18+)
- Check Go version (1.21+)
- Verify all dependencies are installed

## Support

For issues, check:
- Backend logs on hosting platform
- Frontend build logs
- Browser console for frontend errors
- Network tab for API errors
