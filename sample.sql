-- =============================================================================
-- SAMPLE DATA FOR DEVELOPMENT
-- =============================================================================

-- Sample therapists
INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, speaks_english) VALUES
('therapist_001', 'Dr. Sarah Johnson', 'sarah.johnson@therapy.com', '+1555001234', '+1234567890', true),
('therapist_002', 'Dr. Michael Chen', 'michael.chen@therapy.com', '+1555005678', '+1987654321', true);

-- Sample clients
INSERT INTO clients (id, name, whatsapp_number) VALUES
('client_001', 'John Doe', '+1555123456'),
('client_002', 'Jane Smith', '+1555789012');

-- Sample time slots (stored in UTC timezone with duration)
-- Note: These examples assume therapists are in Cairo (UTC+3) timezone
-- Local times: 12:00-1:00 PM → UTC 09:00 (60 min), 2:45-3:45 PM → UTC 11:45 (60 min), etc.
-- Effective time ranges include buffers and maintain 30-minute minimum gaps
INSERT INTO time_slots (id, therapist_id, day_of_week, start_time, duration_minutes, pre_session_buffer, post_session_buffer) VALUES
('slot_001', 'therapist_001', 'Monday', '09:00', 60, 30, 30),      -- Cairo: 12:00-1:00 PM Monday (effective: 11:30-1:15 PM)
('slot_002', 'therapist_001', 'Monday', '11:45', 60, 30, 30),      -- Cairo: 2:45-3:45 PM Monday (effective: 2:15-4:00 PM)
('slot_003', 'therapist_001', 'Tuesday', '14:00', 60, 60, 30),     -- Cairo: 5:00-6:00 PM Tuesday
('slot_004', 'therapist_002', 'Wednesday', '09:00', 60, 120, 30);  -- Cairo: 12:00-1:00 PM Wednesday

-- Sample bookings (1-hour appointments)
INSERT INTO bookings (id, timeslot_id, therapist_id, client_id, start_time, state) VALUES
('booking_001', 'slot_001', 'therapist_001', 'client_001', '2024-01-15 09:00:00', 'confirmed'),
('booking_002', 'slot_002', 'therapist_001', 'client_002', '2024-01-15 11:45:00', 'pending'),
('booking_003', 'slot_003', 'therapist_001', 'client_001', '2024-01-16 14:00:00', 'confirmed');


-- Add sessions for confirmed bookings
INSERT INTO sessions (id, booking_id, therapist_id, client_id, timeslot_id, start_time, paid_amount, language, state, notes, meeting_url) VALUES
('session_001', 'booking_001', 'therapist_001', 'client_001', 'slot_001', '2024-01-15 09:00:00', 10000, 'english', 'done', 'Notes for session 1', 'https://meet.google.com/abc123'),
('session_002', 'booking_002', 'therapist_001', 'client_002', 'slot_002', '2024-01-15 11:45:00', 10000, 'english', 'done', 'Notes for session 2', 'https://meet.google.com/def456'),
('session_003', 'booking_003', 'therapist_001', 'client_001', 'slot_003', '2024-01-16 14:00:00', 10000, 'english', 'done', 'Notes for session 3', 'https://meet.google.com/ghi789'),
('session_004', 'booking_004', 'therapist_001', 'client_002', 'slot_001', '2026-01-15 09:00:00', 10000, 'english', 'done', 'Notes for session 4', 'https://meet.google.com/jkl012'),
('session_005', 'booking_005', 'therapist_001', 'client_001', 'slot_001', '2026-01-15 10:00:00', 10000, 'english', 'done', 'Notes for session 5', 'https://meet.google.com/mno345'),
('session_006', 'booking_006', 'therapist_002', 'client_002', 'slot_004', '2026-01-17 09:00:00', 10000, 'english', 'done', 'Notes for session 6', 'https://meet.google.com/pqr678'),
('session_007', 'booking_007', 'therapist_001', 'client_001', 'slot_003', '2026-01-20 14:00:00', 10000, 'english', 'done', 'Notes for session 7', 'https://meet.google.com/stu901'),
('session_008', 'booking_008', 'therapist_002', 'client_002', 'slot_004', '2026-01-24 09:00:00', 10000, 'english', 'done', 'Notes for session 8', 'https://meet.google.com/vwx234'),
('session_009', 'booking_009', 'therapist_001', 'client_001', 'slot_001', '2026-01-29 09:00:00', 10000, 'english', 'done', 'Notes for session 9', 'https://meet.google.com/yz5678'),
('session_010', 'booking_010', 'therapist_001', 'client_002', 'slot_002', '2026-01-29 10:30:00', 10000, 'english', 'done', 'Notes for session 10', 'https://meet.google.com/abc901'),
('session_011', 'booking_011', 'therapist_002', 'client_001', 'slot_004', '2026-02-05 09:00:00', 10000, 'english', 'done', 'Notes for session 11', 'https://meet.google.com/def234');

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
    WHERE timeslot_id = 'slot_001' 
    AND start_time = '2024-01-22 10:00:00'
    AND state IN ('confirmed')
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
        WHERE timeslot_id = hs.id 
        AND start_time = hs.slot_hour
        AND state IN ('confirmed')
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