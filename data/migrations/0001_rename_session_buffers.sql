ALTER TABLE time_slots
RENAME COLUMN pre_session_buffer TO advance_notice;

ALTER TABLE time_slots
RENAME COLUMN post_session_buffer TO after_session_break_time;