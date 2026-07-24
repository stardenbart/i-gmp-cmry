CREATE TABLE IF NOT EXISTS "Notification" (
    "NotificationID" VARCHAR(50) PRIMARY KEY,
    "UserID" VARCHAR(50) NOT NULL,
    "Type" VARCHAR(20) NOT NULL, -- 'info', 'warning', 'error', 'success'
    "Title" VARCHAR(255) NOT NULL,
    "Message" TEXT NOT NULL,
    "IsRead" BOOLEAN DEFAULT FALSE,
    "Link" VARCHAR(255),
    "CreatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_notification_user FOREIGN KEY ("UserID") REFERENCES "Users"("UserID") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notification_user_id ON "Notification"("UserID");
CREATE INDEX IF NOT EXISTS idx_notification_created_at ON "Notification"("CreatedAt");
