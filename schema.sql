CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hash VARCHAR(255) UNIQUE NOT NULL,
    height BIGINT UNIQUE NOT NULL,
    time TIMESTAMP NOT NULL,
    num_txs INTEGER NOT NULL DEFAULT 0,
    total_txs BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    index_num BIGINT NOT NULL,
    hash VARCHAR(255) UNIQUE NOT NULL,
    block_height BIGINT NOT NULL,
    success BOOLEAN NOT NULL,
    gas_wanted BIGINT,
    gas_used BIGINT,
    memo TEXT,
    gas_fee JSONB NOT NULL,
    messages JSONB NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (block_height) REFERENCES blocks(height)
);

CREATE TABLE token_events (
    id BIGSERIAL PRIMARY KEY,
    transaction_hash VARCHAR(255) NOT NULL,
    tx_event_index INT NOT NULL,
    pkg_path VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    func VARCHAR(50) NOT NULL,
    from_addr VARCHAR(50) NOT NULL,
    to_addr VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(transaction_hash, tx_event_index)
);

CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    token_path VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT uk_balances_address_token UNIQUE(address, token_path)
);