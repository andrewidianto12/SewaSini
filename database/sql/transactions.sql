-- UP
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(id),
    user_id INT NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    external_id VARCHAR(255) UNIQUE,
    xendit_id VARCHAR(255),
    last_webhook_id VARCHAR(255),
    payment_url TEXT,
    email_sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_booking ON transactions(booking_id);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE UNIQUE INDEX idx_transactions_external_id ON transactions(external_id);

-- DOWN
DROP TABLE IF EXISTS transactions;
