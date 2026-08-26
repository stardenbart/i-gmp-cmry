-- AES-256-GCM ciphertext is larger than its plaintext. TEXT prevents valid
-- encrypted descriptions and legacy photo metadata from exceeding VARCHAR(255).
ALTER TABLE "Issue" ALTER COLUMN "Keterangan" TYPE TEXT;
ALTER TABLE "Issue_Photo" ALTER COLUMN "ImageUrl" TYPE TEXT;
ALTER TABLE "Issue_Photo" ALTER COLUMN "FileName" TYPE TEXT;
ALTER TABLE "Issue_Photo" ALTER COLUMN "Keterangan" TYPE TEXT;
