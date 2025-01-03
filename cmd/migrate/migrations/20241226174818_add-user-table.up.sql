CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  firstName VARCHAR(255) NOT NULL,
  lastName VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
