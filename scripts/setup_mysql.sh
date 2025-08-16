#!/bin/bash

# MySQL Setup Script for Meeting Scheduler

echo "Setting up MySQL database for Meeting Scheduler..."

# Check if MySQL is running
if ! mysqladmin ping -h localhost -u root -p --silent; then
    echo "Error: MySQL is not running or not accessible"
    echo "Please start MySQL and ensure you can connect with: mysql -u root -p"
    exit 1
fi

# Create database
echo "Creating database..."
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS meeting_scheduler CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Run migrations
echo "Running migrations..."
go run scripts/migrate.go

echo "MySQL setup completed successfully!"
echo ""
echo "To start the server, run:"
echo "go run cmd/server/main.go"
echo ""
echo "Note: Make sure to update the database connection string in main.go if your MySQL credentials are different." 