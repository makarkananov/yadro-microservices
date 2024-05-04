CREATE TABLE IF NOT EXISTS comics
(
    id       INT PRIMARY KEY,
    img      VARCHAR(255),
    keywords TEXT[]
);