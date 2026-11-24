# Backend Implementation Plan — PT Muda Kreasi Teknologi

Dokumen ini adalah blueprint perencanaan backend untuk menangani:
1. **Admin Login** (Autentikasi & Manajemen Sesi Admin)
2. **CRUD Produk yang Dijual**
3. **CRUD Cashflow** (`PROFIT` / Untung, `LOSS` / Rugi, `SAVING` / Tabungan)
4. **Infotainment Data Logger Arduino & HMI Web System** (Monitoring real-time alat pabrik: Silo, Motor/Conveyor, Drum dengan Induction Coil)

> **Referensi Infrastruktur:** Seluruh fondasi arsitektur, database, autentikasi, HTTP envelope, logging, dan deployment mengacu pada pola yang sudah berjalan di `D:\Workdir\Go\kicau-mania\cloud-app`.

---

## 1. Tech Stack & Infrastruktur (Mengacu pada `cloud-app`)

| Layer / Kebutuhan | Stack / Pola Referensi (`cloud-app`) |
| :--- | :--- |
| **Bahasa** | Go (`go 1.25+`) |
| **HTTP Framework** | `github.com/gin-gonic/gin` |
| **Database** | PostgreSQL + `github.com/jackc/pgx/v5/pgxpool` |
| **SQL Management** | Embedded SQL (`//go:embed *.sql` di `db/queries/queries.go` dengan penanda `-- name: ...`) |
| **Migration** | File SQL terpisah `db/migrations/001_*.up.sql` & `.down.sql` (dijalankan eksplisit via script `migrate.sh`, tanpa auto-migrate saat app start) |
| **Password & Auth** | Argon2id (`golang.org/x/crypto/argon2` dengan bounded concurrency) + Opaque Bearer Access Token & HttpOnly Refresh Cookie (`admin_sessions`) + In-memory Rate Limiter |
| **IoT Device Auth** | API Key Header (`X-Device-Key: <secret>`) dicocokkan dengan hash di tabel `devices` |
| **Realtime Telemetry** | Server-Sent Events (SSE) via Gin `c.Stream(...)` + Go in-memory channel Hub |
| **Precision Math (Uang/Sensor)** | `NUMERIC(15,2)` di PostgreSQL (diparse aman sebagai string/decimal/float64 terkontrol di Go) |
| **Structured Logger** | `go.uber.org/zap` + redaction field sensitif (`password`, `token`, `authorization`, `api_key`) |
| **API Docs** | `github.com/swaggo/swag` + `gin-swagger` (`go generate .` -> `/api/docs`) |
| **Config** | `github.com/joho/godotenv` + env override (`config/config.go`) |
| **Deployment / Infra** | Docker + Nginx reverse proxy (`web_gateway`) + Certbot SSL + scripts operasional (`migrate.sh`, `backup-db.sh`, `deploy.sh`, `bootstrap-admin.sh`) |

---

## 2. Struktur Direktori (Clean Architecture Modular Monolith)

```text
pt-muda-kreasi-teknologi/
├── cmd/
│   ├── api/
│   │   └── main.go                         # Entrypoint HTTP server, load .env/config, graceful shutdown
│   └── bootstrap-admin/
│       └── main.go                         # CLI pembuatan akun admin awal via env (tanpa password default)
├── config/
│   └── config.go                           # Loader env (DATABASE_URL, HTTP_ADDR, ALLOWED_ORIGINS, TRUSTED_PROXIES, dll)
├── db/
│   ├── migrations/
│   │   ├── 001_admins_and_sessions.up.sql
│   │   ├── 001_admins_and_sessions.down.sql
│   │   ├── 002_products.up.sql
│   │   ├── 002_products.down.sql
│   │   ├── 003_cashflows.up.sql
│   │   ├── 003_cashflows.down.sql
│   │   ├── 004_telemetry_and_devices.up.sql
│   │   └── 004_telemetry_and_devices.down.sql
│   └── queries/
│       ├── queries.go                      # //go:embed *.sql parser (-- name: ...)
│       ├── auth.sql                        # Query admin & session
│       ├── product.sql                     # Query CRUD produk & cursor pagination
│       ├── cashflow.sql                    # Query CRUD cashflow & summary aggregation
│       └── telemetry.sql                   # Ingest telemetry, get latest, history, & stats
├── deployment/                             # Referensi dari cloud-app/deployment
│   └── opt/
│       ├── infra/
│       │   ├── docker-compose.yml          # Nginx reverse proxy + Certbot
│       │   └── nginx/conf.d/               # Proxy pass ke 127.0.0.1:<HTTP_PORT>
│       └── scripts/
│           ├── app/                        # deploy.sh, bootstrap-admin.sh
│           └── db/                         # initialize-db.sh, migrate.sh, backup-db.sh, restore-db.sh
├── docs/
│   └── generated/                          # Output otomatis `go generate .` (Swaggo)
├── internal/
│   ├── app/
│   │   └── container.go                    # Explicit DI Container (Pool -> Store -> Repo -> UseCase -> Handlers)
│   ├── apperror/
│   │   └── error.go                        # ErrInvalid, ErrUnauthorized, ErrForbidden, ErrNotFound, ErrConflict, ErrRateLimited, ErrUnavailable
│   ├── auth/                               # Domain 1: Admin Login
│   │   ├── model/
│   │   │   └── auth.go                     # Login, Tokens, AdminIdentity, AdminProfile
│   │   ├── repository/
│   │   │   └── auth.go                     # Implementasi query auth.sql via postgres.Store
│   │   ├── service/
│   │   │   ├── password.go                 # Argon2id Hash & Verify (bounded concurrency)
│   │   │   └── limiter.go                  # Rate limiter login
│   │   └── usecase/
│   │       └── session.go                  # Login, Refresh, Logout, Me, AuthenticateAdmin
│   ├── product/                            # Domain 2: CRUD Produk yang Dijual
│   │   ├── model/
│   │   │   └── product.go                  # Product, CreateProduct, UpdateProduct, ListProductFilter
│   │   ├── repository/
│   │   │   └── product.go                  # Implementasi query product.sql
│   │   └── usecase/
│   │       └── product.go                  # Validasi SKU unik, harga >= 0, stok, status ACTIVE/INACTIVE/OUT_OF_STOCK
│   ├── cashflow/                           # Domain 3: CRUD Cashflow (Untung, Rugi, Tabungan)
│   │   ├── model/
│   │   │   └── cashflow.go                 # Cashflow, CreateCashflow, UpdateCashflow, CashflowSummary
│   │   ├── repository/
│   │   │   └── cashflow.go                 # Implementasi query cashflow.sql
│   │   └── usecase/
│   │       └── cashflow.go                 # Validasi type (PROFIT/LOSS/SAVING), amount > 0, cek product_id opsional
│   ├── telemetry/                          # Domain 4: Infotainment Data Logger Arduino & HMI
│   │   ├── model/
│   │   │   └── telemetry.go                # IngestPayload, Snapshot, HistoryFilter, DeviceModel
│   │   ├── repository/
│   │   │   └── telemetry.go                # IngestTelemetry, GetLatest, ListHistory, GetStats
│   │   ├── service/
│   │   │   └── hub.go                      # In-memory Go channel broadcast Hub untuk real-time SSE
│   │   ├── usecase/
│   │   │   └── telemetry.go                # Validasi sensor, threshold alarm, dispatch SSE event
│   │   └── delivery/
│   │       └── rest/
│   │           ├── handler.go              # Ingest POST, Stream SSE GET, Latest, History, Stats
│   │           └── requests.go             # Request binding structs
│   ├── health/
│   │   └── usecase/                        # Health & Readiness check (/healthz, /readyz)
│   ├── infra/
│   │   ├── logger/                         # Zap logger + redact.go
│   │   └── postgres/
│   │       ├── pool.go                     # pgxpool.NewWithConfig
│   │       └── store.go                    # Store.Do (context txKey), klasifikasi PgError
│   ├── delivery/
│   │   ├── router.go                       # Root Gin engine, global middleware, NoRoute/NoMethod, route groups
│   │   ├── middleware/
│   │   │   ├── auth.go                     # RequireAdmin (Bearer token)
│   │   │   ├── device_auth.go              # RequireDeviceKey (X-Device-Key untuk Arduino)
│   │   │   ├── headers.go                  # SecurityHeaders, RequestID
│   │   │   └── http.go                     # AccessLog, Recover, CanonicalParams, Limits, CORS
│   │   └── rest/
│   │       ├── health.go                   # Live & Ready handlers
│   │       ├── apidocs/                    # Swagger UI (/api/docs)
│   │       ├── auth/                       # handler.go & requests.go
│   │       ├── product/                    # handler.go & requests.go
│   │       ├── cashflow/                   # handler.go & requests.go
│   │       └── response/
│   │           └── response.go             # Envelope JSON & CursorPage[T]
│   └── server/
│       └── server.go                       # HTTP server runner & graceful shutdown
├── generate.go                             # //go:generate swag init
├── .env.example
├── go.mod
└── README.md
```

---

## 3. Desain Database (PostgreSQL)

### 3.1 Migration `001_admins_and_sessions.up.sql`
```sql
CREATE TABLE admins (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(120) NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'ADMIN',       -- ADMIN | SUPER_ADMIN
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_admins_username UNIQUE (username),
    CONSTRAINT uq_admins_email UNIQUE (email)
);

CREATE TABLE admin_sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    admin_id BIGINT NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    access_token_hash TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    access_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_admin_sessions_access UNIQUE (access_token_hash),
    CONSTRAINT uq_admin_sessions_refresh UNIQUE (refresh_token_hash)
);

CREATE INDEX idx_admin_sessions_admin_id ON admin_sessions (admin_id) WHERE revoked_at IS NULL;
```

### 3.2 Migration `002_products.up.sql`
```sql
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sku VARCHAR(64) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    category VARCHAR(80) NOT NULL,
    cost_price NUMERIC(15,2) NOT NULL DEFAULT 0.00 CHECK (cost_price >= 0),
    selling_price NUMERIC(15,2) NOT NULL DEFAULT 0.00 CHECK (selling_price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    unit VARCHAR(30) NOT NULL DEFAULT 'PCS',         -- PCS | BOX | KG | UNIT
    image_url VARCHAR(500) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'OUT_OF_STOCK')),
    created_by BIGINT NULL REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX uq_products_sku_active ON products (sku) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_status_id ON products (status, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products (category) WHERE deleted_at IS NULL;
```

### 3.3 Migration `003_cashflows.up.sql`
```sql
CREATE TABLE cashflows (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('PROFIT', 'LOSS', 'SAVING')), -- Untung | Rugi | Tabungan
    category VARCHAR(80) NOT NULL,                                           -- PENJUALAN, OPERASIONAL, KERUSAKAN, TABUNGAN_USAHA, dll
    title VARCHAR(150) NOT NULL,
    description TEXT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),                        -- Selalu positif; arah ditentukan oleh type
    product_id BIGINT NULL REFERENCES products(id) ON DELETE SET NULL,       -- Opsional jika terkait produk
    transaction_date DATE NOT NULL,
    created_by BIGINT NULL REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_cashflows_type_date ON cashflows (type, transaction_date DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_cashflows_date_id ON cashflows (transaction_date DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_cashflows_product_id ON cashflows (product_id) WHERE deleted_at IS NULL;
```

### 3.4 Migration `004_telemetry_and_devices.up.sql`
```sql
CREATE TABLE devices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    api_key_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_heartbeat_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_devices_code UNIQUE (device_code)
);

CREATE TABLE telemetry_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    silo_weight NUMERIC(10,2) NOT NULL DEFAULT 0.00,       -- Berat di Silo (kg)
    motor_rpm INTEGER NOT NULL DEFAULT 0,                  -- Kecepatan putaran motor conveyor (RPM)
    drum_temperature NUMERIC(5,2) NOT NULL DEFAULT 0.00,   -- Suhu di Drum (°C)
    drum_humidity NUMERIC(5,2) NOT NULL DEFAULT 0.00,      -- Kelembapan di Drum (%)
    coil_status VARCHAR(20) NOT NULL DEFAULT 'STANDBY',    -- ACTIVE | STANDBY | OFF | ERROR
    motor_status VARCHAR(20) NOT NULL DEFAULT 'STOPPED',   -- RUNNING | STOPPED | WARNING
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telemetry_device_time ON telemetry_logs (device_id, recorded_at DESC, id DESC);
CREATE INDEX idx_telemetry_recorded_at ON telemetry_logs (recorded_at DESC);
```

---

## 4. Rincian Endpoint Seluruh Modul

### 4.1 Modul 1: Admin Login (`/api/v1/admin/auth`)

| Method | Endpoint | Middleware | Deskripsi |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/admin/auth/login` | Public (+ Rate Limiter) | Login `identifier`, `password`, `device_id` |
| `POST` | `/api/v1/admin/auth/refresh` | Public (Refresh Cookie) | Rotasi token sesi aktif |
| `POST` | `/api/v1/admin/auth/logout` | `RequireAdmin` | Revoke session aktif |
| `GET` | `/api/v1/admin/auth/me` | `RequireAdmin` | Ambil profil admin aktif |

### 4.2 Modul 2: CRUD Produk yang Dijual (`/api/v1/admin/products`)

| Method | Endpoint | Middleware | Deskripsi |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/admin/products` | `RequireAdmin` | Tambah produk baru |
| `GET` | `/api/v1/admin/products` | `RequireAdmin` | List produk (cursor pagination + filter) |
| `GET` | `/api/v1/admin/products/:id` | `RequireAdmin` | Detail produk by ID |
| `PUT` | `/api/v1/admin/products/:id` | `RequireAdmin` | Update data produk |
| `PATCH` | `/api/v1/admin/products/:id/stock` | `RequireAdmin` | Penyesuaian stok cepat |
| `DELETE` | `/api/v1/admin/products/:id` | `RequireAdmin` | Soft delete produk |

### 4.3 Modul 3: CRUD Cashflow — Untung, Rugi, Tabungan (`/api/v1/admin/cashflows`)

| Method | Endpoint | Middleware | Deskripsi |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/admin/cashflows` | `RequireAdmin` | Catat entri cashflow (`PROFIT`, `LOSS`, `SAVING`) |
| `GET` | `/api/v1/admin/cashflows` | `RequireAdmin` | List cashflow (cursor pagination + filter) |
| `GET` | `/api/v1/admin/cashflows/summary` | `RequireAdmin` | Rekap total untung, rugi, tabungan, & net profit |
| `GET` | `/api/v1/admin/cashflows/:id` | `RequireAdmin` | Detail entri cashflow by ID |
| `PUT` | `/api/v1/admin/cashflows/:id` | `RequireAdmin` | Update entri cashflow |
| `DELETE` | `/api/v1/admin/cashflows/:id` | `RequireAdmin` | Soft delete entri cashflow |

### 4.4 Modul 4: Infotainment Data Logger Arduino & HMI (`/api/v1/telemetry` & `/hmi`)

| Method | Endpoint | Middleware | Deskripsi |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/telemetry/ingest` | `RequireDeviceKey` | Ingest data logger dari Arduino (berat, rpm, suhu, kelembapan) |
| `GET` | `/api/v1/telemetry/stream` | Public / HMI | Realtime streaming SSE untuk HMI Dashboard |
| `GET` | `/api/v1/telemetry/latest` | Public / HMI | Snapshot pembacaan sensor terbaru |
| `GET` | `/api/v1/telemetry/history` | `RequireAdmin` | Log data sensor historis (cursor pagination) |
| `GET` | `/api/v1/telemetry/stats` | `RequireAdmin` | Statistik agregasi untuk chart histori |
| `GET` | `/hmi` | Public / Operator | Halaman Web HMI Infotainment interaktif (M-IoT style) |

---

## 5. Konsep Visual HMI Infotainment (M-IoT Industrial Style)

Antarmuka web HMI (`/hmi`) mengadopsi tampilan industrial monitor seperti Gambar 1 dengan menyajikan 3 unit kerja dari Gambar 2:
1. **Silo Unit**: Silo berbentuk kerucut/tabung dengan visualisasi level bahan di dalamnya dan display sensor berat (**kg**).
2. **Conveyor & Motor Unit**: Konveyor miring penghubung dari Silo menuju drum pengering, dilengkapi display putaran motor (**RPM**) dan status motor.
3. **Rotary Drum with Induction Coil Unit**: Tabung drum horisontal dengan lilitan pemanas induksi luar yang menyala membara saat `coil_status = ACTIVE`, dilengkapi dial gauge suhu drum (**°C**) dan kelembapan produk (**%**).
4. **Live Trend Chart**: Line chart real-time di sudut atas dashboard menampilkan grafik tren suhu, kelembapan, dan rpm yang terus ter-update melalui Server-Sent Events (SSE).
5. **Alarm Preset**: Indikator batas aman suhu dan berat dengan status visual alarm (NORMAL / WARNING / CRITICAL).

---

## 6. Tahapan Eksekusi

| Fase | Ruang Lingkup | Output Utama |
| :--- | :--- | :--- |
| **Fase 1: Fondasi Infrastruktur** | Setup `go.mod`, `config/`, `apperror`, logger Zap, koneksi pool PostgreSQL, `store.go`, middleware, response envelope | Skeleton server, `/healthz`, `/readyz`, `/api/docs` |
| **Fase 2: Skema Database & Migrasi** | Migrasi `001..004` SQL + queries terstruktur `-- name: ...` + `cmd/bootstrap-admin` | Database schema siap di PostgreSQL |
| **Fase 3: Modul Admin Auth** | Password Argon2id, session token, limiter, usecase & handler auth | Login, refresh, logout, me |
| **Fase 4: Modul Produk & Cashflow** | CRUD Produk (cursor pagination, filter) + CRUD Cashflow & Summary (`PROFIT`, `LOSS`, `SAVING`) | API Produk & Cashflow lengkap |
| **Fase 5: Modul Telemetry Arduino & SSE Hub** | Ingestion endpoint `POST /api/v1/telemetry/ingest`, in-memory Go SSE Hub `GET /stream`, API latest, history, stats | Data logger Arduino real-time aktif |
| **Fase 6: HMI Web Dashboard** | Web HMI interaktif (HTML/CSS/JS embedded) dengan skematik Silo -> Motor/Conveyor -> Induction Drum, dial gauge, dan live chart | Layar infotainment HMI siap pakai di `/hmi` |
| **Fase 7: Verifikasi & Deployment** | Unit test, integration test, simulasi cURL Arduino, Docker & Nginx proxy configuration | Sistem teruji dan siap deploy |
