-- Create user table
CREATE TABLE IF NOT EXISTS user (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users
INSERT INTO
    user (name, email, password)
VALUES (
        'Alice',
        'alice@example.com',
        'password123'
    ),
    (
        'Bob',
        'bob@example.com',
        'securepass'
    ),
    (
        'Charlie',
        'charlie@example.com',
        'charliepwd'
    );