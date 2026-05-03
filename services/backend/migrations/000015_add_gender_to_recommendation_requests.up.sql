ALTER TABLE recommendation_requests
    ADD COLUMN recipient_gender TEXT NULL
        CONSTRAINT chk_recommendation_requests_gender
            CHECK (recipient_gender IN ('male', 'female', 'other'));
