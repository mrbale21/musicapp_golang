# Music App Backend

Backend API untuk aplikasi musik dengan fitur rekomendasi berbasis content, collaborative, dan hybrid.

## Setup Development Lokal

### 1. Prerequisites

- Go 1.24+
- PostgreSQL
- Git

### 2. Clone Repository

```bash
git clone <repository-url>
cd musicapp_golang
```

### 3. Setup Database Lokal

```bash
# Buat database PostgreSQL
createdb music_app

# Atau via psql
psql -U postgres
CREATE DATABASE music_app;
\q
```

### 4. Environment Variables

```bash
# Copy file environment
cp .env.example .env

# Edit .env sesuai konfigurasi lokal Anda
# Pastikan DB_PASSWORD sesuai dengan password PostgreSQL Anda
# Untuk YouTube API: https://console.developers.google.com/
```

### 5. YouTube API Setup (Opsional - Recommended)

Untuk fitur audio YouTube yang lebih baik:

1. **Buat Google Cloud Project**: https://console.cloud.google.com/
2. **Enable YouTube Data API v3**
3. **Buat API Key** di Credentials
4. **Set di .env**:
   ```
   YOUTUBE_API_KEY=your-api-key-here
   ```

**Tanpa API Key**: Sistem akan mencoba berbagai strategi pencarian otomatis
**Dengan API Key**: Pencarian lebih akurat dengan filtering cerdas

### 5. Install Dependencies

```bash
go mod download
```

### 6. Run Server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080` dan bisa diakses dari network di `http://[IP-KOMPUTER]:8080`

**🌱 First-time Setup - Automatic Seeding:**
Saat pertama kali app dijalankan di mode development, sistem otomatis akan:

- Membuat table database (migration)
- Membuat akun admin default:
  - **Email**: `admin@musicapp.local`
  - **Password**: `admin123`
  - ⚠️ **GANTI PASSWORD INI DI PRODUCTION!**

## Troubleshooting

### ❌ POST /api/auth/register returns 404 di Production/Ngrok

**Masalah**: Bio POST requests (register, login) return 404 tapi GET /health works fine.

**Penyebab Umum**:

1. **CORS Preflight Failed**: Browser kirim OPTIONS request terlebih dahulu. Jika gagal, browser tidak lanjutkan POST
2. **Ngrok Configuration**: Ngrok URL tidak di-config di CORS_ORIGIN
3. **Content-Type Header**: Missing Content-Type header bisa menyebabkan CORS rejection

**Solusi**:

**Untuk Development (Ngrok Testing)**:

```bash
# Set environment ke development (CORS allow semua)
ENV=development

# Atau set specific ngrok URL (otomatis detect ngrok domain)
ENV=production
```

**Untuk Production (Vercel/Frontend Tertentu)**:

```bash
ENV=production
CORS_ORIGIN=https://your-frontend.vercel.app
```

**Testing dengan cURL**:

```bash
# Kasi Content-Type header
curl -X POST https://your-ngrok-url.ngrok-free.dev/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"password123"}'

# Atau test OPTIONS preflight
curl -X OPTIONS https://your-ngrok-url.ngrok-free.dev/api/auth/register \
  -H "Origin: https://your-frontend.com" \
  -H "Access-Control-Request-Method: POST" \
  -v
```

**Debug Endpoint** (untuk check routes):

```bash
GET /api/debug/routes  # List semua endpoints
POST /api/debug/test-post  # Test POST endpoint
```

**📝 Demo Users juga dibuat otomatis**:

- `demo1@musicapp.local` / `demo123`
- `demo2@musicapp.local` / `demo123`

### 7. Untuk Development dengan React (Network Access)

Jika ingin akses dari HP/device lain di jaringan lokal:

**Backend (sudah siap)**:

- Server bind ke `0.0.0.0:8080`
- CORS allow semua origin di development mode

**Frontend React**:

```bash
# Untuk Vite
npm run dev -- --host 0.0.0.0

# Atau Create React App
npm start -- --host 0.0.0.0
```

**API Base URL di Frontend**:
Gunakan IP komputer (bukan localhost):

```javascript
const API_BASE = "http://192.168.1.15:8080/api"; // Ganti dengan IP komputer Anda
```

- Pastikan `ENV=development` di .env
- Gunakan port 3000, 3001, atau 5173 untuk React app

## Environment Variables

| Variable        | Development Default | Production                       | Description                                   |
| --------------- | ------------------- | -------------------------------- | --------------------------------------------- |
| ENV             | development         | production                       | Environment mode                              |
| DB_HOST         | localhost           | (from Railway)                   | Database host                                 |
| DB_PORT         | 5432                | 5432                             | Database port                                 |
| DB_USER         | postgres            | (from Railway)                   | Database user                                 |
| DB_PASSWORD     | password            | (from Railway)                   | Database password                             |
| DB_NAME         | music_app           | (from Railway)                   | Database name                                 |
| DB_SSLMODE      | disable             | require                          | SSL mode                                      |
| JWT_SECRET      | default-jwt-secret  | (set in Railway)                 | JWT signing secret                            |
| YOUTUBE_API_KEY | -                   | (optional)                       | YouTube Data API v3 key untuk pencarian audio |
| CORS_ORIGIN     | -                   | https://your-frontend.vercel.app | Frontend URL for CORS                         |

## 🌱 Database Seeding

### Development Mode (Otomatis)

Saat app pertama kali dijalankan dengan `ENV=development`:

- Migration otomatis dijalankan (membuat tables)
- Seeding otomatis dijalankan (membuat admin user & demo users)
- **Jika sudah ada admin user, seeding SKIP** (safety check)

### Production Mode (Manual)

Untuk production, Anda punya 2 opsi:

**Opsi 1: Manual SQL Seed**

```bash
# Restore dari backup
psql -U username -d database_name < database-backup.sql
```

**Opsi 2: Programmatic (dari Go app)**
Di production, seeding tidak otomatis. Tapi bisa dipanggil manual via endpoint custom atau CLI command:

```go
// Contoh: dibuat endpoint admin untuk seeding
database.SeedInitialData()
```

## API Endpoints

- `POST /api/auth/register` - Registrasi user
- `POST /api/auth/login` - Login user
- `GET /api/auth/me` - Get current user (protected)
- `GET /api/songs` - Get all songs
- `GET /api/songs/search` - Search songs
- `GET /api/songs/:id` - Get song by ID
- Dan lainnya...

## Deployment

### Railway/Vercel Deployment

1. Set environment variables di Railway:

```ini
ENV=production
PORT=3000
DB_HOST=<railway-db-host>
DB_PORT=5432
DB_USER=<railway-user>
DB_PASSWORD=<railway-password>
DB_NAME=<railway-db-name>
DB_SSLMODE=require
JWT_SECRET=your-secret-here
CORS_ORIGIN=https://your-frontend.vercel.app
```

2. Untuk Ngrok Testing di Production:

```ini
ENV=production
# Biarkan CORS_ORIGIN kosong = auto-detect ngrok domains
```

### Ngrok Setup untuk Testing Production

```bash
# Start ngrok tunnel
ngrok http 8080

# Copy https://xxxxx-free.ngrok-free.dev
# Gunakan URL ini untuk testing

# Frontend harus hit endpoint ini:
# https://xxxxx-free.ngrok-free.dev/api/auth/login
```

**⚠️ PENTING**:

- Request harus include `Content-Type: application/json` header
- CORS preflight (OPTIONS) harus berhasil sebelum POST
- Jika masih 404, cek log dengan `/api/debug/routes`
