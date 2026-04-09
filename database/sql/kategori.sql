-- UP
CREATE TABLE IF NOT EXISTS kategori (
    id SERIAL PRIMARY KEY,
    nama_kategori VARCHAR(100) UNIQUE NOT NULL,
    deskripsi TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO kategori (nama_kategori, deskripsi) VALUES
('Meeting Room', 'Ruangan untuk meeting dan presentasi'),
('Co-working Space', 'Ruang kerja bersama yang fleksibel'),
('Event Space', 'Ruangan untuk event dan gathering'),
('Private Office', 'Kantor pribadi untuk tim')
ON CONFLICT (nama_kategori) DO NOTHING;

-- DOWN
DROP TABLE IF EXISTS kategori;
