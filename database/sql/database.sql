
-- 1. Tabel Users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    nama_lengkap VARCHAR(255) NOT NULL,
    ttl VARCHAR(255) NOT NULL,
    no_hp VARCHAR(20) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    otp_code VARCHAR(6),
    otp_expiry TIMESTAMP,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 2. Tabel Kategori
CREATE TABLE IF NOT EXISTS kategori (
    id SERIAL PRIMARY KEY,
    nama_kategori VARCHAR(100) UNIQUE NOT NULL,
    deskripsi TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Tabel Ruangan
CREATE TABLE IF NOT EXISTS ruangan (
    id SERIAL PRIMARY KEY,
    nama_ruangan VARCHAR(255) NOT NULL,
    kategori_id INT NOT NULL REFERENCES kategori(id),
    deskripsi TEXT,
    alamat TEXT NOT NULL,
    kota VARCHAR(100) NOT NULL,
    kapasitas INTEGER NOT NULL,
    harga_per_jam BIGINT NOT NULL DEFAULT 0,
    harga_per_hari BIGINT NOT NULL DEFAULT 0,
    stock_availability INTEGER NOT NULL DEFAULT 1,
    fasilitas JSONB DEFAULT '[]',
    gambar JSONB DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ruangan_kategori ON ruangan(kategori_id);
CREATE INDEX IF NOT EXISTS idx_ruangan_kota ON ruangan(kota);
CREATE INDEX IF NOT EXISTS idx_ruangan_active ON ruangan(is_active);

-- 4. Tabel Bookings
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    ruangan_id INT NOT NULL REFERENCES ruangan(id),
    tanggal_mulai TIMESTAMP NOT NULL,
    tanggal_selesai TIMESTAMP NOT NULL,
    jumlah_peserta INTEGER NOT NULL,
    total_harga BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'paid', 'refunded')),
    booking_code VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_ruangan ON bookings(ruangan_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_code ON bookings(booking_code);

-- 5. Tabel Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(id),
    user_id INT NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'expired')),
    external_id VARCHAR(255) UNIQUE,
    xendit_id VARCHAR(255),
    last_webhook_id VARCHAR(255),
    payment_url TEXT,
    email_sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_booking ON transactions(booking_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_external_id ON transactions(external_id);

-- 6. Tabel Reviews
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    ruangan_id INT NOT NULL REFERENCES ruangan(id),
    booking_id INT NOT NULL REFERENCES bookings(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    komentar TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, booking_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_ruangan ON reviews(ruangan_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user ON reviews(user_id);

-- Seed Data (idempotent)
INSERT INTO users (email, nama_lengkap, ttl, no_hp, password, role, is_verified)
VALUES
('admin@mail.com', 'Super Admin', '1990-01-01', '+628111111111', '$2a$10$xyz...', 'admin', TRUE),
('user@mail.com', 'Budi Santoso', '1998-05-20', '+6281234567890', '$2a$10$abc...', 'user', TRUE)
ON CONFLICT (email) DO NOTHING;

INSERT INTO kategori (nama_kategori, deskripsi)
VALUES
('Meeting Room', 'Ruangan untuk meeting dan presentasi'),
('Co-working Space', 'Ruang kerja bersama yang fleksibel')
ON CONFLICT (nama_kategori) DO NOTHING;

INSERT INTO ruangan (
    nama_ruangan,
    kategori_id,
    deskripsi,
    alamat,
    kota,
    kapasitas,
    harga_per_jam,
    harga_per_hari,
    stock_availability,
    fasilitas,
    gambar,
    is_active
)
VALUES
(
    'Ruang Mawar',
    (SELECT id FROM kategori WHERE nama_kategori = 'Meeting Room'),
    'Ruang rapat kapasitas 10 orang',
    'Jl. Sudirman No. 1',
    'Jakarta',
    10,
    150000,
    1000000,
    1,
    '["wifi", "projector"]'::jsonb,
    '[]'::jsonb,
    TRUE
),
(
    'Meja Fokus A1',
    (SELECT id FROM kategori WHERE nama_kategori = 'Co-working Space'),
    'Meja personal tenang',
    'Jl. Dago No. 10',
    'Bandung',
    1,
    20000,
    120000,
    5,
    '["wifi", "power_outlet"]'::jsonb,
    '[]'::jsonb,
    TRUE
)
ON CONFLICT DO NOTHING;

INSERT INTO bookings (
    user_id,
    ruangan_id,
    tanggal_mulai,
    tanggal_selesai,
    jumlah_peserta,
    total_harga,
    booking_code,
    status,
    payment_status
)
SELECT
    u.id,
    r.id,
    '2026-04-10 09:00:00',
    '2026-04-10 11:00:00',
    5,
    300000,
    'BKNG-20260406-001',
    'pending',
    'unpaid'
FROM users u
JOIN ruangan r ON r.nama_ruangan = 'Ruang Mawar'
WHERE u.email = 'user@mail.com'
ON CONFLICT (booking_code) DO NOTHING;


INSERT INTO kategori (nama_kategori, deskripsi) VALUES
('Meeting Room', 'Ruangan untuk meeting dan presentasi'),
('Co-working Space', 'Ruang kerja bersama yang fleksibel'),
('Event Space', 'Ruangan untuk event dan gathering'),
('Private Office', 'Kantor pribadi untuk tim')
ON CONFLICT (nama_kategori) DO NOTHING;

INSERT INTO reviews (user_id, ruangan_id, booking_id, rating, komentar)
SELECT
    u.id,
    r.id,
    b.id,
    5,
    'Ruangan nyaman, bersih, dan fasilitas lengkap.'
FROM users u
JOIN bookings b ON b.user_id = u.id
JOIN ruangan r ON r.id = b.ruangan_id
WHERE u.email = 'user@mail.com'
ON CONFLICT (user_id, booking_id) DO NOTHING;

