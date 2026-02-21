#import <Cocoa/Cocoa.h>

const char* DialogOpen(const char* title, const char* defaultPath, int directory, int multiple) {
    __block const char* result = NULL;

    dispatch_sync(dispatch_get_main_queue(), ^{
        NSOpenPanel *panel = [NSOpenPanel openPanel];

        if (title && strlen(title) > 0) {
            [panel setTitle:[NSString stringWithUTF8String:title]];
        }
        if (defaultPath && strlen(defaultPath) > 0) {
            [panel setDirectoryURL:[NSURL fileURLWithPath:[NSString stringWithUTF8String:defaultPath]]];
        }

        [panel setCanChooseFiles:!directory];
        [panel setCanChooseDirectories:(directory != 0)];
        [panel setAllowsMultipleSelection:(multiple != 0)];

        if ([panel runModal] == NSModalResponseOK) {
            NSMutableArray *paths = [NSMutableArray array];
            for (NSURL *url in [panel URLs]) {
                [paths addObject:[url path]];
            }
            NSString *joined = [paths componentsJoinedByString:@"\n"];
            result = strdup([joined UTF8String]);
        }
    });

    return result;
}

const char* DialogSave(const char* title, const char* defaultPath) {
    __block const char* result = NULL;

    dispatch_sync(dispatch_get_main_queue(), ^{
        NSSavePanel *panel = [NSSavePanel savePanel];

        if (title && strlen(title) > 0) {
            [panel setTitle:[NSString stringWithUTF8String:title]];
        }
        if (defaultPath && strlen(defaultPath) > 0) {
            NSString *path = [NSString stringWithUTF8String:defaultPath];
            [panel setDirectoryURL:[NSURL fileURLWithPath:[path stringByDeletingLastPathComponent]]];
            [panel setNameFieldStringValue:[path lastPathComponent]];
        }

        if ([panel runModal] == NSModalResponseOK) {
            result = strdup([[[panel URL] path] UTF8String]);
        }
    });

    return result;
}

void DialogMessage(const char* title, const char* message) {
    dispatch_sync(dispatch_get_main_queue(), ^{
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:[NSString stringWithUTF8String:title]];
        [alert setInformativeText:[NSString stringWithUTF8String:message]];
        [alert setAlertStyle:NSAlertStyleInformational];
        [alert addButtonWithTitle:@"OK"];
        [alert runModal];
    });
}

int DialogConfirm(const char* title, const char* message) {
    __block int result = 0;

    dispatch_sync(dispatch_get_main_queue(), ^{
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:[NSString stringWithUTF8String:title]];
        [alert setInformativeText:[NSString stringWithUTF8String:message]];
        [alert setAlertStyle:NSAlertStyleWarning];
        [alert addButtonWithTitle:@"OK"];
        [alert addButtonWithTitle:@"Cancel"];
        result = ([alert runModal] == NSAlertFirstButtonReturn) ? 1 : 0;
    });

    return result;
}

const char* DialogPrompt(const char* title, const char* defaultValue) {
    __block const char* result = NULL;

    dispatch_sync(dispatch_get_main_queue(), ^{
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:[NSString stringWithUTF8String:title]];
        [alert setAlertStyle:NSAlertStyleInformational];
        [alert addButtonWithTitle:@"OK"];
        [alert addButtonWithTitle:@"Cancel"];

        NSTextField *input = [[NSTextField alloc] initWithFrame:NSMakeRect(0, 0, 300, 24)];
        [input setStringValue:[NSString stringWithUTF8String:defaultValue]];
        [alert setAccessoryView:input];

        if ([alert runModal] == NSAlertFirstButtonReturn) {
            result = strdup([[input stringValue] UTF8String]);
        }
    });

    return result;
}
