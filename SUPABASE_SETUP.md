# Supabase Setup Guide for AlgoVault

## Step 1: Create Supabase Account & Database

1. Go to https://supabase.com
2. Click **"Start your project"** or **"Sign In"**
3. Sign up with GitHub (recommended) or email
4. Click **"New Project"**

### Project Configuration

Fill in the following:

- **Organization**: Create new or select existing
- **Name**: `algovault-db` (or any name you prefer)
- **Database Password**: 
  - Click "Generate a password" 
  - **IMPORTANT**: Copy and save this password securely!
  - You'll need it for the connection string
- **Region**: Choose closest to your location:
  - Asia: `Southeast Asia (Singapore)`
  - Europe: `West EU (Ireland)`
  - US: `East US (North Virginia)`
- **Pricing Plan**: Free (selected by default)

5. Click **"Create new project"**
6. Wait 2-3 minutes for database provisioning

---

## Step 2: Get Your Connection String

1. Once project is ready, go to **Settings** (gear icon in sidebar)
2. Click **Database** in the left menu
3. Scroll down to **Connection string** section
4. Select **URI** tab (not Session mode)
5. You'll see a connection string like:

```
postgresql://postgres.xxxxxxxxxxxxx:[YOUR-PASSWORD]@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
```

6. **Replace `[YOUR-PASSWORD]`** with the password you saved earlier
7. Copy the complete connection string

**Example**:
```
postgresql://postgres.abcdefghijklmnop:MySecurePass123!@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
```

---

## Step 3: Add Connection String to Render

1. Go to https://dashboard.render.com
2. Select your **algovault-backend** service
3. Click **Environment** tab in the left sidebar
4. Click **Add Environment Variable**
5. Add the following:
   - **Key**: `DATABASE_URL`
   - **Value**: (paste your Supabase connection string)
6. Click **Save Changes**

Render will automatically redeploy your service (takes 2-3 minutes).

---

## Step 4: Verify Database Connection

### Check Render Logs

1. In Render dashboard, go to **Logs** tab
2. Look for these messages after deployment:
   ```
   ✅ Using PostgreSQL (DATABASE_URL is set)
   Database: PostgreSQL via DATABASE_URL
   Initializing database...
   ✅ Database ready and connected
   ```

3. If you see errors, check:
   - Connection string is correct
   - Password doesn't have special characters that need URL encoding
   - Supabase project is active

### Test the API

1. **Register a new user**:
   ```bash
   curl -X POST https://your-app.onrender.com/api/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"test123","name":"Test User"}'
   ```

2. **Login**:
   ```bash
   curl -X POST https://your-app.onrender.com/api/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"test123"}'
   ```

3. **Create a category** (use token from login response):
   ```bash
   curl -X POST https://your-app.onrender.com/api/categories \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -d '{"name":"Test Category","icon":"Code","description":"Testing persistence"}'
   ```

4. **Verify persistence**:
   - Go to Render dashboard → Manual Deploy → Deploy latest commit
   - Wait for restart
   - Login again and check if category still exists
   - ✅ Data should persist!

---

## Step 5: Verify in Supabase Dashboard

1. Go to Supabase dashboard
2. Click **Table Editor** in sidebar
3. You should see your tables:
   - `users`
   - `categories`
   - `patterns`
   - `problems`
   - `solutions`

4. Click on `users` table to see your registered user
5. Click on `categories` to see your test category

---

## Troubleshooting

### Error: "connection refused"

**Solution**: Check if DATABASE_URL is correctly set in Render environment variables.

### Error: "password authentication failed"

**Solution**: 
1. Verify password in connection string matches Supabase password
2. If password has special characters, URL-encode them:
   - `@` → `%40`
   - `#` → `%23`
   - `$` → `%24`
   - `&` → `%26`

### Error: "database does not exist"

**Solution**: Make sure you're using the connection string from the **URI** tab, not Session mode.

### Data not persisting

**Solution**:
1. Check Render logs for "Using PostgreSQL" message
2. Verify DATABASE_URL is set in Render environment
3. Make sure you didn't include `-db ./algovault.db` in render.yaml startCommand

---

## Monitoring Your Database

### Check Storage Usage

1. Supabase dashboard → Settings → Usage
2. Monitor:
   - **Database size**: Should stay well under 500 MB
   - **Bandwidth**: Should stay under 2 GB/month
   - **Active connections**: Should be 1-2 for your app

### Set Up Alerts

1. Supabase dashboard → Settings → Billing
2. Enable email notifications for:
   - 80% storage usage
   - 80% bandwidth usage

---

## Backup Strategy (Optional)

### Manual Backup

Run this command locally to create a backup:

```bash
pg_dump "postgresql://postgres.xxxxx:password@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres" > backup_$(date +%Y%m%d).sql
```

### Automatic Backups

Supabase provides automatic daily backups on the free tier. You can restore from:
- Supabase dashboard → Database → Backups
- Point-in-time recovery available for last 7 days

---

## Next Steps

After successful setup:

1. ✅ Remove SQLite database files from backend (optional cleanup):
   ```bash
   cd backend
   rm algovault.db algovault.db-shm algovault.db-wal
   ```

2. ✅ Update `.gitignore` to exclude SQLite files (already done)

3. ✅ Test your frontend application thoroughly

4. ✅ Start adding your practice problems and learning resources!

---

## Connection String Format Reference

```
postgresql://[user]:[password]@[host]:[port]/[database]?[options]

Example:
postgresql://postgres.abcd1234:MyPass123@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres
           └─────┬─────┘ └───┬──┘ └────────────────────┬──────────────────────────┘ └──┬─┘ └──┬──┘
              user      password                    host                              port  database
```

---

## Support

If you encounter issues:

1. **Supabase Support**: https://supabase.com/docs
2. **Render Support**: https://render.com/docs
3. **Check Render Logs**: Dashboard → Your Service → Logs
4. **Check Supabase Logs**: Dashboard → Logs → Postgres Logs
