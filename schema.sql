-- Therapist Scheduling System Database Schema
-- Optimized for PostgreSQL (can be adapted for other databases)

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- CORE ENTITIES
-- =============================================================================

-- Therapists table
CREATE TABLE therapists (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20), -- Business/office phone number
    whatsapp_number VARCHAR(20), -- International format support, nullable
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Clients table  
CREATE TABLE clients (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    whatsapp_number VARCHAR(20), -- International format support, nullable
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Time slots table (therapist availability)
CREATE TABLE time_slots (
    id VARCHAR(50) PRIMARY KEY,
    therapist_id VARCHAR(50) NOT NULL,
    day_of_week VARCHAR(10) NOT NULL CHECK (day_of_week IN ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday')),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    pre_session_buffer INTEGER NOT NULL DEFAULT 0, -- minutes
    post_session_buffer INTEGER NOT NULL DEFAULT 0, -- minutes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_time_slots_therapist FOREIGN KEY (therapist_id) REFERENCES therapists(id) ON DELETE CASCADE,
    CONSTRAINT check_time_order CHECK (end_time > start_time),
    CONSTRAINT check_positive_buffers CHECK (pre_session_buffer >= 0 AND post_session_buffer >= 0)
);

-- Bookings table (actual appointments)
CREATE TABLE bookings (
    id VARCHAR(50) PRIMARY KEY,
    time_slot_id VARCHAR(50) NOT NULL,
    therapist_id VARCHAR(50) NOT NULL,
    client_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL, -- Specific date for this booking
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no_show')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bookings_time_slot FOREIGN KEY (time_slot_id) REFERENCES time_slots(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_therapist FOREIGN KEY (therapist_id) REFERENCES therapists(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_client FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    
    -- Prevent double booking the same slot on the same date
    CONSTRAINT unique_slot_date UNIQUE (time_slot_id, date)
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Therapist queries
CREATE INDEX idx_therapists_email ON therapists(email);

-- Client queries  
CREATE INDEX idx_clients_email ON clients(email);

-- Time slot queries (most critical for scheduling)
CREATE INDEX idx_time_slots_therapist ON time_slots(therapist_id);
CREATE INDEX idx_time_slots_day ON time_slots(day_of_week);
CREATE INDEX idx_time_slots_therapist_day ON time_slots(therapist_id, day_of_week);
CREATE INDEX idx_time_slots_time_range ON time_slots(start_time, end_time);

-- Booking queries
CREATE INDEX idx_bookings_therapist ON bookings(therapist_id);
CREATE INDEX idx_bookings_client ON bookings(client_id);
CREATE INDEX idx_bookings_date ON bookings(date);
CREATE INDEX idx_bookings_therapist_date ON bookings(therapist_id, date);
CREATE INDEX idx_bookings_status ON bookings(status);

-- Combined index for finding available slots
CREATE INDEX idx_availability_search ON time_slots(therapist_id, day_of_week, start_time, end_time);

-- =============================================================================
-- VIEWS FOR COMMON QUERIES
-- =============================================================================

-- Available time slots (slots with no bookings for specific date)
CREATE VIEW available_slots AS
SELECT 
    ts.id,
    ts.therapist_id,
    t.name as therapist_name,
    ts.day_of_week,
    ts.start_time,
    ts.end_time,
    ts.pre_session_buffer,
    ts.post_session_buffer
FROM time_slots ts
JOIN therapists t ON ts.therapist_id = t.id;

-- Therapist schedule with booking status
CREATE VIEW therapist_schedule AS
SELECT 
    ts.id as time_slot_id,
    ts.therapist_id,
    t.name as therapist_name,
    ts.day_of_week,
    ts.start_time,
    ts.end_time,
    ts.pre_session_buffer,
    ts.post_session_buffer,
    b.id as booking_id,
    b.date as booking_date,
    b.status as booking_status,
    c.name as client_name,
    c.email as client_email
FROM time_slots ts
JOIN therapists t ON ts.therapist_id = t.id
LEFT JOIN bookings b ON ts.id = b.time_slot_id
LEFT JOIN clients c ON b.client_id = c.id;

-- =============================================================================
-- FUNCTIONS FOR BUSINESS LOGIC
-- =============================================================================

-- Function to check if a time slot is available on a specific date
CREATE OR REPLACE FUNCTION is_slot_available(
    slot_id VARCHAR(50),
    check_date DATE
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN NOT EXISTS (
        SELECT 1 FROM bookings 
        WHERE time_slot_id = slot_id 
        AND date = check_date 
        AND status IN ('scheduled')
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get effective time range including buffers
CREATE OR REPLACE FUNCTION get_effective_time_range(
    start_time TIME,
    end_time TIME,
    pre_buffer INTEGER,
    post_buffer INTEGER
) RETURNS TABLE(effective_start TIME, effective_end TIME) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (start_time - INTERVAL '1 minute' * pre_buffer)::TIME as effective_start,
        (end_time + INTERVAL '1 minute' * post_buffer)::TIME as effective_end;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- SAMPLE DATA FOR DEVELOPMENT
-- =============================================================================

-- Sample therapists
INSERT INTO therapists (id, name, email, phone_number, whatsapp_number) VALUES
('therapist_001', 'Dr. Sarah Johnson', 'sarah.johnson@therapy.com', '+1555001234', '+1234567890'),
('therapist_002', 'Dr. Michael Chen', 'michael.chen@therapy.com', '+1555005678', '+1987654321');

-- Sample clients
INSERT INTO clients (id, name, email, whatsapp_number) VALUES
('client_001', 'John Doe', 'john.doe@email.com', '+1555123456'),
('client_002', 'Jane Smith', 'jane.smith@email.com', '+1555789012');

-- Sample time slots
INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, end_time, pre_session_buffer, post_session_buffer) VALUES
('slot_001', 'therapist_001', 'Monday', '09:00', '10:00', 30, 15),
('slot_002', 'therapist_001', 'Monday', '10:30', '11:30', 30, 15),
('slot_003', 'therapist_001', 'Tuesday', '14:00', '15:00', 60, 30),
('slot_004', 'therapist_002', 'Wednesday', '09:00', '10:00', 120, 30);

-- Sample bookings
INSERT INTO bookings (id, time_slot_id, therapist_id, client_id, date) VALUES
('booking_001', 'slot_001', 'therapist_001', 'client_001', '2024-01-15'),
('booking_002', 'slot_003', 'therapist_001', 'client_002', '2024-01-16');

-- =============================================================================
-- COMMON QUERIES FOR REFERENCE
-- =============================================================================

/*
-- Find available slots for a therapist on a specific day
SELECT * FROM available_slots 
WHERE therapist_id = 'therapist_001' 
AND day_of_week = 'Monday'
AND is_slot_available(id, '2024-01-22');

-- Get all bookings for a therapist on a date range
SELECT * FROM therapist_schedule
WHERE therapist_id = 'therapist_001'
AND booking_date BETWEEN '2024-01-15' AND '2024-01-31'
ORDER BY booking_date, start_time;

-- Find potential scheduling conflicts (overlapping effective time ranges)
SELECT 
    ts1.id as slot1_id,
    ts2.id as slot2_id,
    ts1.therapist_id
FROM time_slots ts1
JOIN time_slots ts2 ON ts1.therapist_id = ts2.therapist_id 
    AND ts1.day_of_week = ts2.day_of_week
    AND ts1.id != ts2.id
CROSS JOIN get_effective_time_range(ts1.start_time, ts1.end_time, ts1.pre_session_buffer, ts1.post_session_buffer) er1
CROSS JOIN get_effective_time_range(ts2.start_time, ts2.end_time, ts2.pre_session_buffer, ts2.post_session_buffer) er2
WHERE er1.effective_start < er2.effective_end 
AND er1.effective_end > er2.effective_start;
*/ 