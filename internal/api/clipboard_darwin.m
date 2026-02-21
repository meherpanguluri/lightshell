#import <Cocoa/Cocoa.h>

const char* ClipboardRead() {
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    NSString *text = [pb stringForType:NSPasteboardTypeString];
    if (text == nil) return NULL;
    return strdup([text UTF8String]);
}

void ClipboardWrite(const char* text) {
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    [pb clearContents];
    [pb setString:[NSString stringWithUTF8String:text] forType:NSPasteboardTypeString];
}
