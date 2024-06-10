CREATE TABLE IF NOT EXISTS users
(
    username TEXT PRIMARY KEY,
    password TEXT NOT NULL,
    role     TEXT NOT NULL
);

INSERT INTO users (username, password, role) VALUES ('admin', '$2a$10$cFmATgr3Sv9TRDHUQ64qZeEguzTMDIJKWz/euADw0dB5D2Lc5NjYm', 'admin');
INSERT INTO users (username, password, role) VALUES ('user1', '$2a$10$cFmATgr3Sv9TRDHUQ64qZeEguzTMDIJKWz/euADw0dB5D2Lc5NjYm', 'user');
INSERT INTO users (username, password, role) VALUES ('user2', '$2a$10$cFmATgr3Sv9TRDHUQ64qZeEguzTMDIJKWz/euADw0dB5D2Lc5NjYm', 'user');