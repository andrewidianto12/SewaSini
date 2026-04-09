# SewaSini API Documentation

## Base URL
- http://localhost:8080

## API Prefix
- /api/v1
- /api

## Authentication
Untuk endpoint protected gunakan header:
Authorization: Bearer <access_token>

## Core Endpoints

### Auth/User
- POST /api/v1/users/register
- POST /api/v1/users/login
- POST /api/v1/users/send-otp
- POST /api/v1/users/verify-otp

### Ruangan
- GET /api/v1/ruangan
- GET /api/v1/ruangan/search
- GET /api/v1/ruangan/filter
- GET /api/v1/ruangan/:id

### Category
- GET /api/v1/categories
- GET /api/v1/categories/:id
- GET /api/v1/kategori
- POST /api/v1/categories (Admin + Bearer)
- PUT /api/v1/categories/:id (Admin + Bearer)
- DELETE /api/v1/categories/:id (Admin + Bearer)

Create Category payload example:
{
  "nama_kategori": "Training Room",
  "deskripsi": "Ruangan untuk pelatihan dan workshop"
}

### Review (Bearer required)
- POST /api/v1/reviews
- GET /api/v1/reviews
- GET /api/v1/reviews/ruangan/:id
- GET /api/v1/reviews/:id
- PUT /api/v1/reviews/:id
- DELETE /api/v1/reviews/:id

Create Review payload example:
{
  "ruangan_id": "1",
  "booking_id": "10",
  "rating": 5,
  "komentar": "Ruangan sangat nyaman"
}

Supported query params (search/filter):
- search | q | nama | keyword
- kota | location | lokasi | city
- kategori | category | kategori_id
- min_harga | min_price
- max_harga | max_price
- kapasitas
- tanggal_ketersediaan (format: YYYY-MM-DD)
- page (default 1)
- limit (default 10, max 100)

### Booking (Bearer required)
- POST /api/v1/bookings
- GET /api/v1/bookings
- GET /api/v1/bookings/:id
- GET /api/v1/bookings/:id/status
- PUT /api/v1/bookings/:id
- DELETE /api/v1/bookings/:id

Create Booking payload example:
{
  "ruangan_id": "1",
  "tanggal_mulai": "2026-02-20T09:00:00Z",
  "tanggal_selesai": "2026-02-20T12:00:00Z",
  "jumlah_peserta": 10
}

Booking availability notes:
- Booking bentrok dicek pada rentang waktu overlap.
- Status booking yang dihitung: pending dan confirmed.
- Sistem memperhitungkan stock_availability ruangan. Jika jumlah booking overlap sudah mencapai stock, request ditolak.

### Payment
- POST /api/v1/payments (Bearer required)
- GET /api/v1/payments/:id (Bearer required)
- GET /api/v1/payments/invoice/:id (Bearer required)
- POST /api/v1/payments/callback

Callback headers:
- Content-Type: application/json
- x-callback-token: harus sama dengan XENDIT_CALLBACK_TOKEN pada server

Callback body minimal:
{
  "id": "xnd-test-001",
  "external_id": "booking-4-1775748138-480",
  "status": "PAID"
}

## Common Error Messages
- invalid request body
- room is not available for the selected date
- invalid callback token
- booking does not belong to the authenticated user

## Postman
Gunakan file berikut:
- docs/postman/SewaSini-User.postman_collection.json
- docs/postman/SewaSini-local.postman_environment.json

Pastikan environment value xenditCallbackToken sama dengan nilai XENDIT_CALLBACK_TOKEN pada file .env.
