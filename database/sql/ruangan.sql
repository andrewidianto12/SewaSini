-- UP
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

CREATE INDEX idx_ruangan_kategori ON ruangan(kategori_id);
CREATE INDEX idx_ruangan_kota ON ruangan(kota);
CREATE INDEX idx_ruangan_active ON ruangan(is_active);

-- DOWN
DROP TABLE IF EXISTS ruangan;
