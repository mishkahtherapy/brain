ALTER TABLE bookings
ADD COLUMN client_timezone_offset INTEGER DEFAULT 0;

ALTER TABLE sessions
ADD COLUMN client_timezone_offset INTEGER DEFAULT 0;