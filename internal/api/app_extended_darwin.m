#import <Cocoa/Cocoa.h>

void AppSetBadgeCount(int count) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (count > 0) {
            [[NSApp dockTile] setBadgeLabel:[NSString stringWithFormat:@"%d", count]];
        } else {
            [[NSApp dockTile] setBadgeLabel:@""];
        }
    });
}
