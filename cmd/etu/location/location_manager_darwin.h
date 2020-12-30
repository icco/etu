#import <CoreLocation/CoreLocation.h>

typedef struct _Location {
	CLLocationCoordinate2D coordinate;
	double altitude;
	double horizontalAccuracy;
	double verticalAccuracy;
	//NSTimeInterval timestamp;
} Location;

//Location *get_current_location();
int get_current_location(Location *loc);
