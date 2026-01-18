# Render Deployment Instructions

## Step 1: Add DATABASE_URL to Render

1. Go to https://dashboard.render.com
2. Click on your **algovault-backend** service
3. Click **Environment** tab in the left sidebar
4. Click **Add Environment Variable** button
5. Add the following:

   **Key**: `DATABASE_URL`
   
   **Value**: 
   ```
   postgresql://postgres:%5B.YGd!9cmZ%40xxs7e%5D@db.houanevvurgimlngkrqz.supabase.co:5432/postgres
   ```

6. Click **Save Changes**

## Step 2: Wait for Automatic Deployment

Render will automatically redeploy your service (takes 2-3 minutes).

## Step 3: Verify Deployment

### Check Logs

1. In Render dashboard, go to **Logs** tab
2. Look for these success messages:

```
✅ Using PostgreSQL (DATABASE_URL is set)
Database: PostgreSQL via DATABASE_URL
Initializing database...
✅ Database ready and connected
Starting HTTP server on 0.0.0.0:8080
```

### If You See Errors

**Error: "password authentication failed"**
- The URL encoding might be wrong
- Try this alternative encoding:
  ```
  postgresql://postgres:%5B.YGd%219cmZ%40xxs7e%5D@db.houanevvurgimlngkrqz.supabase.co:5432/postgres
  ```

**Error: "connection refused"**
- Check if DATABASE_URL is correctly set
- Verify Supabase project is active

## Step 4: Test the API

### Test 1: Health Check

```bash
curl https://your-app.onrender.com/health
```

Expected: `OK`

### Test 2: Register a User

```bash
curl -X POST https://your-app.onrender.com/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'
```

Expected: JSON with token and user data

### Test 3: Login

```bash
curl -X POST https://your-app.onrender.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

Expected: JSON with token

### Test 4: Create Category (with token from login)

```bash
curl -X POST https://your-app.onrender.com/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"name":"Test Category","icon":"Code","description":"Testing persistence"}'
```

Expected: JSON with created category

## Step 5: Test Persistence

1. Create a category via your frontend or API
2. Go to Render dashboard → **Manual Deploy** → **Deploy latest commit**
3. Wait for restart (2-3 minutes)
4. Check if the category still exists
5. ✅ If it exists, persistence is working!

## Step 6: Verify in Supabase

1. Go to https://supabase.com
2. Open your `algovault-db` project
3. Click **Table Editor** in sidebar
4. You should see tables:
   - `users`
   - `categories`
   - `patterns`
   - `problems`
   - `solutions`
5. Click on tables to see your data

## Troubleshooting

### Password Encoding Issues

If authentication fails, try these encoded versions:

**Version 1** (brackets and @ encoded):
```
postgresql://postgres:%5B.YGd!9cmZ%40xxs7e%5D@db.houanevvurgimlngkrqz.supabase.co:5432/postgres
```

**Version 2** (all special chars encoded):
```
postgresql://postgres:%5B.YGd%219cmZ%40xxs7e%5D@db.houanevvurgimlngkrqz.supabase.co:5432/postgres
```

### Still Not Working?

1. Check Render logs for exact error message
2. Verify Supabase project is active (not paused)
3. Check if DATABASE_URL is visible in Render environment variables
4. Try connecting to Supabase from local machine first

## Success Indicators

✅ Render logs show "Using PostgreSQL"
✅ No database connection errors
✅ Can register and login users
✅ Data persists after Render restart
✅ Can see data in Supabase Table Editor

---

**Once everything works, you'll have:**
- ✅ Persistent database storage
- ✅ No data loss on deployments
- ✅ 500MB free storage
- ✅ Production-ready setup
