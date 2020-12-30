#import <CoreLocation/CoreLocation.h>
#import <Foundation/Foundation.h>

#import "location_manager_darwin.h"

@interface LocationManager : NSObject <CLLocationManagerDelegate>
{
    CLLocationManager *manager;
}

@property(readonly) NSInteger errorCode;

- (CLLocation *)getCurrentLocation;

@end


@implementation LocationManager

- (id)init {
    self = [super init];

    manager = [[CLLocationManager alloc] init];
    manager.delegate = self;
    manager.desiredAccuracy = kCLLocationAccuracyBest;

    return self;
}

-(void)dealloc {
    [manager release];
    [super dealloc];
}

- (CLLocation *)getCurrentLocation {
    [manager requestLocation];

    CFRunLoopRun();

    if (_errorCode != 0) {
        return nil;
    }

    CLLocation* location = manager.location;

    // negative horizontal accuracy means no location fix
    if (location.horizontalAccuracy < 0.0) {
        return nil;
    }

    return location;
}

- (void)locationManager:(CLLocationManager *)manager didUpdateLocations:(NSArray<CLLocation *> *)locations {
    CFRunLoopStop(CFRunLoopGetCurrent());
}

- (void)locationManager:(CLLocationManager *)manager didFailWithError:(NSError *)error {
    _errorCode = error.code;
    CFRunLoopStop(CFRunLoopGetCurrent());
}

@end

int get_current_location(Location *loc) {
    if (![CLLocationManager locationServicesEnabled]) {
        NSLog(@"location service disabled");
        return kCLErrorLocationUnknown;
    }

    @autoreleasepool {
        LocationManager *locationManager = [[LocationManager alloc] init];
        CLLocation *clloc = [locationManager getCurrentLocation];

        if (locationManager.errorCode != 0) {
            return locationManager.errorCode;
        }

        // CLLocationCoordinate2D coordinate = clloc.coordinate;
        // NSLog(@"latitude,logitude : %f, %f", coordinate.latitude, coordinate.longitude);
        // NSLog(@"timestamp         : %@", clloc.timestamp);

        loc->coordinate = clloc.coordinate;
        loc->altitude = clloc.altitude;
        loc->horizontalAccuracy = clloc.horizontalAccuracy;
        loc->verticalAccuracy = clloc.verticalAccuracy;
    }

    return 0;
}