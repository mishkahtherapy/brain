-- =============================================================================
-- SAMPLE DATA FOR DEVELOPMENT
-- =============================================================================

-- Sample therapists
INSERT INTO
    therapists (
        id,
        name,
        email,
        phone_number,
        whatsapp_number,
        speaks_english
    )
VALUES (
        'therapist_001',
        'Dr. Sarah Johnson',
        'sarah.johnson@therapy.com',
        '+1555001234',
        '+1234567890',
        TRUE
    ),
    (
        'therapist_002',
        'Dr. Michael Chen',
        'michael.chen@therapy.com',
        '+1555005678',
        '+1987654321',
        TRUE
    ),
    (
        'therapist_003',
        'Dr. Emily Davis',
        'emily.davis@therapy.com',
        '+1555098765',
        '+1122334455',
        TRUE
    );

-- Sample clients
INSERT INTO
    clients (
        id,
        name,
        whatsapp_number,
        timezone_offset
    )
VALUES (
        'client_001',
        'John Doe',
        '+1555123456',
        -300
    ), -- GMT-5 (America/New_York)
    (
        'client_002',
        'Jane Smith',
        '+1555789012',
        0
    );
-- GMT   (Europe/London)

-- Sample time slots (stored in UTC timezone with duration)
-- Note: These examples assume therapists are in Cairo (UTC+3) timezone
-- Local times: 12:00-1:00 PM → UTC 09:00 (60 min), 2:45-3:45 PM → UTC 11:45 (60 min), etc.
-- Effective time ranges include buffers and maintain 30-minute minimum gaps
INSERT INTO
    time_slots (
        id,
        therapist_id,
        day_of_week,
        start_time,
        duration_minutes,
        pre_session_buffer,
        post_session_buffer
    )
VALUES (
        'slot_001',
        'therapist_001',
        'Monday',
        '09:00',
        60,
        4320,
        30
    ), -- Cairo: 12:00-1:00 PM Monday (effective: 11:30-1:15 PM)
    (
        'slot_002',
        'therapist_001',
        'Monday',
        '11:45',
        60,
        4320,
        30
    ), -- Cairo: 2:45-3:45 PM Monday (effective: 2:15-4:00 PM)
    (
        'slot_003',
        'therapist_001',
        'Tuesday',
        '14:00',
        60,
        4320,
        30
    ), -- Cairo: 5:00-6:00 PM Tuesday
    (
        'slot_004',
        'therapist_002',
        'Wednesday',
        '09:00',
        60,
        4320,
        30
    ), -- Cairo: 12:00-1:00 PM Wednesday
    (
        'slot_005',
        'therapist_002',
        'Monday',
        '09:00',
        60,
        4320,
        30
    ),
    (
        'slot_006',
        'therapist_003',
        'Monday',
        '09:00',
        60,
        300,
        30
    );

-- Sample bookings (1-hour appointments)
INSERT INTO
    bookings (
        id,
        timeslot_id,
        therapist_id,
        client_id,
        start_time,
        duration_minutes,
        timezone_offset,
        state
    )
VALUES (
        'booking_001',
        'slot_001',
        'therapist_001',
        'client_001',
        '2024-01-15 09:00:00',
        60,
        -300,
        'confirmed'
    ),
    (
        'booking_002',
        'slot_002',
        'therapist_001',
        'client_002',
        '2024-01-15 11:45:00',
        60,
        -60,
        'pending'
    ),
    (
        'booking_003',
        'slot_003',
        'therapist_001',
        'client_001',
        '2024-01-16 14:00:00',
        60,
        300,
        'confirmed'
    );

-- Add sessions for confirmed bookings
INSERT INTO
    sessions (
        id,
        booking_id,
        therapist_id,
        client_id,
        timeslot_id,
        start_time,
        paid_amount,
        language,
        state,
        notes,
        meeting_url
    )
VALUES (
        'session_001',
        'booking_001',
        'therapist_001',
        'client_001',
        'slot_001',
        '2024-01-15 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 1',
        'https://meet.google.com/abc123'
    ),
    (
        'session_002',
        'booking_002',
        'therapist_001',
        'client_002',
        'slot_002',
        '2024-01-15 11:45:00',
        10000,
        'english',
        'done',
        'Notes for session 2',
        'https://meet.google.com/def456'
    ),
    (
        'session_003',
        'booking_003',
        'therapist_001',
        'client_001',
        'slot_003',
        '2024-01-16 14:00:00',
        10000,
        'english',
        'done',
        'Notes for session 3',
        'https://meet.google.com/ghi789'
    ),
    (
        'session_004',
        'booking_004',
        'therapist_001',
        'client_002',
        'slot_001',
        '2026-01-15 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 4',
        'https://meet.google.com/jkl012'
    ),
    (
        'session_005',
        'booking_005',
        'therapist_001',
        'client_001',
        'slot_001',
        '2026-01-15 10:00:00',
        10000,
        'english',
        'done',
        'Notes for session 5',
        'https://meet.google.com/mno345'
    ),
    (
        'session_006',
        'booking_006',
        'therapist_002',
        'client_002',
        'slot_004',
        '2026-01-17 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 6',
        'https://meet.google.com/pqr678'
    ),
    (
        'session_007',
        'booking_007',
        'therapist_001',
        'client_001',
        'slot_003',
        '2026-01-20 14:00:00',
        10000,
        'english',
        'done',
        'Notes for session 7',
        'https://meet.google.com/stu901'
    ),
    (
        'session_008',
        'booking_008',
        'therapist_002',
        'client_002',
        'slot_004',
        '2026-01-24 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 8',
        'https://meet.google.com/vwx234'
    ),
    (
        'session_009',
        'booking_009',
        'therapist_001',
        'client_001',
        'slot_001',
        '2026-01-29 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 9',
        'https://meet.google.com/yz5678'
    ),
    (
        'session_010',
        'booking_010',
        'therapist_001',
        'client_002',
        'slot_002',
        '2026-01-29 10:30:00',
        10000,
        'english',
        'done',
        'Notes for session 10',
        'https://meet.google.com/abc901'
    ),
    (
        'session_011',
        'booking_011',
        'therapist_002',
        'client_001',
        'slot_004',
        '2026-02-05 09:00:00',
        10000,
        'english',
        'done',
        'Notes for session 11',
        'https://meet.google.com/def234'
    );

-- Sample specializations
INSERT INTO
    specializations (id, name)
VALUES ('spec_001', 'Anxiety'),
    ('spec_002', 'Couples Therapy'),
    ('spec_003', 'Family Therapy'),
    ('spec_004', 'OCD');

-- Sample therapist specializations
INSERT INTO
    therapist_specializations (
        id,
        therapist_id,
        specialization_id
    )
VALUES (
        'ts_001',
        'therapist_001',
        'spec_001'
    ),
    (
        'ts_002',
        'therapist_001',
        'spec_002'
    ),
    (
        'ts_003',
        'therapist_001',
        'spec_003'
    ),
    (
        'ts_004',
        'therapist_001',
        'spec_004'
    ),
    (
        'ts_005',
        'therapist_002',
        'spec_001'
    ),
    (
        'ts_006',
        'therapist_002',
        'spec_002'
    ),
    (
        'ts_007',
        'therapist_002',
        'spec_003'
    ),
    (
        'ts_008',
        'therapist_002',
        'spec_004'
    ),
    (
        'ts_009',
        'therapist_003',
        'spec_001'
    );