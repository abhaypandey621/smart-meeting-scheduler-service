package endpoint

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/meeting-scheduler/internal/domain"
	"github.com/meeting-scheduler/internal/service"
)

// Endpoints holds all Go kit endpoints for the scheduler service
type Endpoints struct {
	Schedule        endpoint.Endpoint
	GetUserCalendar endpoint.Endpoint
}

// MakeEndpoints creates the service endpoints
func MakeEndpoints(s service.SchedulerService) Endpoints {
	return Endpoints{
		Schedule:        makeScheduleEndpoint(s),
		GetUserCalendar: makeGetUserCalendarEndpoint(s),
	}
}

func makeScheduleEndpoint(s service.SchedulerService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(domain.ScheduleRequest)
		resp, err := s.Schedule(ctx, req)
		// If the error is ErrInvalidRequest, check for a more specific error from validation
		if err != nil && err.Error() != service.ErrInvalidRequest.Error() {
			return nil, err
		}
		return resp, err
	}
}

func makeGetUserCalendarEndpoint(s service.SchedulerService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetUserCalendarRequest)
		return s.GetUserCalendar(ctx, req.UserID, req.Start, req.End)
	}
}

type GetUserCalendarRequest struct {
	UserID string
	Start  time.Time
	End    time.Time
}
