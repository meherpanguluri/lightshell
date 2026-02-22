#import <Cocoa/Cocoa.h>

static NSStatusItem *statusItem = nil;
static NSStatusItem *devStatusItem = nil;

extern void goTrayMenuAction(const char* itemId);

// --- TrayMenuTarget: handles dev menu item clicks ---
@interface TrayMenuTarget : NSObject
- (void)debugToggle:(id)sender;
- (void)appQuit:(id)sender;
@end

@implementation TrayMenuTarget
- (void)debugToggle:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    item.state = (item.state == NSControlStateValueOn) ? NSControlStateValueOff : NSControlStateValueOn;
    goTrayMenuAction("debug.toggle");
}
- (void)appQuit:(id)sender {
    goTrayMenuAction("app.quit");
}
@end

static TrayMenuTarget *menuTarget = nil;

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

void TraySetDevMenu() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (menuTarget == nil) {
            menuTarget = [[TrayMenuTarget alloc] init];
        }

        if (devStatusItem == nil) {
            devStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        }
        devStatusItem.button.title = @"LS";
        devStatusItem.button.toolTip = @"LightShell Dev";

        NSMenu *menu = [[NSMenu alloc] initWithTitle:@"LightShell"];

        NSMenuItem *debugItem = [[NSMenuItem alloc] initWithTitle:@"Debug Console"
                                                           action:@selector(debugToggle:)
                                                    keyEquivalent:@""];
        [debugItem setTarget:menuTarget];
        [menu addItem:debugItem];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                          action:@selector(appQuit:)
                                                   keyEquivalent:@"q"];
        [quitItem setTarget:menuTarget];
        [menu addItem:quitItem];

        devStatusItem.menu = menu;
    });
}
