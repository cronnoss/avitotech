-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id SERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    username      VARCHAR(255) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE balances
(
    id       SERIAL PRIMARY KEY,
    user_id  INT REFERENCES users (id) ON DELETE CASCADE,
    amount   NUMERIC(15, 2) NOT NULL,
    currency VARCHAR(10)    NOT NULL DEFAULT 'RUB'
);

CREATE TABLE transactions
(
    transaction_id SERIAL PRIMARY KEY,
    user_id  INT REFERENCES users (id) ON DELETE CASCADE,
    amount         NUMERIC(15, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'RUB',
    operation      VARCHAR(255)   NOT NULL,
    date           TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_results
(
    from_user_id INT REFERENCES users (id) ON DELETE CASCADE,
    to_user_id   INT REFERENCES users (id) ON DELETE CASCADE,
    amount       NUMERIC(15, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'RUB',
    status       VARCHAR(50)    NOT NULL,
    PRIMARY KEY (from_user_id, to_user_id, status)
);
-- +goose StatementEnd
