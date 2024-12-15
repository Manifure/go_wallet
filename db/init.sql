CREATE TABLE IF NOT EXISTS wallets (
    wallet_id UUID PRIMARY KEY,
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00
);