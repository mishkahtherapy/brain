# Brain

- ✅ Add TimezoneOffset to domain entities that you want to show for a second person. Examples
  - Add timezoneOffset to a TimeSlot to display the therapist's start/end time to an operator.
  - Add timezoneOffset to a session to display it in client's time.

### Assumptions

- ✅ All times that are inputs/outputs of the backend are in UTC timezone
- ✅ All durations are in minutes
- ✅ If a timezone is attached to a domain entity, then the timezone is a frontend hint to do proper adjustments, no timezone adjustments happen on the backend.
- ❌ There is no such thing as an anonymuous session
