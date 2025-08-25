package get_schedule

import (
	"errors"
	"sort"
	"time"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/schedule"
	"github.com/mishkahtherapy/brain/core/domain/therapist"
	"github.com/mishkahtherapy/brain/core/domain/timeslot"
	"github.com/mishkahtherapy/brain/core/ports"
)

type timeRange struct {
	start domain.UTCTimestamp
	end   domain.UTCTimestamp
}

type Input struct {
	SpecializationTag string
	MustSpeakEnglish  bool
	TherapistIDs      []domain.TherapistID
	StartDate         time.Time
	EndDate           time.Time
}

type Usecase struct {
	therapistRepo                   ports.TherapistRepository
	timeSlotRepo                    ports.TimeSlotRepository
	bookingRepo                     ports.BookingRepository
	timeRangeMinimumDurationMinutes domain.DurationMinutes
}

var ErrSpecializationTagOrTherapistIDsIsRequired = errors.New("specialization tag or therapist ids is required")
var ErrInvalidDateRange = errors.New("invalid date range")
var ErrSpecializationTagAndTherapistIDsCannotBeUsedTogether = errors.New("specialization tag and therapist ids cannot be used together")

func NewUsecase(
	therapistRepo ports.TherapistRepository,
	timeSlotRepo ports.TimeSlotRepository,
	bookingRepo ports.BookingRepository,
	timeRangeMinimumDurationMinutes domain.DurationMinutes,
) *Usecase {
	return &Usecase{
		therapistRepo:                   therapistRepo,
		timeSlotRepo:                    timeSlotRepo,
		bookingRepo:                     bookingRepo,
		timeRangeMinimumDurationMinutes: timeRangeMinimumDurationMinutes,
	}
}

func (u *Usecase) Execute(input Input) ([]schedule.AvailableTimeRange, error) {
	// Validate input
	if input.SpecializationTag == "" && len(input.TherapistIDs) == 0 {
		return nil, ErrSpecializationTagOrTherapistIDsIsRequired
	}

	if input.SpecializationTag != "" && len(input.TherapistIDs) > 0 {
		return nil, ErrSpecializationTagAndTherapistIDsCannotBeUsedTogether
	}

	if input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}

	// Set default date range if not provided
	if input.StartDate.IsZero() {
		input.StartDate = time.Now().UTC()
	}

	if input.EndDate.IsZero() {
		input.EndDate = input.StartDate.AddDate(0, 0, 14) // Default to 2 weeks ahead
	}

	var therapists []*therapist.Therapist
	var err error

	if len(input.TherapistIDs) > 0 {
		therapists, err = u.therapistRepo.FindByIDs(input.TherapistIDs)
	} else {
		therapists, err = u.therapistRepo.FindBySpecializationAndLanguage(input.SpecializationTag, input.MustSpeakEnglish)
	}

	if err != nil {
		return nil, err
	}

	therapistIDs := make([]domain.TherapistID, len(therapists))
	for i, therapist := range therapists {
		therapistIDs[i] = therapist.ID
	}

	// For each therapist, calculate their available time ranges
	therapistTimeSlots := make(map[domain.TherapistID][]*timeslot.TimeSlot)
	therapistSlots, err := u.timeSlotRepo.BulkListByTherapist(therapistIDs)
	if err != nil {
		return nil, err
	}
	bookings, err := u.bookingRepo.BulkListByTherapistForDateRange(
		therapistIDs,
		[]booking.BookingState{booking.BookingStateConfirmed},
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return nil, err
	}

	allTherapistAvailabilities := []therapistAvailability{}
	nowUTC := domain.NewUTCTimestamp()
	for _, therapist := range therapists {
		// Get all time slots for this therapist
		timeSlots := therapistSlots[therapist.ID]
		therapistTimeSlots[therapist.ID] = timeSlots

		// Get confirmed bookings for this therapist in the date range
		// Convert bookings to a map for efficient lookup
		bookingMap := makeBookingMap(bookings[therapist.ID])

		// For each day in the date range
		for renderedSlotDay := input.StartDate; !renderedSlotDay.After(input.EndDate); renderedSlotDay = renderedSlotDay.AddDate(0, 0, 1) {
			availableDaySlots := filterAvailableDaySlots(timeSlots, renderedSlotDay, nowUTC)

			for _, slot := range availableDaySlots {
				// Get bookings for this slot on this day
				slotBookings := getBookingsForSlot(bookingMap, slot.ID, renderedSlotDay)

				therapistAvailabilities := findTherapistAvailabilities(
					therapist,
					slot,
					slotBookings,
					renderedSlotDay,
					u.timeRangeMinimumDurationMinutes,
				)
				allTherapistAvailabilities = append(allTherapistAvailabilities, therapistAvailabilities...)
			}
		}
	}

	// Step 2: Apply the line sweep algorithm to merge overlapping ranges
	return applyLineSweepAlgorithm(allTherapistAvailabilities, u.timeRangeMinimumDurationMinutes), nil
}

func findTherapistAvailabilities(
	therapist *therapist.Therapist,
	slot *timeslot.TimeSlot,
	slotBookings []*booking.Booking,
	renderedSlotDay time.Time,
	timeRangeMinimumDurationMinutes domain.DurationMinutes,
) []therapistAvailability {
	slotStart, slotEnd := slot.ApplyToDate(renderedSlotDay)

	// If no bookings, add the entire slot as available
	if len(slotBookings) == 0 {
		return []therapistAvailability{
			{
				TherapistID: therapist.ID,
				Therapist:   therapist,
				StartTime:   slotStart,
				EndTime:     slotEnd,
				TimeSlotID:  slot.ID,
			},
		}
	}

	slotTimeRange := timeRange{
		start: slotStart,
		end:   slotEnd,
	}

	// Convert bookings to time ranges
	bookingsTimeRanges := []timeRange{}
	for _, booking := range slotBookings {
		bookingsTimeRanges = append(bookingsTimeRanges, timeRange{
			start: booking.StartTime,
			end:   booking.StartTime.Add(time.Duration(booking.Duration) * time.Minute),
		})
	}

	// Calculate available ranges between bookings
	availableRanges := findInterBookingAvailabilities(
		slotTimeRange,
		slot.AfterSessionBreakTime,
		bookingsTimeRanges,
		timeRangeMinimumDurationMinutes,
	)

	therapistAvailabilities := []therapistAvailability{}
	// Add each available range
	for _, r := range availableRanges {
		therapistAvailabilities = append(therapistAvailabilities, therapistAvailability{
			TherapistID: therapist.ID,
			Therapist:   therapist,
			StartTime:   r.From,
			EndTime:     r.To,
			TimeSlotID:  slot.ID,
		})
	}

	return therapistAvailabilities
}

func filterAvailableDaySlots(
	timeSlots []*timeslot.TimeSlot,
	renderedSlotDay time.Time,
	nowUTC domain.UTCTimestamp,
) []*timeslot.TimeSlot {
	availableDaySlots := []*timeslot.TimeSlot{}
	// For each time slot on this day
	for _, slot := range timeSlots {
		if slot.DayOfWeek != timeslot.MapToDayOfWeek(renderedSlotDay.Weekday()) {
			continue
		}

		// If we're now past slot's pre-session buffer, skip.
		advanceNoticeDate := time.Duration(slot.AdvanceNotice) * time.Minute
		if nowUTC.Time().After(renderedSlotDay.Add(-1 * advanceNoticeDate)) {
			continue
		}

		// Calculate the specific time slot for this date
		_, slotEnd := slot.ApplyToDate(renderedSlotDay)

		// If the slot is in the past, skip it
		if slotEnd.Before(nowUTC) {
			continue
		}
		availableDaySlots = append(availableDaySlots, slot)
	}

	return availableDaySlots
}

// Step 1: Collect all therapist availabilities as individual time ranges
type therapistAvailability struct {
	TherapistID domain.TherapistID
	Therapist   *therapist.Therapist
	StartTime   domain.UTCTimestamp
	EndTime     domain.UTCTimestamp
	TimeSlotID  domain.TimeSlotID
}

// Helper functions for calculateAvailableTimeRanges
func makeBookingMap(bookings []*booking.Booking) map[string]map[domain.TimeSlotID][]*booking.Booking {
	bookingMap := make(map[string]map[domain.TimeSlotID][]*booking.Booking)
	for _, bookingEntry := range bookings {
		dateStr := time.Time(bookingEntry.StartTime).Format("2006-01-02")
		if bookingMap[dateStr] == nil {
			bookingMap[dateStr] = make(map[domain.TimeSlotID][]*booking.Booking)
		}
		bookingMap[dateStr][bookingEntry.TimeSlotID] = append(bookingMap[dateStr][bookingEntry.TimeSlotID], bookingEntry)
	}
	return bookingMap
}

func getBookingsForSlot(bookingMap map[string]map[domain.TimeSlotID][]*booking.Booking, slotID domain.TimeSlotID, date time.Time) []*booking.Booking {
	dateStr := date.Format("2006-01-02")
	if dateMap, ok := bookingMap[dateStr]; ok {
		return dateMap[slotID]
	}
	return nil
}

// applyLineSweepAlgorithm implements the line sweep algorithm to find all unique time ranges
// and the therapists available during each range
func applyLineSweepAlgorithm(
	availabilities []therapistAvailability,
	timeRangeMinimumDurationMinutes domain.DurationMinutes,
) []schedule.AvailableTimeRange {
	if len(availabilities) == 0 {
		return []schedule.AvailableTimeRange{}
	}

	type TherapistPointInfo struct {
		Therapist  *therapist.Therapist
		TimeSlotID domain.TimeSlotID
	}

	// Step 1: Collect all time points (start and end times)
	type TimePoint struct {
		Time          domain.UTCTimestamp
		IsStart       bool
		TherapistInfo TherapistPointInfo
	}

	timePoints := []TimePoint{}

	for _, avail := range availabilities {
		timePoints = append(timePoints, TimePoint{
			Time:    avail.StartTime,
			IsStart: true,
			TherapistInfo: TherapistPointInfo{
				Therapist:  avail.Therapist,
				TimeSlotID: avail.TimeSlotID,
			},
		})

		timePoints = append(timePoints, TimePoint{
			Time:    avail.EndTime,
			IsStart: false,
			TherapistInfo: TherapistPointInfo{
				Therapist:  avail.Therapist,
				TimeSlotID: avail.TimeSlotID,
			},
		})
	}

	// Step 2: Sort time points
	sort.Slice(timePoints, func(i, j int) bool {
		if timePoints[i].Time.Equal(timePoints[j].Time) {
			// If times are equal, prioritize end points before start points
			return !timePoints[i].IsStart && timePoints[j].IsStart
		}
		return timePoints[i].Time.Before(timePoints[j].Time)
	})

	// Step 3: Sweep through time points
	result := []schedule.AvailableTimeRange{}
	activeTherapists := map[domain.TherapistID]TherapistPointInfo{}
	var lastTime domain.UTCTimestamp

	for i, point := range timePoints {
		// Initialize lastTime with the first point's time
		if i == 0 {
			lastTime = point.Time
		}

		// If there are active therapists and time has advanced, create a range
		if len(activeTherapists) > 0 && !lastTime.Equal(point.Time) {
			// Convert active therapists to TherapistInfo
			therapistInfos := []schedule.TherapistInfo{}
			for _, t := range activeTherapists {
				therapistInfos = append(therapistInfos, schedule.TherapistInfo{
					TherapistID:     t.Therapist.ID,
					Name:            t.Therapist.Name,
					Specializations: t.Therapist.Specializations,
					SpeaksEnglish:   t.Therapist.SpeaksEnglish,
					TimeSlotID:      t.TimeSlotID,
				})
			}

			// Only add ranges with at least 60 minutes duration
			duration := int(point.Time.Sub(lastTime).Minutes())
			if duration >= int(timeRangeMinimumDurationMinutes) {

				// Sort therapists by name
				sort.Slice(therapistInfos, func(i, j int) bool {
					return therapistInfos[i].Name < therapistInfos[j].Name
				})

				result = append(result, schedule.AvailableTimeRange{
					From:       lastTime,
					To:         point.Time,
					Duration:   domain.DurationMinutes(duration),
					Therapists: therapistInfos,
				})
			}
		}

		// Update active therapists
		if point.IsStart {
			activeTherapists[point.TherapistInfo.Therapist.ID] = point.TherapistInfo
		} else {
			delete(activeTherapists, point.TherapistInfo.Therapist.ID)
		}

		lastTime = point.Time
	}

	return result
}

func findInterBookingAvailabilities(
	slot timeRange,
	afterSessionBreakTime domain.AfterSessionBreakTimeMinutes,
	bookings []timeRange,
	timeRangeMinimumDurationMinutes domain.DurationMinutes,
) []schedule.AvailableTimeRange {
	if len(bookings) == 0 {
		return []schedule.AvailableTimeRange{
			{
				From: slot.start,
				To:   slot.end,
			},
		}
	}

	bufferedBookings := []timeRange{}
	for _, booking := range bookings {
		bufferedBookings = append(bufferedBookings, timeRange{
			start: booking.start.Add(-time.Duration(afterSessionBreakTime) * time.Minute),
			end:   booking.end.Add(time.Duration(afterSessionBreakTime) * time.Minute),
		})
	}

	sortedBufferedBookings := sortTimeRangesByStartTime(bufferedBookings)
	lastEndTime := slot.start
	availableRanges := []schedule.AvailableTimeRange{}

	for _, booking := range sortedBufferedBookings {
		if lastEndTime.Before(booking.start) {
			duration := int(booking.start.Sub(lastEndTime).Minutes())
			if duration < int(timeRangeMinimumDurationMinutes) {
				continue
			}

			availableRanges = append(availableRanges, schedule.AvailableTimeRange{
				From: lastEndTime,
				To:   booking.start,
			})
		}
		lastEndTime = booking.end
	}

	// If there is a remaining time after the last booking, add it as an available range
	if lastEndTime.Before(slot.end) {
		duration := int(slot.end.Sub(lastEndTime).Minutes())
		if duration < int(timeRangeMinimumDurationMinutes) {
			return availableRanges
		}

		availableRanges = append(availableRanges, schedule.AvailableTimeRange{
			From: lastEndTime,
			To:   slot.end,
		})
	}

	return availableRanges
}

func sortTimeRangesByStartTime(bookings []timeRange) []timeRange {
	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].start.Before(bookings[j].start)
	})
	return bookings
}
