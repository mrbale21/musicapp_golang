# CORS & Ngrok Configuration Guide

## Masalah: POST Returns 404 tapi GET /health OK

Ini adalah error klasik yang terjadi karena:

1. **CORS Preflight gagal** - Browser mengirim OPTIONS request dulu, jika gagal browser abort POST
2. **Route tidak terdaftar** - Tapi ini jarang karena GET /health sudah bekerja
3. **Middleware block** - JWT atau CORS middleware block request

## Cara Kerja Request POST di Browser

```
Browser: OPTIONS /api/auth/register (preflight)
         ↓ (jika 200, lanjut ke)
Browser: POST /api/auth/register (actual request)
```

Jika step 1 gagal (404, 403, dll), step 2 tidak dieksekusi.

## Solution untuk Ngrok + Production

### 1. Development Mode (Recommended untuk testing)

```bash
# .env
ENV=development
```

**Hasil**: CORS allow semua origin (nyaman untuk development)

```go
// internal/routes/routes.go
corsConfig.AllowOriginFunc = func(origin string) bool {
    return true  // ✅ Accept semua
}
```

### 2. Production Mode dengan Auto-Detect Ngrok

```bash
# .env
ENV=production
CORS_ORIGIN=  # (kosong = auto-detect)
```

**Hasil**: Auto detect ngrok + vercel domains

```go
corsConfig.AllowOriginFunc = func(origin string) bool {
    return strings.Contains(origin, "ngrok") ||
           strings.Contains(origin, "vercel.app")
}
```

### 3. Production Mode dengan Specific Domain

```bash
# .env
ENV=production
CORS_ORIGIN=https://my-frontend.vercel.app
```

**Hasil**: HANYA allow specific domain (paling aman)

## Testing Steps

### Step 1: Test OPTIONS Preflight

```bash
curl -X OPTIONS https://xxxx-free.ngrok-free.dev/api/auth/register \
  -H "Origin: https://your-frontend.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v

# Output yang OK:
# < HTTP/1.1 200 OK
# < Access-Control-Allow-Origin: https://your-frontend.com
# < Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

### Step 2: Test POST Request

```bash
curl -X POST https://xxxx-free.ngrok-free.dev/api/auth/register \
  -H "Content-Type: application/json" \
  -H "Origin: https://your-frontend.com" \
  -d '{
    "username": "testuser",
    "email": "test@test.com",
    "password": "password123"
  }' \
  -v
```

### Step 3: Debug Endpoint

```bash
# List semua routes
curl https://xxxx-free.ngrok-free.dev/api/debug/routes

# Test POST capability
curl -X POST https://xxxx-free.ngrok-free.dev/api/debug/test-post \
  -H "Content-Type: application/json"
```

## Frontend Configuration

### React/Fetch API

```javascript
const API_URL = "https://xxxx-free.ngrok-free.dev/api";

// Register
const response = await fetch(`${API_URL}/auth/register`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    // PENTING: Include Origin header (browser auto add ini)
  },
  body: JSON.stringify({
    username: "testuser",
    email: "test@test.com",
    password: "password123",
  }),
});
```

### Axios Configuration

```javascript
import axios from "axios";

const api = axios.create({
  baseURL: "https://xxxx-free.ngrok-free.dev/api",
  headers: {
    "Content-Type": "application/json",
  },
});

// Register
api.post("/auth/register", {
  username: "testuser",
  email: "test@test.com",
  password: "password123",
});
```

## Common Mistakes

### ❌ Mistake 1: Missing Content-Type Header

```javascript
// SALAH
fetch(`${API_URL}/auth/register`, {
  method: "POST",
  body: JSON.stringify(data),
  // ← Missing Content-Type!
});

// BENAR
fetch(`${API_URL}/auth/register`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify(data),
});
```

### ❌ Mistake 2: Wrong Frontend URL in .env

```bash
# SALAH - Tidak match dengan origin
CORS_ORIGIN=http://localhost:3000
# (tapi frontend kirim origin: https://my-app.com)

# BENAR
CORS_ORIGIN=https://my-app.com
```

### ❌ Mistake 3: Ngrok URL Expired

```bash
# Ngrok generate URL baru setiap kali di-restart
ngrok http 8080
# https://xxxx-free.ngrok-free.dev ← URL baru!

# Jika menjadi 404, kemungkinan tunnel sudah restart
# Update frontend dengan URL baru
```

## Production Checklist

- [ ] ENV=production di Railway
- [ ] CORS_ORIGIN set ke frontend domain (atau kosong untuk auto)
- [ ] JWT_SECRET set ke value yang aman
- [ ] Database credentials correct
- [ ] Test `/health` endpoint dulu
- [ ] Test `/api/debug/routes` untuk verify routes
- [ ] Test POST /api/auth/register dari frontend
- [ ] Monitoring logs untuk error

## Useful Links

- Gin CORS Middleware: https://github.com/gin-contrib/cors
- Ngrok Documentation: https://ngrok.com/docs
- MDN CORS: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
