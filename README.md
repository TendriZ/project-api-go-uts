# Sistem Manajemen Perpustakaan — RESTful API (Go + Fiber)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Framework](https://img.shields.io/badge/Framework-Fiber_v2-00AC47?style=flat&logo=fiber)](https://gofiber.io)
[![Database](https://img.shields.io/badge/Database-PostgreSQL_16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Architecture](https://img.shields.io/badge/Architecture-Strict_Clean_Architecture-orange?style=flat)](#arsitektur-sistem)
[![RBAC](https://img.shields.io/badge/RBAC-3--Tier_Authorization-blueviolet?style=flat)](#matriks-role-based-access-control-rbac)
[![Test Coverage](https://img.shields.io/badge/Tests-Passing-brightgreen?style=flat)](#pengujian--testing)

Proyek ini adalah implementasi backend RESTful API untuk **Sistem Manajemen Perpustakaan** yang dibangun menggunakan **Go (Golang)** dan framework **Fiber v2**, dengan basis data **PostgreSQL** melalui driver berkinerja tinggi **pgx/v5**.

Aplikasi ini dirancang dengan standar *production-ready* yang menerapkan **Strict Clean Architecture**, keamanan sesi berbasis **Dual JWT & Refresh Token Rotation**, kontrol hak akses berlapis (**3-Tier Role-Based Access Control / RBAC**), proteksi transaksi konkuren (**ACID Transaction + Pessimistic Row Locking**), serta penanganan data tingkat lanjut (**Keyset/Cursor-based Pagination** dan **Content Negotiation CSV/JSON**).

---

## Identitas Pengembang

- **Nama Mahasiswa:** Muhammad Raka Razzani
- **NIM:** 434241056
- **Mata Kuliah:** Praktikum Pemrograman Backend Lanjut (Semester 5)
- **Tugas:** Laporan Proyek Ujian Tengah Semester (UTS)
- **Repositori GitHub:** [TendriZ/project-api-go-uts](https://github.com/TendriZ/project-api-go-uts)

---

## Daftar Isi

1. [Fitur Unggulan](#fitur-unggulan)
2. [Teknologi yang Digunakan](#teknologi-yang-digunakan)
3. [Arsitektur Sistem (Clean Architecture)](#arsitektur-sistem)
4. [Skema Basis Data & Migrasi](#skema-basis-data--migrasi)
5. [Matriks Role-Based Access Control (RBAC)](#matriks-role-based-access-control-rbac)
6. [Daftar Endpoint API](#daftar-endpoint-api)
7. [Panduan Instalasi & Menjalankan Proyek](#panduan-instalasi--menjalankan-proyek)
8. [Standar Keamanan & Reliabilitas](#standar-keamanan--reliabilitas)
9. [Pengujian & Testing](#pengujian--testing)

---

## Fitur Unggulan

- **Strict Clean Architecture:** Pemisahan kode yang tegas antara lapisan *Transport/Handler*, *Service/Use Case*, *Repository/Data Access*, dan *Domain Model*, dengan *dependency injection* dan *unidirectional data flow*.
- **Keamanan Otentikasi Berlapis (Modul 5):**
  - Penyimpanan kata sandi terenkripsi menggunakan **Bcrypt (Cost 12)**.
  - **Dual Token JWT (HMAC-SHA256):** Access Token (masa aktif 15 menit) dan Refresh Token (masa aktif 7 hari).
  - **Cryptographic Refresh Token Rotation:** Setiap penyegaran token menerbitkan token baru dan membatalkan (*revoke*) token lama; hash SHA-256 disimpan di basis data untuk audit dan pencegahan *replay attack*.
  - **Rate Limiting:** Proteksi *brute-force* pada endpoint login (maksimal 5 request per menit per alamat IP).
  - **Timing Attack Safe:** Pengecekan dummy hash untuk mencegah enumerasi akun pengguna melalui analisis waktu eksekusi.
- **Role-Based Access Control / RBAC 3-Tier (Modul 6):**
  - Tiga tingkatan peran: `admin`, `staff`, dan `user`.
  - Caching izin statis (*in-memory*) saat *booting* untuk performa kueri O(1) tanpa membebani basis data di setiap request.
  - *Fine-grained ownership check*: Pengguna biasa hanya dapat mengakses dan melihat riwayat peminjamannya sendiri, sedangkan staf dan admin dapat mengakses seluruh rekaman.
- **Konsistensi Transaksi Sirkulasi ACID (Modul 3):**
  - Logika peminjaman buku (*insert loan* + *decrement stock*) dan pengembalian (*update loan* + *increment stock*) berjalan dalam satu transaksi atomik `pgx.Tx`.
  - Proteksi *race condition* menggunakan *pessimistic locking* (`SELECT ... FOR UPDATE`) untuk mencegah peminjaman saat stok bernilai 0.
- **Pagination & Content Negotiation Modern (Modul 7):**
  - **Keyset / Cursor-based Pagination:** Navigasi katalog buku menggunakan composite pointer `(title, id) > ($1, $2)` dengan dukungan composite index B-Tree, menjamin kompleksitas $O(1)$ dan stabil pada jutaan baris data.
  - **Content Negotiation:** Mendukung ekspor data katalog buku ke format **CSV** melalui header `Accept: text/csv`, serta penolakan eksplisit berstandar RFC 7807 (**406 Not Acceptable**) jika klien meminta format yang tidak didukung (seperti XML).
- **Error Handling & Observabilitas Terpadu (Modul 7):**
  - Envelope respon JSON yang konsisten (`success`, `message`, `data`, `meta`).
  - Pemetaan error terpusat (`AppError`) yang menghasilkan UUID `request_id` unik untuk penelusuran log audit.

---

## Teknologi yang Digunakan

| Komponen | Pilihan Teknologi | Deskripsi / Alasan Pemilihan |
|---|---|---|
| **Bahasa Pemrograman** | Go (Golang) 1.22+ | Menawarkan konkurensi efisien (*goroutines*), *type-safety*, dan *footprint* memori minimal. |
| **HTTP Framework** | Fiber v2 (`gofiber/fiber/v2`) | Framework berbasis fasthttp dengan performa *throughput* tinggi dan arsitektur routing modular. |
| **Basis Data** | PostgreSQL 16 | RDBMS relasional tangguh dengan jaminan ACID penuh, konstrain FK, dan indeks komposit. |
| **Database Driver** | `jackc/pgx/v5` & `pgxpool` | Driver native PostgreSQL dengan pooling koneksi otomatis dan kueri terparameterisasi aman. |
| **Token Sesi** | `golang-jwt/jwt/v5` | Standar industri untuk pembuatan dan validasi token stateless JWT HMAC-SHA256. |
| **Password Hashing** | `golang.org/x/crypto/bcrypt` | Standar keamanan resisten *brute-force* dengan faktor komputasi Cost 12. |
| **Validasi Deklaratif**| `go-playground/validator/v10` | Validasi struct deklaratif berbasis tag dengan dukungan kustom `nospace`. |
| **UUID Generator** | `google/uuid` | Pembuatan pelacak unik `request_id` untuk setiap transaksi HTTP. |

---

## Arsitektur Sistem

Aplikasi ini menerapkan **Clean Architecture** yang disederhanakan secara pragmatis untuk ekosistem Go tanpa mengorbankan isolasi tanggung jawab (*Single Responsibility Principle*):

```
project-api-go-uts/
├── app/
│   ├── model/             # [Entities] Definisi struct data murni, DTO request/response, konstanta status
│   │   ├── auth.go
│   │   ├── book.go
│   │   ├── common.go
│   │   ├── loan.go
│   │   └── user.go
│   ├── repository/        # [Data Access Layer] Interaksi langsung ke PostgreSQL via pgxpool
│   │   ├── book_repository.go
│   │   ├── loan_errors.go
│   │   ├── loan_repository.go
│   │   ├── role_repository.go
│   │   ├── token_repository.go
│   │   └── user_repository.go
│   └── service/           # [Use Cases & Business Logic] Validasi domain, orkestrasi transaksi, aturan kepemilikan
│       ├── auth_service.go
│       ├── authz_rules.go
│       ├── book_rules.go
│       ├── book_service.go
│       ├── loan_rules.go
│       ├── loan_service.go
│       ├── service_rules_test.go
│       ├── user_rules.go
│       └── user_service.go
├── config/                # [Framework & Drivers] Pembacaan variabel lingkungan, logger slog, inisialisasi Fiber
│   ├── app.go
│   ├── env.go
│   └── logger.go
├── database/              # [Framework & Drivers] Inisialisasi pool koneksi basis data (pgxpool)
│   └── postgres.go
├── helper/                # [Utilities] Fungsi murni tanpa efek samping (JWT, password, response, cursor, errors)
│   ├── authz.go
│   ├── context.go
│   ├── cursor.go
│   ├── errors.go
│   ├── jwt.go
│   ├── negotiate.go
│   ├── request.go
│   ├── response.go
│   ├── security.go
│   └── validator.go
├── middleware/            # [Interface Adapters] Otentikasi JWT, pemeriksaan izin RBAC, rate limiter, request ID
│   ├── auth.go
│   ├── authz.go
│   └── middleware.go
├── migrations/            # Berkas skrip DDL SQL untuk pembuatan skema tabel dan indeks
│   ├── 001_create_users.sql
│   ├── 002_auth.sql
│   ├── 003_rbac.sql
│   ├── 004_books.sql
│   ├── 005_loans.sql
│   └── 006_cursor_index.sql
├── route/                 # Registrasi pemetaan endpoint HTTP ke controller/service
│   └── route.go
├── .env.example           # Contoh template variabel lingkungan
├── .gitignore             # Konfigurasi pengabaian berkas Git
├── go.mod                 # Modul dependensi Go
├── go.sum                 # Checksum verifikasi dependensi
├── main.go                # Entry-point aplikasi (dependency injection & server lifecycle)
└── README.md              # Dokumentasi teknis proyek
```

### Arah Aliran Ketergantungan (*Dependency Rule*)

```
[ HTTP Request ]
       │
       ▼
[ middleware ] ──► Validasi Token JWT & Izin RBAC
       │
       ▼
[ route ] ───────► Meneruskan Request Context
       │
       ▼
[ app/service ] ─► Business Rules, Validasi DTO, Enkripsi, Orkestrasi Transaksi
       │
       ▼
[ app/repository ] ──► Kueri Parameter Terlindungi (pgxpool)
       │
       ▼
[ PostgreSQL DB ]
```

---

## Skema Basis Data & Migrasi

Basis data terdiri dari 7 tabel relasional yang saling terintegrasi:

1. **`users`:** Menyimpan data pengguna perpustakaan (`id`, `username`, `email`, `password` hash bcrypt, `role`, `is_active`, `created_at`).
2. **`roles`:** Tabel referensi nama peran (`admin`, `staff`, `user`).
3. **`permissions`:** Tabel daftar hak akses granular (16 izin, contoh: `book:create`, `loan:return`, dll).
4. **`role_permissions`:** Tabel penghubung *many-to-many* antara peran dan izin dengan konstrain `ON DELETE CASCADE`.
5. **`refresh_tokens`:** Menyimpan hash token sesi (SHA-256) terhubung ke `users(id)` dengan konstrain `ON DELETE CASCADE`.
6. **`books`:** Katalog inventaris buku (`id`, `isbn` unik, `title`, `author`, `publisher`, `year`, `stock`, `category`).
7. **`loans`:** Rekaman transaksi sirkulasi (`id`, `book_id`, `user_id`, `loaned_by`, `loan_date`, `due_date`, `return_date`, `status`). Konstrain foreign key menggunakan **`RESTRICT / NO ACTION`** demi menjaga integritas data audit legal perpustakaan.

### Indeks Performa Penting
- `books_title_id_idx` (`title ASC, id ASC`): Composite index B-Tree khusus untuk optimasi **Keyset Cursor Pagination**.
- `loans_user_id_idx`, `loans_book_id_idx`, `loans_status_idx`: Index B-Tree untuk mempercepat kueri filter transaksi peminjaman.

---

## Matriks Role-Based Access Control (RBAC)

Sistem membedakan hak akses secara ketat berdasarkan matriks izin:

| Modul / Domain | Izin (Permission) | Admin | Staff | User | Keterangan |
|---|---|:---:|:---:|:---:|---|
| **Auth** | *(Public / Authenticated)* |  |  |  | Login, Register, Refresh, Logout, dan `/auth/me` |
| **Katalog Buku** | `book:list` |  |  |  | Melihat katalog (JSON / CSV) & keyset pagination |
| | `book:read:any` |  |  |  | Melihat detail buku spesifik berdasarkan ID |
| | `book:create` |  |  | ❌ | Menambahkan koleksi buku baru |
| | `book:update:any`|  |  | ❌ | Memperbarui data atau stok buku (PUT/PATCH) |
| | `book:delete` |  | ❌ | ❌ | Menghapus data buku dari katalog |
| **Sirkulasi Pinjam** | `loan:create` |  |  | ❌ | Memproses peminjaman buku (kurangi stok) |
| | `loan:return` |  |  | ❌ | Memproses pengembalian buku (tambah stok) |
| | `loan:list:any` |  |  | ❌ | Staf/Admin melihat semua; User hanya miliknya |
| | `loan:read:any` |  |  | ❌ | Staf/Admin akses semua; User hanya miliknya |
| | `loan:delete` |  | ❌ | ❌ | Menghapus arsip rekaman peminjaman |
| **Pengguna** | `user:list` |  |  | ❌ | Melihat daftar seluruh pengguna terdaftar |
| | `user:create` |  | ❌ | ❌ | Mendaftarkan staf/user baru oleh admin |
| | `user:read:any` |  |  | ❌ | Staf/Admin baca semua; User hanya profil sendiri |
| | `user:update:any`|  | ❌ | ❌ | Admin ubah user lain; User hanya profil sendiri |
| | `user:delete` |  | ❌ | ❌ | Menghapus akun pengguna (dilarang hapus diri sendiri) |
| | `role:assign` |  | ❌ | ❌ | Mengubah peran pengguna (dilarang ubah peran sendiri) |

---

## Daftar Endpoint API

Seluruh endpoint diawali dengan prefix path `/api/v1`:

### 1. Autentikasi & Sesi
| Method | Endpoint | Auth | Role / Izin | Deskripsi |
|---|---|---|---|---|
| `GET` | `/health` | Publik | Bebas | Health check & verifikasi koneksi pool basis data |
| `POST` | `/auth/register` | Publik | Bebas | Pendaftaran akun baru (default role: `user`) |
| `POST` | `/auth/login` | Publik | Rate Limit (5/m) | Login pengguna, menghasilkan access & refresh token |
| `POST` | `/auth/refresh` | Publik | Valid Refresh | Rotasi refresh token dan penerbitan access token baru |
| `POST` | `/auth/logout` | Publik | Valid Refresh | Membatalkan (*revoke*) sesi refresh token |
| `GET` | `/auth/me` | Bearer | Semua Role | Menampilkan data profil dan izin pengguna aktif |

### 2. Manajemen Buku (Katalog)
| Method | Endpoint | Auth | Role / Izin | Deskripsi |
|---|---|---|---|---|
| `GET` | `/books` | Bearer | `book:list` | Daftar buku (cursor pagination, filter, JSON/CSV) |
| `GET` | `/books/:id` | Bearer | `book:read:any` | Menampilkan informasi detail buku berdasarkan ID |
| `POST` | `/books` | Bearer | `book:create` | Menambah koleksi buku baru (Admin/Staff) |
| `PUT` | `/books/:id` | Bearer | `book:update:any` | Mengganti seluruh data buku (Admin/Staff) |
| `PATCH` | `/books/:id` | Bearer | `book:update:any` | Mengubah stok atau metadata parsial buku |
| `DELETE`| `/books/:id` | Bearer | `book:delete` | Menghapus buku (Admin saja; ditolak jika ada pinjaman) |

### 3. Sirkulasi Peminjaman (Loans)
| Method | Endpoint | Auth | Role / Izin | Deskripsi |
|---|---|---|---|---|
| `GET` | `/loans` | Bearer | Kepemilikan | Daftar pinjaman (Staf: semua; User: milik sendiri) |
| `GET` | `/loans/:id` | Bearer | Kepemilikan | Detail pinjaman (Staf: semua; User: milik sendiri) |
| `POST` | `/loans` | Bearer | `loan:create` | Transaksi peminjaman buku (Atomik potong stok) |
| `PATCH` | `/loans/:id/return`| Bearer | `loan:return` | Transaksi pengembalian buku (Atomik tambah stok) |
| `DELETE`| `/loans/:id` | Bearer | `loan:delete` | Menghapus riwayat transaksi pinjaman (Admin) |

### 4. Manajemen Pengguna (Users)
| Method | Endpoint | Auth | Role / Izin | Deskripsi |
|---|---|---|---|---|
| `GET` | `/users` | Bearer | `user:list` | Daftar seluruh pengguna (Admin & Staff) |
| `POST` | `/users` | Bearer | `user:create` | Membuat akun baru oleh Admin |
| `GET` | `/users/:id` | Bearer | Kepemilikan | Detail user (Admin/Staff: semua; User: sendiri) |
| `PUT` | `/users/:id` | Bearer | Kepemilikan | Mengganti profil user (Admin: semua; User: sendiri) |
| `PATCH` | `/users/:id` | Bearer | Kepemilikan | Memperbarui parsial data user |
| `PATCH` | `/users/:id/role`| Bearer | `role:assign` | Mengubah role user (Admin saja; dilarang ubah diri sendiri) |
| `DELETE`| `/users/:id` | Bearer | `user:delete` | Menghapus user (Admin saja; ditolak jika ada relasi data) |

---

## Panduan Instalasi & Menjalankan Proyek

### 1. Prasyarat Sistem
- **Go:** Versi 1.22 atau lebih baru ([Unduh Go](https://golang.org/dl/))
- **PostgreSQL:** Versi 14 atau lebih baru ([Unduh PostgreSQL](https://www.postgresql.org/download/))
- **Git CLI**

### 2. Kloning Repositori
```bash
git clone https://github.com/TendriZ/project-api-go-uts.git
cd project-api-go-uts
```

### 3. Konfigurasi Variabel Lingkungan (`.env`)
Salin berkas contoh konfigurasi `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```
Sesuaikan konfigurasi koneksi basis data dan secret JWT pada berkas `.env`:
```ini
APP_NAME=API Perpustakaan
APP_PORT=3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_postgres_anda
DB_NAME=perpustakaan
DB_SSLMODE=disable
DB_MAX_CONNS=10

# Minimal 32 karakter
JWT_SECRET=super_secret_jwt_key_perpustakaan_uts_2026_juara
JWT_ISSUER=api-perpustakaan
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7

ALLOWED_ORIGINS=http://localhost:5173
LOG_LEVEL=info
```

### 4. Eksekusi Migrasi Basis Data
Buat basis data baru bernama `perpustakaan` di PostgreSQL, lalu jalankan seluruh skrip migrasi secara berurutan:
```bash
# Membuat basis data (jika belum ada)
psql -U postgres -c "CREATE DATABASE perpustakaan;"

# Menjalankan migrasi DDL 001 s.d. 006
psql -U postgres -d perpustakaan -f migrations/001_create_users.sql
psql -U postgres -d perpustakaan -f migrations/002_auth.sql
psql -U postgres -d perpustakaan -f migrations/003_rbac.sql
psql -U postgres -d perpustakaan -f migrations/004_books.sql
psql -U postgres -d perpustakaan -f migrations/005_loans.sql
psql -U postgres -d perpustakaan -f migrations/006_cursor_index.sql
```

### 5. Seeding Data Pengguna & Katalog Awal
Jalankan skrip seeding akun bawaan:
```bash
psql -U postgres -d perpustakaan -f scripts/seed_users.sql
```
> **Akun Bawaan Hasil Seeding:**
> - **Admin:** `admin` / Password: `Perpus2026Juara`
> - **Staff:** `staff` / Password: `Perpus2026Juara`
> - **User:** `user1` / Password: `Perpus2026Juara`

### 6. Menjalankan Server API
Unduh dependensi modul dan jalankan server:
```bash
go mod tidy
go run .
```
Server akan aktif dan siap menerima request di:
`http://localhost:3000/api/v1`

---

## Standar Keamanan & Reliabilitas

1. **SQL Injection Prevention:** Seluruh operasi data menggunakan kueri SQL terparameterisasi bawaan `pgx/v5` (`$1`, `$2`, dst.). Tidak ada interpolasi string mentah (*string concatenation*) pada kueri basis data.
2. **ACID Transaction & Concurrency Control:** Transaksi pinjam dan kembali buku membungkus operasi pembaruan status dan stok dalam blok `tx.Begin(ctx)` dan `tx.Commit(ctx)`. Eksekusi `SELECT ... FOR UPDATE` mencegah anomali *race condition* ketika banyak permintaan peminjaman terjadi serentak.
3. **Database Relational Integrity:**
   - Tabel `refresh_tokens` dan `role_permissions` menggunakan `ON DELETE CASCADE` untuk otomatis memusnahkan rekaman usang saat entitas induk dihapus.
   - Tabel `loans` menggunakan `RESTRICT / NO ACTION` untuk mencegah penghapusan akun atau buku yang masih memiliki rekam jejak audit transaksi legal.
4. **Brute-Force & Flooding Defense:** Middleware *Rate Limiter* membatasi percobaan login maksimal 5 kali per menit per alamat IP, mengembalikan respon terstandarisasi **429 Too Many Requests**.
5. **Request Tracking & Logging:** Middleware mencatat `request_id` unik berbasis UUIDv4 pada header respon `X-Request-Id` dan log terstruktur `log/slog` untuk mempermudah investigasi kendala server tanpa membocorkan detail teknis ke klien.

---

## Pengujian & Testing

### 1. Menjalankan Unit Test Otomatis
Unit test mencakup pengujian menyeluruh terhadap seluruh aturan bisnis domain di lapisan *Service*:
```bash
go test -v -cover ./app/service/...
```
Hasil uji mencakup:
- Validasi rentang hari peminjaman (1 s.d. 30 hari).
- Pengecekan izin otorisasi (*own vs any*).
- Pencegahan admin menghapus atau menurunkan peran akunnya sendiri.
- Validasi aturan format patch dan mutasi entitas.

### 2. Pengujian Manual via Postman
Repositori ini menyediakan koleksi pengujian Postman yang mencakup seluruh skenario positif (*Happy Path*) dan skenario negatif (*Negative Tests*):
- Koleksi Postman: `postman_collection.json`
- Panduan Pengujian: `PANDUAN_TESTING_POSTMAN_DAN_SCREENSHOT.md`

Fitur pengujian otomatis pada koleksi Postman mencakup *auto-capture token* dan *variable assignment* untuk access token, refresh token, ID buku, dan ID pinjaman.

---

## Lisensi

Proyek ini dikembangkan untuk tujuan akademik dalam rangka Ujian Tengah Semester (UTS) Praktikum Pemrograman Backend Lanjut di Universitas Airlangga. Bebas digunakan dan dimodifikasi untuk tujuan pembelajaran dengan tetap mencantumkan atribusi pengembang.
