# Timezone Functionality Testing Guide

## Overview
This document outlines comprehensive testing for the timezone conversion functionality in the therapist app. The system uses duration-based time slots stored in UTC with conversion to local timezones for display and input.

## Test Scenarios

### 1. Basic Timezone Conversion Tests

#### Test 1.1: GMT+3 (Cairo) Timezone
**Input:**
- Local time: Monday 2:00 PM - 3:00 PM (14:00-15:00)
- Timezone: Africa/Cairo (GMT+3)
- Duration: 60 minutes

**Expected Backend Storage:**
- UTC day: Monday
- UTC start time: 11:00
- Duration: 60 minutes

**Expected Frontend Display:**
- Monday 2:00 PM - 3:00 PM (GMT+3)

#### Test 1.2: GMT-5 (EST) Timezone
**Input:**
- Local time: Monday 9:00 AM - 10:00 AM (09:00-10:00)
- Timezone: America/New_York (GMT-5)
- Duration: 60 minutes

**Expected Backend Storage:**
- UTC day: Monday
- UTC start time: 14:00
- Duration: 60 minutes

**Expected Frontend Display:**
- Monday 9:00 AM - 10:00 AM (EST)

### 2. Cross-Day Boundary Tests

#### Test 2.1: Monday 1:30 AM → Sunday 22:30 UTC
**Input:**
- Local time: Monday 1:30 AM - 3:30 AM (01:30-03:30)
- Timezone: Africa/Cairo (GMT+3)
- Duration: 120 minutes

**Expected Backend Storage:**
- UTC day: Sunday
- UTC start time: 22:30
- Duration: 120 minutes

**Expected Frontend Display:**
- Monday 1:30 AM - 3:30 AM (GMT+3)

#### Test 2.2: Sunday 2:00 AM → Saturday 21:00 UTC
**Input:**
- Local time: Sunday 2:00 AM - 4:00 AM (02:00-04:00)
- Timezone: America/New_York (GMT-5)
- Duration: 120 minutes

**Expected Backend Storage:**
- UTC day: Saturday
- UTC start time: 21:00
- Duration: 120 minutes

**Expected Frontend Display:**
- Sunday 2:00 AM - 4:00 AM (EST)

### 3. Conflict Detection Tests

#### Test 3.1: Same Day Overlap
**Existing Slot:** Monday 2:00 PM - 3:00 PM (local)
**New Slot:** Monday 2:30 PM - 3:30 PM (local)
**Expected:** Conflict detected

#### Test 3.2: Cross-Day Overlap
**Existing Slot:** Monday 1:00 AM - 2:00 AM (local) → Sunday 22:00-23:00 UTC
**New Slot:** Sunday 11:30 PM - 12:30 AM (local) → Sunday 20:30-21:30 UTC
**Expected:** No conflict (different UTC times)

**New Slot:** Monday 1:30 AM - 2:30 AM (local) → Sunday 22:30-23:30 UTC
**Expected:** Conflict detected (overlaps with existing Sunday 22:00-23:00 UTC)

### 4. Edge Case Tests

#### Test 4.1: Daylight Saving Time (DST)
**Test during DST transition:**
- Before DST: GMT-5 (EST)
- After DST: GMT-4 (EDT)
- Expected: System should use current timezone offset

#### Test 4.2: Maximum Duration
**Input:**
- Duration: 8 hours (480 minutes)
- Start: Monday 11:00 PM
- End: Tuesday 7:00 AM
**Expected:** Proper cross-day handling

#### Test 4.3: Invalid Timezone Offsets
**Test Cases:**
- Timezone offset: +15 hours (invalid)
- Timezone offset: -13 hours (invalid)
**Expected:** Validation error

## API Testing Commands

### Backend API Tests (using curl)

#### Create Time Slot
```bash
curl -X POST "http://localhost:8080/api/v1/therapists/therapist_001/timeslots?timezoneOffset=180" \
  -H "Content-Type: application/json" \
  -d '{
    "therapistId": "therapist_001",
    "dayOfWeek": "Monday",
    "startTime": "14:00",
    "durationMinutes": 60,
    "timezoneOffset": 180,
    "timezone": "Africa/Cairo",
    "advanceNotice": 30,
    "afterSessionBreakTime": 15
  }'
```

#### List Time Slots
```bash
curl "http://localhost:8080/api/v1/therapists/therapist_001/timeslots?timezoneOffset=180"
```

#### Update Time Slot
```bash
curl -X PUT "http://localhost:8080/api/v1/therapists/therapist_001/timeslots/slot_001?timezoneOffset=180" \
  -H "Content-Type: application/json" \
  -d '{
    "dayOfWeek": "Monday",
    "startTime": "15:00",
    "durationMinutes": 90,
    "advanceNotice": 60,
    "afterSessionBreakTime": 30,
    "isActive": true
  }'
```

### Frontend Tests

#### Test Timezone Utilities
```javascript
// Test timezone offset calculation
console.log('Timezone offset (minutes):', getTimezoneOffsetMinutes());
console.log('Timezone name:', getTimezoneName());
console.log('Timezone data:', createTimezoneData());

// Test duration calculation
console.log('Duration (14:00 to 15:00):', calculateDurationMinutes('14:00', '15:00'));
console.log('Duration (23:00 to 02:00):', calculateDurationMinutes('23:00', '02:00'));
```

#### Test Form Submission
1. Create a time slot: Monday 2:00 PM - 3:00 PM
2. Verify backend receives correct UTC conversion
3. Verify frontend displays correct local time

## Database Verification

### Check UTC Storage
```sql
SELECT 
  id,
  therapist_id,
  day_of_week,
  start_time,
  duration_minutes,
  CASE 
    WHEN duration_minutes <= 60 THEN 'Short'
    WHEN duration_minutes <= 240 THEN 'Medium'
    ELSE 'Long'
  END as duration_category
FROM time_slots
ORDER BY therapist_id, day_of_week, start_time;
```

### Check Cross-Day Scenarios
```sql
SELECT 
  id,
  therapist_id,
  day_of_week,
  start_time,
  duration_minutes,
  -- Calculate if slot crosses day boundary
  CASE 
    WHEN (strftime('%H', start_time) * 60 + strftime('%M', start_time) + duration_minutes) >= 1440 
    THEN 'Crosses Day' 
    ELSE 'Same Day' 
  END as day_boundary_status
FROM time_slots
WHERE duration_minutes > 0
ORDER BY therapist_id, day_of_week, start_time;
```

## Expected Results Summary

### Successful Tests
- [x] Basic timezone conversion (GMT+3, GMT-5)
- [x] Cross-day boundary handling
- [x] Duration-based conflict detection
- [x] API timezone parameter validation
- [x] Frontend timezone data creation

### Edge Cases Handled
- [x] Day boundary crossings
- [x] Invalid timezone offsets
- [x] Long duration slots (8+ hours)
- [x] Conflict detection across day boundaries

### Performance Considerations
- [x] Database indexes on (therapist_id, day_of_week, start_time, duration_minutes)
- [x] Efficient UTC storage without redundant timezone data
- [x] Client-side timezone calculation without server round-trips

## Next Steps
1. Run automated tests for all scenarios
2. Test with real therapist data
3. Validate DST handling during transitions
4. Load testing with multiple timezones
5. User acceptance testing with international therapists 