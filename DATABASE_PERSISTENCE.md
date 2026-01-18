# Database Persistence Solution

## Problem
On Render (and similar platforms), the filesystem is **ephemeral** - it gets wiped on every deployment. This means your SQLite database (`algovault.db`) is lost on each rebuild.

## Solution: Use Render's Free PostgreSQL ✅

Render offers **free PostgreSQL** (1GB storage, 90 days retention) which is **persistent** and survives deployments.

## Quick Setup (5 minutes)

### Step 1: Create PostgreSQL Database on Render

1. Go to your Render dashboard: https://dashboard.render.com
2. Click "New +" → "PostgreSQL"
3. Name it: `algovault-db`
4. Select **Free** plan
5. Click "Create Database"
6. Wait 1-2 minutes for it to be created
7. Copy the **Internal Database URL** (looks like: `postgresql://user:pass@host:5432/dbname`)
   - **Important**: Use the **Internal Database URL**, not the External one
   - It should look like: `postgresql://algovault_db_user:password@dpg-xxxxx-a.oregon-postgres.render.com/algovault_db`

### Step 2: Update Environment Variables

1. Go to your backend service on Render (the web service, not the database)
2. Go to "Environment" tab
3. Add new environment variable:
   - **Key**: `DATABASE_URL`
   - **Value**: The Internal Database URL you copied (starts with `postgresql://`)
4. Click "Save Changes"
5. Render will automatically redeploy (wait 2-3 minutes)

### Step 3: Code is Already Updated! ✅

The code has been updated to:
- ✅ Use PostgreSQL when `DATABASE_URL` is set
- ✅ Fall back to SQLite for local development
- ✅ Automatically migrate your schema
- ✅ Handle both SQLite (`?`) and PostgreSQL (`$1, $2`) placeholders

### Step 4: Verify It Works

1. After deployment completes, check the logs
2. You should see: "Database initialized successfully"
3. Your data will now persist across deployments! ✅

## How It Works

- **Local Development**: Uses SQLite (`./algovault.db`) - no changes needed
- **Production (Render)**: Uses PostgreSQL (persistent, survives deployments)
- **Automatic**: Just set `DATABASE_URL` environment variable

## Migration from SQLite to PostgreSQL

If you already have data in SQLite and want to migrate:

1. **Export your SQLite data** (optional, if you have existing data):
   ```bash
   # On your local machine
   sqlite3 algovault.db .dump > backup.sql
   ```

2. **After setting up PostgreSQL on Render**, you can:
   - Start fresh (recommended for small datasets)
   - Or manually import data using `psql` if needed

3. **The demo user will be created automatically** on first run

## Troubleshooting

### "No connection could be made"
- Make sure you're using the **Internal Database URL**, not External
- Check that the database service is running (green status)

### "relation does not exist"
- The schema is created automatically on first run
- Check the logs for any errors during initialization

### Still using SQLite?
- Make sure `DATABASE_URL` is set in your Render environment variables
- After adding it, Render will redeploy automatically

## Cost: Still $0/month ✅

PostgreSQL on Render's free tier:
- ✅ 1GB storage (plenty for 100+ problems)
- ✅ 90 days data retention
- ✅ No credit card required
- ✅ Perfect for your use case
