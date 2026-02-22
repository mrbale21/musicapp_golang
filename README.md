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

| Variable    | Development Default | Production                       | Description           |
| ----------- | ------------------- | -------------------------------- | --------------------- |
| ENV         | development         | production                       | Environment mode      |
| DB_HOST     | localhost           | (from Railway)                   | Database host         |
| DB_PORT     | 5432                | 5432                             | Database port         |
| DB_USER     | postgres            | (from Railway)                   | Database user         |
| DB_PASSWORD | password            | (from Railway)                   | Database password     |
| DB_NAME     | music_app           | (from Railway)                   | Database name         |
| DB_SSLMODE  | disable             | require                          | SSL mode              |
| JWT_SECRET  | default-jwt-secret  | (set in Railway)                 | JWT signing secret    || YOUTUBE_API_KEY | -                   | (optional)                        | YouTube Data API v3 key untuk pencarian audio || CORS_ORIGIN | -                   | https://your-frontend.vercel.app | Frontend URL for CORS |

## API Endpoints

- `POST /api/auth/register` - Registrasi user
- `POST /api/auth/login` - Login user
- `GET /api/auth/me` - Get current user (protected)
- `GET /api/songs` - Get all songs
- `GET /api/songs/search` - Search songs
- `GET /api/songs/:id` - Get song by ID
- Dan lainnya...

## Deployment

Untuk production, set environment variables di Railway/Supabase sesuai dengan database yang disediakan.
