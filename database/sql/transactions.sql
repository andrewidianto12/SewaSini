-- UP
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::varchar,
    booking_id VARCHAR(36) NOT NULL REFERENCES bookings(id),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    xendit_id VARCHAR(255),
    payment_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_booking ON transactions(booking_id);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);

-- DOWN
DROP TABLE IF EXISTS transactions;
