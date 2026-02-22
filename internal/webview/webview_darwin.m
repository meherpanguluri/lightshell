#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

// Forward declaration of Go callback
extern void goMessageHandler(const char* msg);

static NSWindow *mainWindow = nil;
static WKWebView *webView = nil;
static NSApplication *app = nil;

// Script message handler — receives postMessage from JS
@interface MessageHandler : NSObject <WKScriptMessageHandler>
@end

@implementation MessageHandler
- (void)userContentController:(WKUserContentController *)controller
      didReceiveScriptMessage:(WKScriptMessage *)message {
    if ([message.body isKindOfClass:[NSString class]]) {
        const char *msg = [message.body UTF8String];
        goMessageHandler(msg);
    }
}
@end

// Window delegate — handles close, resize, move, focus events
@interface WindowDelegate : NSObject <NSWindowDelegate>
@end

@implementation WindowDelegate
- (BOOL)windowShouldClose:(NSWindow *)sender {
    [NSApp terminate:nil];
    return YES;
}
@end

static MessageHandler *msgHandler = nil;
static WindowDelegate *winDelegate = nil;

void WebviewCreate(const char* title, int width, int height, int minWidth, int minHeight,
    int resizable, int frameless, int alwaysOnTop, int transparent, int devTools) {

    app = [NSApplication sharedApplication];
    [app setActivationPolicy:NSApplicationActivationPolicyRegular];

    // Window style mask
    NSWindowStyleMask styleMask = NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable;
    if (resizable) {
        styleMask |= NSWindowStyleMaskResizable;
    }
    if (frameless) {
        styleMask = NSWindowStyleMaskBorderless;
        if (resizable) {
            styleMask |= NSWindowStyleMaskResizable;
        }
    }

    NSRect frame = NSMakeRect(0, 0, width, height);
    mainWindow = [[NSWindow alloc] initWithContentRect:frame
                                             styleMask:styleMask
                                               backing:NSBackingStoreBuffered
                                                 defer:NO];

    NSString *nsTitle = [NSString stringWithUTF8String:title];
    [mainWindow setTitle:nsTitle];
    [mainWindow center];

    if (minWidth > 0 && minHeight > 0) {
        [mainWindow setMinSize:NSMakeSize(minWidth, minHeight)];
    }

    if (alwaysOnTop) {
        [mainWindow setLevel:NSFloatingWindowLevel];
    }

    if (transparent) {
        [mainWindow setOpaque:NO];
        [mainWindow setBackgroundColor:[NSColor clearColor]];
    }

    winDelegate = [[WindowDelegate alloc] init];
    [mainWindow setDelegate:winDelegate];

    // WKWebView configuration
    WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
    WKUserContentController *contentController = [[WKUserContentController alloc] init];

    msgHandler = [[MessageHandler alloc] init];
    [contentController addScriptMessageHandler:msgHandler name:@"lightshell"];
    config.userContentController = contentController;

    // Enable DevTools in dev mode
    if (devTools) {
        WKPreferences *prefs = config.preferences;
        [prefs setValue:@YES forKey:@"developerExtrasEnabled"];
    }

    webView = [[WKWebView alloc] initWithFrame:[mainWindow.contentView bounds] configuration:config];
    [webView setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];

    if (transparent) {
        [webView setValue:@NO forKey:@"drawsBackground"];
    }

    [mainWindow.contentView addSubview:webView];
    [mainWindow makeKeyAndOrderFront:nil];
    [app activateIgnoringOtherApps:YES];
}

void WebviewLoadHTML(const char* html) {
    if (webView) {
        NSString *nsHTML = [NSString stringWithUTF8String:html];
        [webView loadHTMLString:nsHTML baseURL:nil];
    }
}

void WebviewLoadURL(const char* url) {
    if (webView) {
        NSString *nsURL = [NSString stringWithUTF8String:url];
        NSURL *nsurl = [NSURL URLWithString:nsURL];
        if ([nsurl.scheme isEqualToString:@"file"]) {
            // For file URLs, allow read access to the parent directory
            NSURL *dirURL = [nsurl URLByDeletingLastPathComponent];
            [webView loadFileURL:nsurl allowingReadAccessToURL:dirURL];
        } else {
            NSURLRequest *request = [NSURLRequest requestWithURL:nsurl];
            [webView loadRequest:request];
        }
    }
}

void WebviewEval(const char* js) {
    if (webView) {
        NSString *nsJS = [NSString stringWithUTF8String:js];
        dispatch_async(dispatch_get_main_queue(), ^{
            [webView evaluateJavaScript:nsJS completionHandler:nil];
        });
    }
}

void WebviewAddUserScript(const char* js) {
    if (webView) {
        NSString *nsJS = [NSString stringWithUTF8String:js];
        WKUserScript *script = [[WKUserScript alloc] initWithSource:nsJS
            injectionTime:WKUserScriptInjectionTimeAtDocumentStart
            forMainFrameOnly:YES];
        [webView.configuration.userContentController addUserScript:script];
    }
}

void WebviewSetTitle(const char* title) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            NSString *nsTitle = [NSString stringWithUTF8String:title];
            [mainWindow setTitle:nsTitle];
        });
    }
}

void WebviewSetSize(int width, int height) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            NSRect frame = [mainWindow frame];
            frame.size = NSMakeSize(width, height);
            [mainWindow setFrame:frame display:YES animate:YES];
        });
    }
}

void WebviewSetMinSize(int width, int height) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [mainWindow setMinSize:NSMakeSize(width, height)];
        });
    }
}

void WebviewSetMaxSize(int width, int height) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [mainWindow setMaxSize:NSMakeSize(width, height)];
        });
    }
}

void WebviewSetPosition(int x, int y) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            NSScreen *screen = [NSScreen mainScreen];
            CGFloat screenHeight = screen.frame.size.height;
            CGFloat windowHeight = mainWindow.frame.size.height;
            // Convert from top-left origin (web convention) to bottom-left (macOS)
            NSPoint origin = NSMakePoint(x, screenHeight - y - windowHeight);
            [mainWindow setFrameOrigin:origin];
        });
    }
}

void WebviewFullscreen(void) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            if (!([mainWindow styleMask] & NSWindowStyleMaskFullScreen)) {
                [mainWindow toggleFullScreen:nil];
            }
        });
    }
}

void WebviewMinimize(void) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [mainWindow miniaturize:nil];
        });
    }
}

void WebviewMaximize(void) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [mainWindow zoom:nil];
        });
    }
}

void WebviewRestore(void) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            if ([mainWindow styleMask] & NSWindowStyleMaskFullScreen) {
                [mainWindow toggleFullScreen:nil];
            }
            if ([mainWindow isMiniaturized]) {
                [mainWindow deminiaturize:nil];
            }
        });
    }
}

void WebviewClose(void) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [mainWindow close];
        });
    }
}

void WebviewRun(void) {
    [app run];
}

void WebviewDestroy(void) {
    if (webView) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [webView removeFromSuperview];
            webView = nil;
            mainWindow = nil;
        });
    }
}

void WebviewSetContentProtection(int enabled) {
    if (mainWindow) {
        dispatch_async(dispatch_get_main_queue(), ^{
            if (enabled) {
                mainWindow.sharingType = NSWindowSharingNone;
            } else {
                mainWindow.sharingType = NSWindowSharingReadOnly;
            }
        });
    }
}

void WebviewSetVibrancy(const char* style) {
    if (mainWindow && webView) {
        dispatch_async(dispatch_get_main_queue(), ^{
            NSString *nsStyle = [NSString stringWithUTF8String:style];

            // Remove any existing vibrancy view
            for (NSView *subview in [mainWindow.contentView.subviews copy]) {
                if ([subview isKindOfClass:[NSVisualEffectView class]]) {
                    [subview removeFromSuperview];
                }
            }

            NSVisualEffectMaterial material;
            if ([nsStyle isEqualToString:@"sidebar"]) {
                material = NSVisualEffectMaterialSidebar;
            } else if ([nsStyle isEqualToString:@"header"]) {
                material = NSVisualEffectMaterialHeaderView;
            } else if ([nsStyle isEqualToString:@"content"]) {
                material = NSVisualEffectMaterialContentBackground;
            } else if ([nsStyle isEqualToString:@"sheet"]) {
                material = NSVisualEffectMaterialSheet;
            } else {
                return;
            }

            NSVisualEffectView *effectView = [[NSVisualEffectView alloc] initWithFrame:mainWindow.contentView.bounds];
            effectView.material = material;
            effectView.blendingMode = NSVisualEffectBlendingModeBehindWindow;
            effectView.state = NSVisualEffectStateActive;
            effectView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;

            // Make the webview transparent so vibrancy shows through
            [webView setValue:@NO forKey:@"drawsBackground"];

            // Insert vibrancy view behind the webview
            [mainWindow.contentView addSubview:effectView positioned:NSWindowBelow relativeTo:webView];
        });
    }
}

void WebviewSetColorScheme(const char* scheme) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSString *nsScheme = [NSString stringWithUTF8String:scheme];
        if ([nsScheme isEqualToString:@"light"]) {
            NSApp.appearance = [NSAppearance appearanceNamed:NSAppearanceNameAqua];
        } else if ([nsScheme isEqualToString:@"dark"]) {
            NSApp.appearance = [NSAppearance appearanceNamed:NSAppearanceNameDarkAqua];
        } else {
            NSApp.appearance = nil;
        }
    });
}

void WebviewEnableFileDrop(void) {
    if (webView) {
        dispatch_async(dispatch_get_main_queue(), ^{
            [webView registerForDraggedTypes:@[NSPasteboardTypeFileURL]];
        });
    }
}

int WebviewGetWidth(void) {
    if (mainWindow) {
        return (int)mainWindow.frame.size.width;
    }
    return 0;
}

int WebviewGetHeight(void) {
    if (mainWindow) {
        return (int)mainWindow.frame.size.height;
    }
    return 0;
}

int WebviewGetX(void) {
    if (mainWindow) {
        return (int)mainWindow.frame.origin.x;
    }
    return 0;
}

int WebviewGetY(void) {
    if (mainWindow) {
        // Convert from macOS bottom-left origin to top-left origin
        NSScreen *screen = [NSScreen mainScreen];
        CGFloat screenHeight = screen.frame.size.height;
        CGFloat windowHeight = mainWindow.frame.size.height;
        return (int)(screenHeight - mainWindow.frame.origin.y - windowHeight);
    }
    return 0;
}

// Returns PNG data as a malloc'd buffer. Caller must free. Sets *outLen to data length.
// Returns NULL on failure.
void* WebviewScreenshot(int* outLen) {
    __block NSData *pngData = nil;

    dispatch_semaphore_t sem = dispatch_semaphore_create(0);

    dispatch_async(dispatch_get_main_queue(), ^{
        if (webView == nil) {
            dispatch_semaphore_signal(sem);
            return;
        }
        WKSnapshotConfiguration *config = [[WKSnapshotConfiguration alloc] init];
        [webView takeSnapshotWithConfiguration:config completionHandler:^(NSImage *image, NSError *error) {
            if (image && !error) {
                CGImageRef cgRef = [image CGImageForProposedRect:NULL context:nil hints:nil];
                if (cgRef) {
                    NSBitmapImageRep *rep = [[NSBitmapImageRep alloc] initWithCGImage:cgRef];
                    pngData = [rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
                    [pngData retain];
                }
            }
            dispatch_semaphore_signal(sem);
        }];
    });

    // Wait up to 5 seconds for the screenshot to complete
    dispatch_semaphore_wait(sem, dispatch_time(DISPATCH_TIME_NOW, 5 * NSEC_PER_SEC));

    if (pngData == nil) {
        *outLen = 0;
        return NULL;
    }

    *outLen = (int)[pngData length];
    void *buf = malloc(*outLen);
    memcpy(buf, [pngData bytes], *outLen);
    [pngData release];
    return buf;
}
