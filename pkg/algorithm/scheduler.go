package algorithm

import (
	"sort"
	"time"

	"github.com/meeting-scheduler/internal/domain"
)

const (
	// Time slot scoring weights
	workingHoursWeight    = 1.0
	earlySlotWeight       = 0.8
	gapMinimizationWeight = 0.6
	bufferTimeWeight      = 0.4

	// Buffer time in minutes
	desiredBufferTime = 15

	// Working hours
	workDayStart = 9  // 9 AM
	workDayEnd   = 17 // 5 PM
)

type TimeSlot struct {
	Start time.Time
	End   time.Time
	Score float64
}

// FindOptimalSlot finds the best time slot for a meeting based on various criteria
func FindOptimalSlot(req domain.ScheduleRequest, events map[string][]domain.CalendarEvent) (*TimeSlot, error) {
	// Get all available slots
	availableSlots := findAvailableSlots(req, events)
	if len(availableSlots) == 0 {
		return nil, nil
	}

	// Score each slot
	scoredSlots := scoreSlots(availableSlots, events)

	// Sort by score (highest first)
	sort.Slice(scoredSlots, func(i, j int) bool {
		return scoredSlots[i].Score > scoredSlots[j].Score
	})

	return &scoredSlots[0], nil
}

// findAvailableSlots finds all possible time slots that work for all participants
func findAvailableSlots(req domain.ScheduleRequest, events map[string][]domain.CalendarEvent) []TimeSlot {
	var slots []TimeSlot
	current := req.TimeRange.Start

	for current.Before(req.TimeRange.End) {
		slotEnd := current.Add(time.Duration(req.DurationMinutes) * time.Minute)

		if slotEnd.Before(req.TimeRange.End) && isSlotAvailable(current, slotEnd, events) {
			slots = append(slots, TimeSlot{
				Start: current,
				End:   slotEnd,
			})
		}

		current = current.Add(15 * time.Minute)
	}

	return slots
}

// isSlotAvailable checks if a time slot is available for all participants
func isSlotAvailable(start, end time.Time, events map[string][]domain.CalendarEvent) bool {
	for _, userEvents := range events {
		for _, event := range userEvents {
			// Check for overlap
			if !(end.Before(event.StartTime) || start.After(event.EndTime)) {
				return false
			}
		}
	}
	return true
}

// scoreSlots scores each available slot based on our criteria
func scoreSlots(slots []TimeSlot, events map[string][]domain.CalendarEvent) []TimeSlot {
	for i := range slots {
		slots[i].Score = calculateSlotScore(slots[i], events)
	}
	return slots
}

// calculateSlotScore calculates a score for a time slot based on various criteria
func calculateSlotScore(slot TimeSlot, events map[string][]domain.CalendarEvent) float64 {
	var score float64

	score += workingHoursScore(slot) * workingHoursWeight

	score += earlySlotScore(slot) * earlySlotWeight

	score += gapMinimizationScore(slot, events) * gapMinimizationWeight

	score += bufferTimeScore(slot, events) * bufferTimeWeight

	return score
}

// workingHoursScore prefers slots during working hours
func workingHoursScore(slot TimeSlot) float64 {
	hour := slot.Start.Hour()

	if hour >= workDayStart && hour < workDayEnd {
		return 1.0
	}

	if hour >= workDayStart-1 && hour < workDayStart || hour >= workDayEnd && hour < workDayEnd+1 {
		return 0.5
	}

	return 0.0
}

func earlySlotScore(slot TimeSlot) float64 {
	hour := float64(slot.Start.Hour())

	if hour >= float64(workDayStart) && hour <= float64(workDayEnd) {
		return 1.0 - (hour-float64(workDayStart))/float64(workDayEnd-workDayStart)
	}

	return 0.0
}

func gapMinimizationScore(slot TimeSlot, events map[string][]domain.CalendarEvent) float64 {
	var totalScore float64
	count := 0

	for _, userEvents := range events {
		score := 1.0 // Default score for perfect back-to-back scheduling

		for _, event := range userEvents {
			if event.EndTime.Before(slot.Start) {
				gap := slot.Start.Sub(event.EndTime).Minutes()
				if gap < desiredBufferTime {
					score *= 0.5 // Penalize small gaps
				} else if gap > 60 {
					score *= 0.8 // Slightly penalize large gaps
				}
			}

			if event.StartTime.After(slot.End) {
				gap := event.StartTime.Sub(slot.End).Minutes()
				if gap < desiredBufferTime {
					score *= 0.5
				} else if gap > 60 {
					score *= 0.8
				}
			}
		}

		totalScore += score
		count++
	}

	if count == 0 {
		return 1.0
	}
	return totalScore / float64(count)
}

func bufferTimeScore(slot TimeSlot, events map[string][]domain.CalendarEvent) float64 {
	var totalScore float64
	count := 0

	for _, userEvents := range events {
		score := 1.0

		for _, event := range userEvents {
			if event.EndTime.Before(slot.Start) {
				bufferBefore := slot.Start.Sub(event.EndTime).Minutes()
				if bufferBefore < desiredBufferTime {
					score *= float64(bufferBefore) / float64(desiredBufferTime)
				}
			}

			if event.StartTime.After(slot.End) {
				bufferAfter := event.StartTime.Sub(slot.End).Minutes()
				if bufferAfter < desiredBufferTime {
					score *= float64(bufferAfter) / float64(desiredBufferTime)
				}
			}
		}

		totalScore += score
		count++
	}

	if count == 0 {
		return 1.0
	}
	return totalScore / float64(count)
}
