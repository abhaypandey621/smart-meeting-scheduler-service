-- Up Migration: Create calendar_events table
CREATE TABLE IF NOT EXISTS calendar_events (
    id VARCHAR(191) NOT NULL,
    title LONGTEXT,
    start_time DATETIME(3) DEFAULT NULL,
    end_time DATETIME(3) DEFAULT NULL,
    user_id VARCHAR(191) DEFAULT NULL,
    created_at DATETIME(3) DEFAULT NULL,
    updated_at DATETIME(3) DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_calendar_events_user_id (user_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;