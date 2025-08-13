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

type Input struct {
	SpecializationTag string
	MustSpeakEnglish  bool
	TherapistIDs      []domain.TherapistID
	StartDate         time.Time
	EndDate           time.Time
}

type Usecase struct {
	therapistRepo ports.TherapistRepository
	timeSlotRepo  ports.TimeSlotRepository
	bookingRepo   ports.BookingRepository
}

var ErrSpecializationTagOrTherapistIDsIsRequired = errors.New("specialization tag or therapist ids is required")
var ErrInvalidDateRange = errors.New("invalid date range")
var ErrSpecializationTagAndTherapistIDsCannotBeUsedTogether = errors.New("specialization tag and therapist ids cannot be used together")

var timeRangeMinimumDurationMinutes = 60

func NewUsecase(
	therapistRepo ports.TherapistRepository,
	timeSlotRepo ports.TimeSlotRepository,
	bookingRepo ports.BookingRepository,
) *Usecase {
	return &Usecase{
		therapistRepo: therapistRepo,
		timeSlotRepo:  timeSlotRepo,
		bookingRepo:   bookingRepo,
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
	therapistBookings := make(map[domain.TherapistID][]*booking.Booking)

	therapistSlots, err := u.timeSlotRepo.BulkListByTherapist(therapistIDs)
	if err != nil {
		return nil, err
	}
	bookings, err := u.bookingRepo.BulkListByTherapistForDateRange(
		therapistIDs,
		booking.BookingStateConfirmed,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return nil, err
	}

	for _, therapist := range therapists {
		// Get all time slots for this therapist
		timeSlots := therapistSlots[therapist.ID]
		therapistTimeSlots[therapist.ID] = timeSlots

		// Get confirmed bookings for this therapist in the date range
		therapistBookings[therapist.ID] = bookings[therapist.ID]
	}

	// Calculate available time ranges using the line sweep algorithm
	response := u.calculateAvailableTimeRanges(
		therapists,
		therapistTimeSlots,
		therapistBookings,
		input.StartDate,
		input.EndDate,
	)

	return response, nil
}

// Step 1: Collect all therapist availabilities as individual time ranges
type therapistAvailability struct {
	TherapistID domain.TherapistID
	Therapist   *therapist.Therapist
	StartTime   domain.UTCTimestamp
	EndTime     domain.UTCTimestamp
	TimeSlotID  domain.TimeSlotID
}

// calculateAvailableTimeRanges calculates available time ranges using a line sweep algorithm
func (u *Usecase) calculateAvailableTimeRanges(
	therapists []*therapist.Therapist,
	therapistTimeSlots map[domain.TherapistID][]*timeslot.TimeSlot,
	therapistBookings map[domain.TherapistID][]*booking.Booking,
	startDate, endDate time.Time,
) []schedule.AvailableTimeRange {

	allAvailabilities := []therapistAvailability{}
	nowUTC := domain.NewUTCTimestamp()
	// For each therapist, calculate their available time ranges
	for _, therapist := range therapists {
		therapistID := therapist.ID
		timeSlots := therapistTimeSlots[therapistID]
		bookings := therapistBookings[therapistID]

		// Convert bookings to a map for efficient lookup
		bookingMap := makeBookingMap(bookings)

		// For each day in the date range
		for renderedSlotDay := startDate; !renderedSlotDay.After(endDate); renderedSlotDay = renderedSlotDay.AddDate(0, 0, 1) {

			// For each time slot on this day
			for _, slot := range timeSlots {

				if slot.DayOfWeek != timeslot.MapToDayOfWeek(renderedSlotDay.Weekday()) {
					continue
				}

				// If we're now past slot's pre-session buffer, skip.
				preSessionBuffer := time.Duration(slot.PreSessionBuffer) * time.Minute
				if nowUTC.Time().After(renderedSlotDay.Add(-1 * preSessionBuffer)) {
					continue
				}

				// Calculate the specific time slot for this date
				slotStart, slotEnd := slot.ApplyToDate(renderedSlotDay)

				// If the slot is in the past, skip it
				if slotEnd.Before(nowUTC) {
					continue
				}

				// Get bookings for this slot on this day
				dayBookings := getBookingsForSlot(bookingMap, slot.ID, renderedSlotDay)

				// If no bookings, add the entire slot as available
				if len(dayBookings) == 0 {
					allAvailabilities = append(allAvailabilities, therapistAvailability{
						TherapistID: therapistID,
						Therapist:   therapist,
						StartTime:   slotStart,
						EndTime:     slotEnd,
						TimeSlotID:  slot.ID,
					})
					continue
				}

				// Calculate available ranges between bookings
				availableRanges := calculateAvailableRangesBetweenBookings(
					slotStart, slotEnd, dayBookings, slot.PreSessionBuffer, slot.PostSessionBuffer)

				// Add each available range
				for _, r := range availableRanges {
					allAvailabilities = append(allAvailabilities, therapistAvailability{
						TherapistID: r.TherapistID,
						Therapist:   therapist,
						StartTime:   r.StartTime,
						EndTime:     r.EndTime,
						TimeSlotID:  slot.ID,
					})
				}
			}
		}
	}

	// Step 2: Apply the line sweep algorithm to merge overlapping ranges
	return applyLineSweepAlgorithm(allAvailabilities)
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

type calculatedRange struct {
	StartTime   domain.UTCTimestamp
	EndTime     domain.UTCTimestamp
	TherapistID domain.TherapistID
}

func calculateAvailableRangesBetweenBookings(
	slotStart, slotEnd domain.UTCTimestamp,
	bookings []*booking.Booking,
	preBuffer, postBuffer domain.DurationMinutes,
) []calculatedRange {
	availableRanges := []calculatedRange{}
	sortBookingsByStartTime(bookings)

	lastEndTime := slotStart

	for _, booking := range bookings {
		bookingStartTime := domain.UTCTimestamp(booking.StartTime)
		bookingEndTime := bookingStartTime.Add(time.Duration(booking.Duration) * time.Minute)

		// FIXME: This is wrong. This is only used for eliminating timeslots.
		bufferedStartTime := bookingStartTime.Add(-time.Duration(preBuffer) * time.Minute)
		bufferedEndTime := bookingEndTime.Add(time.Duration(postBuffer) * time.Minute)

		// Ensure buffered times are within the slot boundaries
		if bufferedStartTime.Before(slotStart) {
			bufferedStartTime = slotStart
		}
		if bufferedEndTime.After(slotEnd) {
			bufferedEndTime = slotEnd
		}

		if lastEndTime.Before(bufferedStartTime) {
			duration := int(bufferedStartTime.Sub(lastEndTime).Minutes())
			if duration >= timeRangeMinimumDurationMinutes { // Minimum 30 minutes for an available slot
				availableRanges = append(availableRanges, calculatedRange{
					StartTime:   lastEndTime,
					EndTime:     bufferedStartTime,
					TherapistID: booking.TherapistID, // This is the therapist for the original slot
				})
			}
		}
		lastEndTime = bufferedEndTime
	}

	if lastEndTime.Before(slotEnd) {
		duration := int(slotEnd.Sub(lastEndTime).Minutes())
		if duration >= 30 {
			availableRanges = append(availableRanges, calculatedRange{
				StartTime:   lastEndTime,
				EndTime:     slotEnd,
				TherapistID: bookings[0].TherapistID, // Assuming all bookings in this slot are for the same therapist
			})
		}
	}

	return availableRanges
}

func sortBookingsByStartTime(bookings []*booking.Booking) {
	sort.Slice(bookings, func(i, j int) bool {
		return time.Time(bookings[i].StartTime).Before(time.Time(bookings[j].StartTime))
	})
}

// applyLineSweepAlgorithm implements the line sweep algorithm to find all unique time ranges
// and the therapists available during each range
func applyLineSweepAlgorithm(availabilities []therapistAvailability) []schedule.AvailableTimeRange {
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
			if duration >= timeRangeMinimumDurationMinutes {

				// Sort therapists by name
				sort.Slice(therapistInfos, func(i, j int) bool {
					return therapistInfos[i].Name < therapistInfos[j].Name
				})

				result = append(result, schedule.AvailableTimeRange{
					From:            lastTime,
					To:              point.Time,
					DurationMinutes: duration,
					Therapists:      therapistInfos,
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
