package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	kitendpoint "github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/meeting-scheduler/internal/domain"
	"github.com/meeting-scheduler/internal/endpoint"
	"github.com/meeting-scheduler/internal/service"
)

// NewHTTPHandler returns an HTTP handler for the scheduler service
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/schedule").HandlerFunc(scheduleHandler(endpoints.Schedule, logger, options))

	r.Methods("GET").Path("/users/{userId}/calendar").Handler(httptransport.NewServer(
		endpoints.GetUserCalendar,
		decodeGetUserCalendarRequest,
		encodeResponse,
		options...,
	))

	return r
}

func decodeScheduleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req domain.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeGetUserCalendarRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return nil, err
	}

	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return nil, err
	}

	return endpoint.GetUserCalendarRequest{
		UserID: userID,
		Start:  startTime,
		End:    endTime,
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeScheduleResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func scheduleHandler(ep kitendpoint.Endpoint, logger log.Logger, options []httptransport.ServerOption) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server := httptransport.NewServer(
			ep,
			decodeScheduleRequest,
			encodeScheduleResponse,
			options...,
		)

		server.ServeHTTP(w, r)
	}
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err {
	case service.ErrInvalidRequest:
		w.WriteHeader(http.StatusBadRequest)
	case service.ErrNoAvailableSlot:
		w.WriteHeader(http.StatusConflict)
	case service.ErrUserNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}
