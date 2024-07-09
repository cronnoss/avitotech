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
    id      SERIAL PRIMARY KEY,
    user_id INT REFERENCES users (id) ON DELETE CASCADE UNIQUE,
    amount  NUMERIC(15, 2) NOT NULL
);

CREATE TABLE transactions
(
    id        SERIAL PRIMARY KEY,
    user_id   INT REFERENCES users (id) ON DELETE CASCADE,
    amount    NUMERIC(15, 2) NOT NULL,
    operation VARCHAR(255)   NOT NULL,
    date      TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_results
(
    from_user_id INT REFERENCES users (id) ON DELETE CASCADE,
    to_user_id   INT REFERENCES users (id) ON DELETE CASCADE,
    amount       NUMERIC(15, 2) NOT NULL,
    status       VARCHAR(50)    NOT NULL,
    PRIMARY KEY (from_user_id, to_user_id, status)
);

DO
$$
    BEGIN
        IF (SELECT COUNT(*) FROM users) = 0 THEN
            INSERT INTO users (name, username, email, password_hash)
            VALUES ('John Doe', 'johndoe', 'john@example.com', 'hashedpassword123'),
                   ('Jane Smith', 'janesmith', 'jane@example.com', 'hashedpassword456'),
                   ('Alice Johnson', 'alicej', 'alice@example.com', 'hashedpassword789');
        END IF;
    END
$$;
-- +goose StatementEnd
