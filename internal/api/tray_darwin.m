#import <Cocoa/Cocoa.h>

static NSStatusItem *statusItem = nil;

void TraySet(const char* tooltip) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem == nil) {
            statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        }
        statusItem.button.title = @"LS";
        if (tooltip && strlen(tooltip) > 0) {
            statusItem.button.toolTip = [NSString stringWithUTF8String:tooltip];
        }
    });
}

void TrayRemove() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (statusItem != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
            statusItem = nil;
        }
    });
}
