#import <Cocoa/Cocoa.h>

void ShellOpen(const char* url) {
    NSString *nsURL = [NSString stringWithUTF8String:url];
    NSURL *nsurl = [NSURL URLWithString:nsURL];
    [[NSWorkspace sharedWorkspace] openURL:nsurl];
}
