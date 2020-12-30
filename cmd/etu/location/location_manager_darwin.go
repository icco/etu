package location

import (
	"fmt"
	"time"
)

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=10.14
#cgo LDFLAGS: -framework CoreLocation -framework Foundation -mmacosx-version-min=10.14

#import "location_manager_darwin.h"
*/
import "C"

type Location struct {
	Coordinate         Coordinate2D
	Altitude           float64
	HorizontalAccuracy float64
	VerticalAccuracy   float64
	Timestamp          time.Time
}

type Coordinate2D struct {
	Latitude  float64
	Longitude float64
}

func CurrentLocation() (Location, error) {
	var cloc C.Location
	if ret := C.get_current_location(&cloc); int(ret) != 0 {
		return Location{}, fmt.Errorf("failed to get location, code %d", ret)
	}

	loc := Location{
		Coordinate: Coordinate2D{
			Latitude:  float64(C.float(cloc.coordinate.latitude)),
			Longitude: float64(C.float(cloc.coordinate.longitude)),
		},
		Altitude:           float64(C.float(cloc.altitude)),
		HorizontalAccuracy: float64(C.float(cloc.horizontalAccuracy)),
		VerticalAccuracy:   float64(C.float(cloc.verticalAccuracy)),
	}
	return loc, nil
}
