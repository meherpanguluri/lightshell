#import <Cocoa/Cocoa.h>
#import <UserNotifications/UserNotifications.h>

void NotifySend(const char* title, const char* body) {
    UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];

    // Request permission (idempotent — only shows prompt once)
    [center requestAuthorizationWithOptions:(UNAuthorizationOptionAlert | UNAuthorizationOptionSound)
                          completionHandler:^(BOOL granted, NSError *error) {
        if (!granted) return;

        UNMutableNotificationContent *content = [[UNMutableNotificationContent alloc] init];
        content.title = [NSString stringWithUTF8String:title];
        content.body = [NSString stringWithUTF8String:body];
        content.sound = [UNNotificationSound defaultSound];

        NSString *identifier = [[NSUUID UUID] UUIDString];
        UNNotificationRequest *request = [UNNotificationRequest requestWithIdentifier:identifier
                                                                              content:content
                                                                              trigger:nil];
        [center addNotificationRequest:request withCompletionHandler:nil];
    }];
}
