
-- 1. Tabel Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nama_lengkap VARCHAR(255) NOT NULL,
    ttl DATE,
    no_hp VARCHAR(20),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    otp_code VARCHAR(10),
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Tabel Kategori
CREATE TABLE kategori (
    id SERIAL PRIMARY KEY,
    nama_kategori VARCHAR(100) NOT NULL,
    deskripsi TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Tabel Ruangan
CREATE TABLE ruangan (
    id SERIAL PRIMARY KEY,
    nama_ruangan VARCHAR(255) NOT NULL,
    kategori_id INT REFERENCES kategori(id) ON DELETE SET NULL,
    deskripsi TEXT,
    alamat VARCHAR(255),
    kota VARCHAR(100),
    kapasitas INT,
    harga_per_jam DECIMAL(12, 2),
    harga_per_hari DECIMAL(12, 2),
    stock_availability INT DEFAULT 1,
    fasilitas JSONB, -- Menggunakan JSONB untuk fleksibilitas di PostgreSQL
    gambar VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Tabel Bookings
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    ruangan_id INT REFERENCES ruangan(id) ON DELETE CASCADE,
    tanggal_mulai TIMESTAMPTZ NOT NULL,
    tanggal_selesai TIMESTAMPTZ NOT NULL,
    jumlah_peserta INT,
    total_harga DECIMAL(15, 2),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    payment_status VARCHAR(20) DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'paid', 'refunded')),
    booking_code VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Tabel Transactions (Sesuai integrasi Xendit)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    booking_id INT REFERENCES bookings(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,
    payment_method VARCHAR(50),
    transaction_date TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'expired')),
    xendit_id VARCHAR(255), -- ID transaksi dari Xendit API
    payment_url TEXT,      -- URL Invoice Xendit
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Tabel Reviews
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    ruangan_id INT REFERENCES ruangan(id) ON DELETE CASCADE,
    booking_id INT REFERENCES bookings(id) ON DELETE CASCADE,
    rating INT CHECK (rating >= 1 AND rating <= 5),
    komentar TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);


-- Insert Data Users 
INSERT INTO users (email, password, nama_lengkap, role, is_verified) 
VALUES 
('admin@mail.com', '$2a$10$xyz...', 'Super Admin', 'admin', true),
('user@mail.com', '$2a$10$abc...', 'Budi Santoso', 'user', true);

-- Insert Data Kategori
INSERT INTO kategori (nama_kategori, deskripsi) 
VALUES 
('Meeting Room', 'Ruangan formal untuk rapat'),
('Coworking Space', 'Meja kerja individu di area terbuka');

-- Insert Data Ruangan
INSERT INTO ruangan (nama_ruangan, kategori_id, deskripsi, kota, kapasitas, harga_per_jam, fasilitas) 
VALUES 
('Ruang Mawar', 1, 'Ruang rapat kapasitas 10 orang', 'Jakarta', 10, 150000.00, '{"wifi": true, "projector": true, "coffee": false}'),
('Meja Fokus A1', 2, 'Meja personal tenang', 'Bandung', 1, 20000.00, '{"wifi": true, "power_outlet": true}');

-- Insert Contoh Booking
INSERT INTO bookings (user_id, ruangan_id, tanggal_mulai, tanggal_selesai, jumlah_peserta, total_harga, booking_code, status)
VALUES 
(2, 1, '2026-04-10 09:00:00+07', '2026-04-10 11:00:00+07', 5, 300000.00, 'BKNG-20260406-001', 'pending');