# SewaSini

Backend API untuk platform penyewaan ruangan (booking ruang meeting, coworking, event space, dan pembayaran online).

## Highlights
- Arsitektur modular: handler -> service -> repository -> database.
- REST API lengkap untuk user, ruangan, kategori, review, booking, pembayaran.
- JWT authentication + role-based authorization (admin/user).
- Integrasi payment gateway (Xendit) dengan callback verification token.
- Collection Postman siap pakai untuk end-to-end testing.

## Tech Stack
- Go (Echo Framework)
- PostgreSQL
- JWT (auth)
- Xendit (invoice/payment)
- Mailjet (email notification)

## Struktur Proyek
- app/handler: HTTP handlers per domain.
- app/sewasini: entry point aplikasi, middleware, validator.
- service: business logic.
- repository: query SQL dan akses data.
- models: request/response/domain model.
- database/sql: schema dan script SQL.
- docs/postman: collection + environment Postman.

## Dokumentasi
- API lengkap: docs/API_DOCUMENTATION.md

## Quick Start

### 1) Konfigurasi Environment
Copy file .env lalu sesuaikan value berikut:
- koneksi PostgreSQL
- JWT secret
- XENDIT_SECRET_KEY
- XENDIT_CALLBACK_TOKEN
- Mailjet key

Contoh callback token lokal:
- XENDIT_CALLBACK_TOKEN=sewasini-callback-token

### 2) Setup Database
Jalankan schema utama:
- database/sql/database.sql

Atau gunakan file SQL per tabel di folder:
- database/sql/users.sql
- database/sql/kategori.sql
- database/sql/ruangan.sql
- database/sql/bookings.sql
- database/sql/transactions.sql
- database/sql/reviews.sql

Optional DML seed tambahan:
- database/sql/kategori_dml.sql
- database/sql/reviews_dml.sql

### 3) Jalankan Server
Jalankan dari folder app/sewasini:

```bash
go run main.go
```

Health check:

```bash
curl http://localhost:8080/health
```

## Base URL dan Prefix
- Base URL: http://localhost:8080
- Prefix yang tersedia:
- /api/v1
- /api

## Alur Testing Disarankan (Postman)
Gunakan file:
- docs/postman/SewaSini-User.postman_collection.json
- docs/postman/SewaSini-local.postman_environment.json

Urutan test paling aman:
1. Register User
2. Send OTP
3. Verify OTP
4. Login User
5. List/Get Ruangan
6. Create Booking
7. Create Payment
8. Payment Callback
9. Get Payment Status
10. Get Booking Status
11. Create Review

Pastikan environment Postman ini terisi:
- baseUrl
- accessToken
- ruanganId
- bookingId
- paymentId
- externalId
- categoryId
- reviewId
- xenditCallbackToken

## Ringkasan Endpoint

### Auth dan User
- POST /api/v1/users/register
- POST /api/v1/users/login
- POST /api/v1/users/send-otp
- POST /api/v1/users/verify-otp

### Ruangan
- GET /api/v1/ruangan
- GET /api/v1/ruangan/search
- GET /api/v1/ruangan/filter
- GET /api/v1/ruangan/:id

Parameter search/filter yang didukung:
- search, q, nama, keyword
- kota, location, lokasi, city
- kategori, category, kategori_id
- min_harga, min_price
- max_harga, max_price
- kapasitas
- tanggal_ketersediaan (YYYY-MM-DD)
- page, limit

### Category
- GET /api/v1/categories
- GET /api/v1/categories/:id
- GET /api/v1/kategori
- POST /api/v1/categories (admin)
- PUT /api/v1/categories/:id (admin)
- DELETE /api/v1/categories/:id (admin)

### Booking (Bearer)
- POST /api/v1/bookings
- GET /api/v1/bookings
- GET /api/v1/bookings/:id
- GET /api/v1/bookings/:id/status
- PUT /api/v1/bookings/:id
- DELETE /api/v1/bookings/:id

Catatan booking availability:
- Sistem menghitung overlap waktu booking.
- Status yang dianggap mengunci slot: pending dan confirmed.
- Slot memperhitungkan stock_availability ruangan.

### Review (Bearer)
- POST /api/v1/reviews
- GET /api/v1/reviews
- GET /api/v1/reviews/ruangan/:id
- GET /api/v1/reviews/:id
- PUT /api/v1/reviews/:id
- DELETE /api/v1/reviews/:id

### Payment
- POST /api/v1/payments (Bearer)
- GET /api/v1/payments/:id (Bearer)
- GET /api/v1/payments/invoice/:id (Bearer)
- POST /api/v1/payments/callback

## Payment Callback (Penting)
Header callback harus menyertakan token yang sama dengan nilai di env server.

Contoh:

```bash
curl -X POST "http://localhost:8080/api/v1/payments/callback" \
	-H "Content-Type: application/json" \
	-H "x-callback-token: sewasini-callback-token" \
	-d '{
		"id": "xnd-test-001",
		"external_id": "booking-4-1775748138-480",
		"status": "PAID"
	}'
```

Jika token tidak cocok, API akan mengembalikan:
- message: invalid callback token

## Error Umum
- invalid request body
- room is not available for the selected date
- invalid callback token
- booking does not belong to the authenticated user

## Catatan Pengembangan
- Endpoint tersedia pada dua prefix untuk kompatibilitas: /api/v1 dan /api.
- Beberapa endpoint admin membutuhkan role admin di JWT user.
- Untuk perubahan field payload dan response, referensi utama ada di docs/API_DOCUMENTATION.md.