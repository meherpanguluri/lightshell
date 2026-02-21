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
