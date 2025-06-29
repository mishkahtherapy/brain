
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

-- Sample bookings (1-hour appointments)
INSERT INTO bookings (id, time_slot_id, therapist_id, client_id, start_time) VALUES
('booking_001', 'slot_001', 'therapist_001', 'client_001', '2024-01-15 09:00:00'),
('booking_002', 'slot_001', 'therapist_001', 'client_002', '2024-01-15 10:00:00'),
('booking_003', 'slot_003', 'therapist_001', 'client_001', '2024-01-16 14:00:00');

-- =============================================================================
-- COMMON QUERIES FOR REFERENCE (SQLite Compatible)
-- =============================================================================

/*
-- Find available slots for a therapist on a specific day
SELECT * FROM available_slots 
WHERE therapist_id = 'therapist_001' 
AND day_of_week = 'Monday';

-- Check if a specific 1-hour slot is available (replaces is_slot_available function)
SELECT NOT EXISTS (
    SELECT 1 FROM bookings 
    WHERE time_slot_id = 'slot_001' 
    AND start_time = '2024-01-22 10:00:00'
    AND status IN ('scheduled')
) as is_available;

-- Get all bookings for a therapist on a date range
SELECT * FROM therapist_schedule
WHERE therapist_id = 'therapist_001'
AND date(booking_start_time) BETWEEN '2024-01-15' AND '2024-01-31'
ORDER BY booking_start_time;

-- Find available 1-hour slots within a timeslot on a specific date
WITH RECURSIVE hourly_slots AS (
    SELECT 
        ts.id,
        ts.therapist_id,
        datetime('2024-01-15 ' || ts.start_time) as slot_hour
    FROM time_slots ts
    WHERE ts.id = 'slot_001'
    
    UNION ALL
    
    SELECT 
        hs.id,
        hs.therapist_id,
        datetime(hs.slot_hour, '+1 hour')
    FROM hourly_slots hs
    JOIN time_slots ts ON hs.id = ts.id
    WHERE datetime(hs.slot_hour, '+1 hour') <= datetime('2024-01-15 ' || ts.end_time)
)
SELECT 
    hs.*,
    NOT EXISTS (
        SELECT 1 FROM bookings 
        WHERE time_slot_id = hs.id 
        AND start_time = hs.slot_hour
        AND status IN ('scheduled')
    ) as available
FROM hourly_slots hs;

-- Find potential scheduling conflicts (replaces get_effective_time_range function)
SELECT 
    ts1.id as slot1_id,
    ts2.id as slot2_id,
    ts1.therapist_id
FROM time_slots ts1
JOIN time_slots ts2 ON ts1.therapist_id = ts2.therapist_id 
    AND ts1.day_of_week = ts2.day_of_week
    AND ts1.id != ts2.id
WHERE time(ts1.start_time, '-' || ts1.pre_session_buffer || ' minutes') < 
      time(ts2.end_time, '+' || ts2.post_session_buffer || ' minutes')
AND time(ts1.end_time, '+' || ts1.post_session_buffer || ' minutes') > 
    time(ts2.start_time, '-' || ts2.pre_session_buffer || ' minutes');
*/