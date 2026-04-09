-- UP
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

CREATE INDEX idx_reviews_ruangan ON reviews(ruangan_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- DOWN
DROP TABLE IF EXISTS reviews;
