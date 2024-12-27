DROP TABLE IF EXISTS documents;

CREATE TABLE documents (
    id STRING PRIMARY KEY,
    document_content TEXT NOT NULL,
    document_title TEXT,
    signer_name VARCHAR(100) NOT NULL,
    signer_email VARCHAR(100) NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    callback_url VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    signature_data TEXT,
    consents JSON
);