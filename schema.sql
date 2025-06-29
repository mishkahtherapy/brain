-- Brain Database Schema
-- Optimized for SQLite (can be adapted for other databases)

-- =============================================================================
-- CORE ENTITIES
-- =============================================================================

-- Therapists table
CREATE TABLE therapists (
    id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20), -- Business/office phone number
    whatsapp_number VARCHAR(20), -- International format support, nullable
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Therapist specializations table
CREATE TABLE specializations (
    id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Therapist specializations table
CREATE TABLE therapist_specializations (
    id VARCHAR(128) PRIMARY KEY,
    therapist_id VARCHAR(128) NOT NULL,
    specialization_id VARCHAR(128) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_therapist_specializations_therapist FOREIGN KEY (therapist_id) REFERENCES therapists(id) ON DELETE CASCADE,
    CONSTRAINT fk_therapist_specializations_specialization FOREIGN KEY (specialization_id) REFERENCES specializations(id) ON DELETE CASCADE,
    -- Prevent duplicate specializations for the same therapist
    CONSTRAINT unique_therapist_specialization UNIQUE (therapist_id, specialization_id)
);

-- Clients table  
CREATE TABLE clients (
    id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    whatsapp_number VARCHAR(20), -- International format support, nullable
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Time slots table (therapist availability)
CREATE TABLE time_slots (
    id VARCHAR(128) PRIMARY KEY,
    therapist_id VARCHAR(128) NOT NULL,
    day_of_week VARCHAR(10) NOT NULL CHECK (day_of_week IN ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday')),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    pre_session_buffer INTEGER NOT NULL DEFAULT 0, -- minutes
    post_session_buffer INTEGER NOT NULL DEFAULT 0, -- minutes
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_time_slots_therapist FOREIGN KEY (therapist_id) REFERENCES therapists(id) ON DELETE CASCADE,
    CONSTRAINT check_time_order CHECK (end_time > start_time),
    CONSTRAINT check_positive_buffers CHECK (pre_session_buffer >= 0 AND post_session_buffer >= 0)
);

-- Bookings table (actual appointments - always 1 hour duration)
CREATE TABLE bookings (
    id VARCHAR(128) PRIMARY KEY,
    time_slot_id VARCHAR(128) NOT NULL,
    therapist_id VARCHAR(128) NOT NULL,
    client_id VARCHAR(128) NOT NULL,
    start_time DATETIME NOT NULL, -- Specific start datetime for this 1-hour booking
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no_show')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bookings_time_slot FOREIGN KEY (time_slot_id) REFERENCES time_slots(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_therapist FOREIGN KEY (therapist_id) REFERENCES therapists(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_client FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
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
CREATE INDEX idx_bookings_start_time ON bookings(start_time);
CREATE INDEX idx_bookings_therapist_start_time ON bookings(therapist_id, start_time);
CREATE INDEX idx_bookings_status ON bookings(status);

-- Prevent overlapping 1-hour bookings for the same therapist
CREATE UNIQUE INDEX idx_no_overlapping_bookings ON bookings(therapist_id, start_time) WHERE status = 'scheduled';

-- Combined index for finding available slots
CREATE INDEX idx_availability_search ON time_slots(therapist_id, day_of_week, start_time, end_time);

-- =============================================================================
-- VIEWS FOR COMMON QUERIES
-- =============================================================================

-- Available time slots (slots with no bookings for specific date)
-- CREATE VIEW available_slots AS
-- SELECT 
--     ts.id,
--     ts.therapist_id,
--     t.name as therapist_name,
--     ts.day_of_week,
--     ts.start_time,
--     ts.end_time,
--     ts.pre_session_buffer,
--     ts.post_session_buffer
-- FROM time_slots ts
-- JOIN therapists t ON ts.therapist_id = t.id;

-- -- Therapist schedule with booking status
-- CREATE VIEW therapist_schedule AS
-- SELECT 
--     ts.id as time_slot_id,
--     ts.therapist_id,
--     t.name as therapist_name,
--     ts.day_of_week,
--     ts.start_time as slot_start_time,
--     ts.end_time as slot_end_time,
--     ts.pre_session_buffer,
--     ts.post_session_buffer,
--     b.id as booking_id,
--     b.start_time as booking_start_time,
--     datetime(b.start_time, '+1 hour') as booking_end_time,
--     b.status as booking_status,
--     c.name as client_name,
--     c.email as client_email
-- FROM time_slots ts
-- JOIN therapists t ON ts.therapist_id = t.id
-- LEFT JOIN bookings b ON ts.id = b.time_slot_id
-- LEFT JOIN clients c ON b.client_id = c.id;
