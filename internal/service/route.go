package service

import (
	"errors"
	"math"
	"time"
)

var (
	ErrRouteInsufficientPoints = errors.New("route must contain at least 2 points")
)

type RoutePoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RouteCalculationResult struct {
	DistanceKm      float64 `json:"distance_km"`
	DurationMinutes int     `json:"duration_minutes"`
	SuggestedStart  string  `json:"suggested_start,omitempty"`
	SuggestedEnd    string  `json:"suggested_end,omitempty"`
}

type RouteService struct {
	avgSpeed float64
}

func NewRouteService(avgSpeed float64) *RouteService {
	if avgSpeed <= 0 {
		avgSpeed = 35 // km/h default
	}
	return &RouteService{avgSpeed: avgSpeed}
}

func (s *RouteService) Calculate(points []RoutePoint) (RouteCalculationResult, error) {
	if len(points) < 2 {
		return RouteCalculationResult{}, ErrRouteInsufficientPoints
	}
	var distance float64
	for i := 1; i < len(points); i++ {
		distance += haversine(points[i-1], points[i])
	}
	durationHours := distance / s.avgSpeed
	durationMinutes := int(math.Ceil(durationHours * 60))
	start := time.Now().Add(time.Hour).Truncate(time.Minute)
	end := start.Add(time.Duration(durationMinutes) * time.Minute)
	return RouteCalculationResult{
		DistanceKm:      math.Round(distance*10) / 10,
		DurationMinutes: durationMinutes,
		SuggestedStart:  start.Format("15:04"),
		SuggestedEnd:    end.Format("15:04"),
	}, nil
}

func haversine(p1, p2 RoutePoint) float64 {
	const earthRadius = 6371.0
	dLat := degreesToRadians(p2.Latitude - p1.Latitude)
	dLon := degreesToRadians(p2.Longitude - p1.Longitude)
	lat1 := degreesToRadians(p1.Latitude)
	lat2 := degreesToRadians(p2.Latitude)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
