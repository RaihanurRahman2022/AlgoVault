# Backend & Frontend Restructuring - Quick Implementation Guide

## What's Been Done

✅ Created new folder structures for backend and frontend
✅ Created infrastructure layer:
- `pkg/config/config.go` - Configuration management
- `internal/shared/response/response.go` - Response utilities
- `internal/infrastructure/database/database.go` - Database connection
- `internal/infrastructure/database/schema.go` - Schema & migrations

## Simplified Approach

Given the complexity of full restructuring, I recommend a **hybrid approach**:

### Keep Current Structure + Add New Organization

Instead of moving all existing files (which risks breaking things), we'll:

1. **Keep existing files working** (`main.go`, `handlers.go`, `models.go`, `middleware.go`)
2. **Use new infrastructure** for shared code
3. **Add new features** using the new structure
4. **Gradually migrate** as needed

This way:
- ✅ Nothing breaks
- ✅ New infrastructure is ready
- ✅ Future features use clean structure
- ✅ Can migrate old code gradually

## Immediate Changes (Safe)

### 1. Update `go.mod` module path

The new structure uses `internal` and `pkg` folders, so we need to ensure imports work correctly.

**File**: `backend/go.mod`
```go
module algovault-backend

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.22
	golang.org/x/crypto v0.23.0
)
```

### 2. Update `main.go` to use new config

**Minimal change** - just use the config package:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"algovault-backend/pkg/config"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration using new config package
	cfg := config.Load()

	// Log configuration
	log.Printf("Starting server...")
	log.Printf("PORT environment variable: %s", os.Getenv("PORT"))
	log.Printf("Using port: %s", cfg.Port)

	// Check which database will be used
	if os.Getenv("DATABASE_URL") != "" {
		log.Printf("✅ Using PostgreSQL (DATABASE_URL is set)")
		log.Printf("Database: PostgreSQL via DATABASE_URL")
	} else {
		log.Printf("⚠️  Using SQLite (DATABASE_URL not set)")
		log.Printf("Database path: %s", cfg.DBPath)
		log.Printf("Note: SQLite data will be lost on Render restarts. Use DATABASE_URL for persistence.")
	}

	// Initialize database FIRST
	log.Printf("Initializing database...")
	db, err := NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("✅ Database ready and connected")

	// Initialize handlers
	handlers := &Handlers{
		DB:        db,
		JWTSecret: cfg.JWTSecret,
		AIAPIKey:  cfg.AIAPIKey,
	}

	// ... rest of main.go stays the same ...
}
```

## Frontend Structure (Simpler Approach)

For frontend, let's use a simpler reorganization:

### Current (Keep as-is for now)
```
frontend/
├── components/
├── pages/
├── services/
├── App.tsx
├── index.tsx
└── types.ts
```

### Add `src/` wrapper (Optional)

If you want cleaner organization:

1. Create `frontend/src/` folder
2. Move `App.tsx`, `index.tsx`, `types.ts`, `constants.tsx` into `src/`
3. Move `components/`, `pages/`, `services/` into `src/`
4. Update `index.html` to point to `/src/index.tsx`
5. Update `vite.config.ts` if needed

**But this is optional** - current structure works fine!

## Recommendation

### For Now (Next 30 minutes):
1. ✅ Keep all existing code as-is
2. ✅ Infrastructure is ready for future use
3. ✅ Test that everything still works
4. ✅ Deploy to Render with Supabase

### For Future (When adding learning resources):
1. Use new structure: `internal/domain/learning/`
2. Create clean separation
3. Gradually migrate old code if needed

## Benefits of This Approach

✅ **Zero risk** - Nothing breaks
✅ **Infrastructure ready** - Can use new packages
✅ **Future-proof** - New features use clean structure
✅ **Gradual migration** - Move code when convenient

## What to Do Next

1. **Test current code** - Make sure everything works
2. **Setup Supabase** - Get database persistence working
3. **Deploy** - Get app running in production
4. **Then** - Add new features using new structure

---

## Full Restructuring (If You Want It)

If you still want full restructuring, I can:

1. Create a PowerShell script to move all files
2. Update all imports automatically
3. Test each domain separately
4. Estimated time: 2-3 hours

**But I recommend the hybrid approach above for now!**

What would you like to do?
