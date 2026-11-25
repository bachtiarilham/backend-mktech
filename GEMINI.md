# PT Muda Kreasi Teknologi (Backend) - AI Agent & Developer Guidelines

Dokumen ini adalah aturan utama (*Project Rules*) yang **wajib dipatuhi secara otomatis** oleh AI Agent dan developer saat bekerja di repositori backend Go `pt-muda-kreasi-teknologi`.

---

## 1. Arsitektur & Desain Kode (`go-clean-architecture`)
Backend ini **lebih memprioritaskan pendekatan Vertical Slice (Domain-First Modular Monolith)** berbasis Clean Architecture dengan referensi infrastruktur `cloud-app`:

- **Vertical Slice & Isolasi Domain (`internal/<domain>/`):**
  Setiap fitur bisnis diorganisasikan secara vertikal dan mandiri sehingga perubahan satu fitur terlokalisasi dalam satu *slice*:
  - `auth` (Admin Login, Argon2id, opaque session & refresh cookie)
  - `product` (CRUD Produk yang dijual)
  - `cashflow` (CRUD Cashflow: `PROFIT`, `LOSS`, `SAVING` & summary)
  - `telemetry` (Arduino data logger ingestion via `X-Device-Key` & SSE Broadcast Hub)
- **Aturan Layer:**
  - **Delivery (`internal/<domain>/delivery/`):** Hanya parsing request, validasi sintaks JSON/query, ekstraksi identity dari context (`middleware.Admin`, `middleware.Device`), pemanggilan UseCase, dan penulisan response via `response.Success` / `response.Created` / `response.WriteError`.
  - **UseCase (`internal/<domain>/usecase/`):** Memegang seluruh logika bisnis murni dan koordinasi transaksi (`store.Do(ctx, ...)`). Dilarang mengimpor `net/http` atau `github.com/gin-gonic/gin`.
  - **Repository (`internal/<domain>/repository/`):** Hanya mengeksekusi named SQL dari `db/queries/*.sql` via `postgres.Store`. Repository **dilarang** memanggil repository lain.
- **Explicit Dependency Injection:** Seluruh wiring dilakukan secara eksplisit di `internal/app/container.go` tanpa reflection magic.
- **SQL Source Tunggal (`db/queries/*.sql`):** Seluruh query ditulis dalam file `.sql` dengan penanda `-- name: <QueryName>` dan di-embed via `//go:embed *.sql`. Dilarang menduplikasi query SQL bisnis di kode Go.
- **Migrasi Manual (`db/migrations/`):** Menggunakan pasangan `.up.sql` dan `.down.sql`. Dilarang menjalankan auto-migrate saat startup aplikasi.
- **Pagination Universal:** Seluruh endpoint list koleksi wajib menggunakan **cursor-based pagination** (`response.CursorPage[T]`). Dilarang menggunakan `page` dan `offset`.
- **Penanganan Error & Keamanan:**
  - Gunakan sentinel errors di `internal/apperror` (`ErrInvalid`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrConflict`, `ErrUnavailable`, `ErrRateLimited`) dan `apperror.Public(...)` untuk pesan aman.
  - Dilarang membocorkan detail error SQL/driver ke response HTTP.
  - Logger Zap (`internal/infra/logger`) wajib menyensor field sensitif (`password`, `token`, `authorization`, `api_key`, `device_key`).
- **Integritas Data & Larangan Emoji:** Status dan enum backend wajib berupa string uppercase bersih (`ACTIVE`, `INACTIVE`, `OUT_OF_STOCK`, `PROFIT`, `LOSS`, `SAVING`, `STANDBY`, `RUNNING`, `STOPPED`). Dilarang menyisipkan karakter emoji ke data backend.

---

## 2. Standar Zero-Comment & Executable Constraints (`no-code-comments`)
Dilarang mengandalkan komentar pasif agar developer atau AI mengingat suatu aturan. Jadikan aturan dapat dieksekusi oleh mesin:
1. **Constraint -> API & Type System:** Gunakan `NUMERIC(15,2)` untuk uang/sensor, tag validasi deklaratif (`binding:"required"`, `binding:"gte=0"`), serta konstanta/enum tervalidasi di UseCase.
2. **Konvensi Tim -> Linter & Test:** Tegakkan batas antar-layer melalui unit test dan compiler checks.
3. **Business Rule -> Unit Test:** Buktikan aturan bisnis (validasi SKU unik, kalkulasi `net_profit`, threshold alarm telemetri, Argon2id verification) menggunakan table-driven unit tests.
4. **Pengecualian yang Diizinkan:** Hanya compiler directives (`//go:embed`, `//go:generate`, `//go:build`), anotasi Swagger/OpenAPI (`// @Summary`, `// @Router`), dan 1 baris `// Package <name> ...` di awal file.

---

## 3. Protokol Eliminasi Workaround (`eliminate-workarounds`)
- **Single Source of Truth & Git as Archive:**
  - Dilarang mengomentari kode mati (`/* ... */` atau `// func Old()`); hapus langsung karena riwayat tersimpan di Git.
  - Dilarang membiarkan implementasi deprecated atau duplikat hidup berdampingan dengan implementasi baru.
  - Dilarang membungkus (*wrap*) workaround lama; selesaikan langsung pada *root cause*.
  - Dilarang meninggalkan komentar `TODO`, `FIXME`, `HACK`, atau `time.Sleep(...)` arbitrari di kode produksi.
- **Boy Scout Rule:** Setiap menyentuh file, bersihkan dead code, unused imports, dan duplikasi yang ditemukan.

---

## 4. Pemisahan Frontend & Backend
- Repositori ini (`d:\Workdir\Go\pt-muda-kreasi-teknologi`) adalah **Headless REST API & SSE Backend** murni.
- Seluruh kode UI/HMI (termasuk dashboard monitoring Silo, Motor/Conveyor, dan Induction Drum) berada di repositori frontend SvelteKit: `D:\Workdir\sveltekit\muda-kreasi-teknologi`.
- Dilarang menaruh aset HTML/CSS/JS UI di backend ini.

---

## 5. Protokol Verifikasi Wajib
Sebelum menyatakan perubahan selesai, AI Agent wajib menjalankan verifikasi berikut dan melaporkan hasil nyatanya:

```powershell
# 1. Jalankan seluruh unit & integration test
go test -v ./...

# 2. Pastikan kedua binary utama terkompilasi bersih
go build ./cmd/api
go build ./cmd/bootstrap-admin
```
*(Hapus binary `.exe` hasil uji build lokal agar tidak mengotori working tree).*

---

## 6. Rujukan Spesifikasi Proyek
- Master Plan Backend & Skema DB: `BACKEND_PLAN.md`
- Referensi Infrastruktur Go: `D:\Workdir\Go\kicau-mania\cloud-app`
- Frontend SvelteKit Pair: `D:\Workdir\sveltekit\muda-kreasi-teknologi`
