-- +goose Up
-- +goose Down
-- +goose StatementBegin
DROP TABLE transfer_results;
DROP TABLE transactions;
DROP TABLE balances;
DROP TABLE users;
-- +goose StatementEnd
