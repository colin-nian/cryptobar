#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#import "settings.h"

void settingsChanged(const char *jsonString);

static NSWindow *_settingsWindow = nil;

@interface SettingsWindowDelegate : NSObject <NSWindowDelegate, NSTabViewDelegate>
@property(nonatomic, strong) NSTabView *tabView;
@property(nonatomic, strong) NSDictionary *currentSettings;
@property(nonatomic, strong) NSDictionary *strings;

// Tab 1
@property(nonatomic, strong) NSPopUpButton *fontSizePopup;
@property(nonatomic, strong) NSPopUpButton *iconModePopup;
@property(nonatomic, strong) NSPopUpButton *logoColorPopup;
@property(nonatomic, strong) NSPopUpButton *languagePopup;
@property(nonatomic, strong) NSButton *showChangeSwitch;
@property(nonatomic, strong) NSButton *compactNameSwitch;

// Tab 2
@property(nonatomic, strong) NSTextField *bnWSField;
@property(nonatomic, strong) NSTextField *bnAPIField;
@property(nonatomic, strong) NSTextField *bnStatusLabel;
@property(nonatomic, strong) NSTextField *htxWSField;
@property(nonatomic, strong) NSTextField *htxAPIField;
@property(nonatomic, strong) NSTextField *htxStatusLabel;
@property(nonatomic, strong) NSTextField *gateWSField;
@property(nonatomic, strong) NSTextField *gateAPIField;
@property(nonatomic, strong) NSTextField *gateStatusLabel;
@property(nonatomic, strong) NSMatrix *sourceRadio;

// Tab 3
@property(nonatomic, strong) NSTextField *logPathLabel;

// Tab 4
@property(nonatomic, strong) NSTextField *usdtField;
@end

static NSArray *_langCodes = nil;
static NSArray *_langNames = nil;

@implementation SettingsWindowDelegate

+ (void)initialize {
    if (self == [SettingsWindowDelegate class]) {
        _langCodes = @[@"en", @"zh-CN", @"zh-TW", @"ja", @"ko"];
        _langNames = @[@"English", @"简体中文", @"繁體中文", @"日本語", @"한국어"];
    }
}

- (void)setupWithJSON:(NSDictionary *)settings {
    self.currentSettings = settings;
    self.strings = settings[@"strings"];
    if (!self.strings) self.strings = @{};
}

- (NSString *)S:(NSString *)key {
    NSString *val = self.strings[key];
    return val ? val : key;
}

#pragma mark - Tab 1: Personalization

- (NSView *)createPersonalizationTab {
    NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 480, 360)];

    CGFloat y = 300;
    CGFloat labelW = 160;
    CGFloat controlX = 170;

    // Language
    NSTextField *langLabel = [NSTextField labelWithString:[self S:@"language"]];
    langLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:langLabel];

    self.languagePopup = [[NSPopUpButton alloc] initWithFrame:NSMakeRect(controlX, y, 200, 26) pullsDown:NO];
    [self.languagePopup addItemsWithTitles:_langNames];
    NSString *currentLang = self.currentSettings[@"language"];
    NSInteger langIdx = [_langCodes indexOfObject:currentLang ?: @"en"];
    if (langIdx == NSNotFound) langIdx = 0;
    [self.languagePopup selectItemAtIndex:langIdx];
    [self.languagePopup setTarget:self];
    [self.languagePopup setAction:@selector(languageChanged:)];
    [view addSubview:self.languagePopup];

    // Font Size
    y -= 38;
    NSTextField *fontLabel = [NSTextField labelWithString:[self S:@"font_size"]];
    fontLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:fontLabel];

    self.fontSizePopup = [[NSPopUpButton alloc] initWithFrame:NSMakeRect(controlX, y, 200, 26) pullsDown:NO];
    [self.fontSizePopup addItemsWithTitles:@[[self S:@"font_small"], [self S:@"font_medium"], [self S:@"font_large"]]];
    NSNumber *fontSize = self.currentSettings[@"font_size"];
    if (fontSize) {
        int fs = fontSize.intValue;
        if (fs <= 10) [self.fontSizePopup selectItemAtIndex:0];
        else if (fs <= 12) [self.fontSizePopup selectItemAtIndex:1];
        else [self.fontSizePopup selectItemAtIndex:2];
    } else {
        [self.fontSizePopup selectItemAtIndex:1];
    }
    [self.fontSizePopup setTarget:self];
    [self.fontSizePopup setAction:@selector(settingChanged:)];
    [view addSubview:self.fontSizePopup];

    // Icon Mode
    y -= 38;
    NSTextField *iconLabel = [NSTextField labelWithString:[self S:@"coin_display"]];
    iconLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:iconLabel];

    self.iconModePopup = [[NSPopUpButton alloc] initWithFrame:NSMakeRect(controlX, y, 200, 26) pullsDown:NO];
    [self.iconModePopup addItemsWithTitles:@[[self S:@"icon_logo"], [self S:@"icon_text"], [self S:@"icon_both"]]];
    NSString *iconMode = self.currentSettings[@"icon_mode"];
    if ([iconMode isEqualToString:@"text"]) [self.iconModePopup selectItemAtIndex:1];
    else if ([iconMode isEqualToString:@"both"]) [self.iconModePopup selectItemAtIndex:2];
    else [self.iconModePopup selectItemAtIndex:0];
    [self.iconModePopup setTarget:self];
    [self.iconModePopup setAction:@selector(settingChanged:)];
    [view addSubview:self.iconModePopup];

    // Logo Color
    y -= 38;
    NSTextField *logoColorLabel = [NSTextField labelWithString:[self S:@"logo_color"]];
    logoColorLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:logoColorLabel];

    self.logoColorPopup = [[NSPopUpButton alloc] initWithFrame:NSMakeRect(controlX, y, 200, 26) pullsDown:NO];
    [self.logoColorPopup addItemsWithTitles:@[[self S:@"logo_color_color"], [self S:@"logo_color_gray"]]];
    NSString *logoColor = self.currentSettings[@"logo_color"];
    if ([logoColor isEqualToString:@"gray"]) [self.logoColorPopup selectItemAtIndex:1];
    else [self.logoColorPopup selectItemAtIndex:0];
    [self.logoColorPopup setTarget:self];
    [self.logoColorPopup setAction:@selector(settingChanged:)];
    [view addSubview:self.logoColorPopup];

    // Show 24h Change
    y -= 38;
    NSTextField *changeLabel = [NSTextField labelWithString:[self S:@"show_24h"]];
    changeLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:changeLabel];

    self.showChangeSwitch = [NSButton checkboxWithTitle:@"" target:self action:@selector(settingChanged:)];
    self.showChangeSwitch.frame = NSMakeRect(controlX, y, 30, 22);
    self.showChangeSwitch.state = [self.currentSettings[@"show_change"] boolValue] ? NSControlStateValueOn : NSControlStateValueOff;
    [view addSubview:self.showChangeSwitch];

    // Compact Name
    y -= 38;
    NSTextField *compactLabel = [NSTextField labelWithString:[self S:@"compact_name"]];
    compactLabel.frame = NSMakeRect(20, y, labelW, 22);
    [view addSubview:compactLabel];

    self.compactNameSwitch = [NSButton checkboxWithTitle:@"" target:self action:@selector(settingChanged:)];
    self.compactNameSwitch.frame = NSMakeRect(controlX, y, 30, 22);
    self.compactNameSwitch.state = [self.currentSettings[@"compact_name"] boolValue] ? NSControlStateValueOn : NSControlStateValueOff;
    [view addSubview:self.compactNameSwitch];

    return view;
}

#pragma mark - Tab 2: Data Sources

- (NSView *)createDataSourceTab {
    NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 480, 360)];

    NSDictionary *sourceURLs = self.currentSettings[@"source_urls"];
    if (!sourceURLs) sourceURLs = @{};
    NSDictionary *defaultURLs = self.currentSettings[@"default_urls"];
    if (!defaultURLs) defaultURLs = @{};
    NSString *currentSource = self.currentSettings[@"data_source"];

    NSArray *sources = @[@"binance", @"htx", @"gateio"];
    NSArray *labels = @[@"Binance", @"HTX (Huobi)", @"Gate.io"];

    self.sourceRadio = [[NSMatrix alloc] initWithFrame:NSMakeRect(20, 310, 440, 20)
                                                  mode:NSRadioModeMatrix
                                             prototype:[[NSButtonCell alloc] init]
                                          numberOfRows:1
                                       numberOfColumns:3];
    for (int i = 0; i < 3; i++) {
        NSButtonCell *cell = [self.sourceRadio cellAtRow:0 column:i];
        cell.title = labels[i];
        cell.buttonType = NSButtonTypeRadio;
        if ([currentSource isEqualToString:sources[i]]) {
            [self.sourceRadio selectCellAtRow:0 column:i];
        }
    }
    [self.sourceRadio setTarget:self];
    [self.sourceRadio setAction:@selector(settingChanged:)];
    [view addSubview:self.sourceRadio];

    CGFloat y = 270;

    for (int i = 0; i < 3; i++) {
        NSString *srcName = sources[i];
        NSDictionary *custom = sourceURLs[srcName];
        NSDictionary *defaults = defaultURLs[srcName];

        NSString *wsURL = custom[@"ws_url"];
        if (!wsURL || wsURL.length == 0) wsURL = defaults[@"ws_url"];
        if (!wsURL) wsURL = @"";

        NSString *apiURL = custom[@"api_url"];
        if (!apiURL || apiURL.length == 0) apiURL = defaults[@"api_url"];
        if (!apiURL) apiURL = @"";

        NSTextField *sectionLabel = [NSTextField labelWithString:[NSString stringWithFormat:@"── %@ ──", labels[i]]];
        sectionLabel.frame = NSMakeRect(20, y, 440, 18);
        sectionLabel.font = [NSFont boldSystemFontOfSize:11];
        [view addSubview:sectionLabel];
        y -= 22;

        NSTextField *wsLabel = [NSTextField labelWithString:@"WS:"];
        wsLabel.frame = NSMakeRect(20, y, 30, 18);
        wsLabel.font = [NSFont systemFontOfSize:10];
        [view addSubview:wsLabel];

        NSTextField *wsField = [[NSTextField alloc] initWithFrame:NSMakeRect(50, y, 330, 20)];
        wsField.stringValue = wsURL;
        wsField.font = [NSFont systemFontOfSize:10];
        wsField.placeholderString = @"WebSocket URL";
        [view addSubview:wsField];

        y -= 22;
        NSTextField *apiLabel = [NSTextField labelWithString:@"API:"];
        apiLabel.frame = NSMakeRect(20, y, 30, 18);
        apiLabel.font = [NSFont systemFontOfSize:10];
        [view addSubview:apiLabel];

        NSTextField *apiField = [[NSTextField alloc] initWithFrame:NSMakeRect(50, y, 330, 20)];
        apiField.stringValue = apiURL;
        apiField.font = [NSFont systemFontOfSize:10];
        apiField.placeholderString = @"REST API URL";
        [view addSubview:apiField];

        NSButton *testBtn = [NSButton buttonWithTitle:[self S:@"test"] target:self action:@selector(testConnection:)];
        testBtn.frame = NSMakeRect(390, y, 60, 22);
        testBtn.font = [NSFont systemFontOfSize:10];
        testBtn.tag = i;
        [view addSubview:testBtn];

        NSTextField *statusLabel = [NSTextField labelWithString:@""];
        statusLabel.frame = NSMakeRect(390, y + 22, 80, 18);
        statusLabel.font = [NSFont systemFontOfSize:10];
        statusLabel.editable = NO;
        statusLabel.bordered = NO;
        statusLabel.backgroundColor = NSColor.clearColor;
        [view addSubview:statusLabel];

        if (i == 0) {
            self.bnWSField = wsField; self.bnAPIField = apiField; self.bnStatusLabel = statusLabel;
        } else if (i == 1) {
            self.htxWSField = wsField; self.htxAPIField = apiField; self.htxStatusLabel = statusLabel;
        } else {
            self.gateWSField = wsField; self.gateAPIField = apiField; self.gateStatusLabel = statusLabel;
        }

        y -= 30;
    }

    NSButton *saveBtn = [NSButton buttonWithTitle:[self S:@"save_apply"] target:self action:@selector(settingChanged:)];
    saveBtn.frame = NSMakeRect(180, 10, 140, 30);
    [view addSubview:saveBtn];

    return view;
}

- (void)testConnection:(NSButton *)sender {
    NSTextField *apiField = nil;
    NSTextField *statusLabel = nil;

    switch (sender.tag) {
        case 0: apiField = self.bnAPIField; statusLabel = self.bnStatusLabel; break;
        case 1: apiField = self.htxAPIField; statusLabel = self.htxStatusLabel; break;
        case 2: apiField = self.gateAPIField; statusLabel = self.gateStatusLabel; break;
    }
    if (!apiField) return;

    statusLabel.stringValue = [self S:@"testing"];
    statusLabel.textColor = NSColor.secondaryLabelColor;

    NSString *urlStr = apiField.stringValue;
    if (urlStr.length == 0) {
        statusLabel.stringValue = [self S:@"no_url"];
        statusLabel.textColor = NSColor.systemRedColor;
        return;
    }

    NSURL *url = [NSURL URLWithString:urlStr];
    if (!url) {
        statusLabel.stringValue = [self S:@"invalid_url"];
        statusLabel.textColor = NSColor.systemRedColor;
        return;
    }

    NSURLSessionConfiguration *config = [NSURLSessionConfiguration defaultSessionConfiguration];
    config.timeoutIntervalForRequest = 5.0;
    NSURLSession *session = [NSURLSession sessionWithConfiguration:config];

    NSString *failedStr = [self S:@"failed"];
    NSString *connStr = [self S:@"connected"];

    [[session dataTaskWithURL:url completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
        dispatch_async(dispatch_get_main_queue(), ^{
            if (error) {
                statusLabel.stringValue = failedStr;
                statusLabel.textColor = NSColor.systemRedColor;
            } else {
                NSHTTPURLResponse *httpResp = (NSHTTPURLResponse *)response;
                if (httpResp.statusCode >= 200 && httpResp.statusCode < 400) {
                    statusLabel.stringValue = connStr;
                    statusLabel.textColor = NSColor.systemGreenColor;
                } else {
                    statusLabel.stringValue = [NSString stringWithFormat:@"HTTP %ld", (long)httpResp.statusCode];
                    statusLabel.textColor = NSColor.systemOrangeColor;
                }
            }
        });
    }] resume];
}

#pragma mark - Tab 3: Privacy

- (NSView *)createPrivacyTab {
    NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 480, 360)];

    CGFloat y = 280;

    NSTextField *title = [NSTextField labelWithString:[self S:@"log_mgmt"]];
    title.frame = NSMakeRect(20, y, 200, 22);
    title.font = [NSFont boldSystemFontOfSize:14];
    [view addSubview:title];

    y -= 35;
    NSTextField *pathTitle = [NSTextField labelWithString:[self S:@"log_file"]];
    pathTitle.frame = NSMakeRect(20, y, 80, 18);
    pathTitle.font = [NSFont systemFontOfSize:11];
    [view addSubview:pathTitle];

    NSString *logPath = self.currentSettings[@"log_path"];
    if (!logPath) logPath = @"~/.cryptobar/logs/cryptobar.log";
    self.logPathLabel = [NSTextField labelWithString:logPath];
    self.logPathLabel.frame = NSMakeRect(100, y, 370, 18);
    self.logPathLabel.font = [NSFont monospacedSystemFontOfSize:10 weight:NSFontWeightRegular];
    self.logPathLabel.lineBreakMode = NSLineBreakByTruncatingMiddle;
    [view addSubview:self.logPathLabel];

    y -= 40;
    NSButton *exportBtn = [NSButton buttonWithTitle:[self S:@"export_logs"] target:self action:@selector(exportLogs:)];
    exportBtn.frame = NSMakeRect(20, y, 150, 30);
    [view addSubview:exportBtn];

    NSButton *openBtn = [NSButton buttonWithTitle:[self S:@"open_log_folder"] target:self action:@selector(openLogFolder:)];
    openBtn.frame = NSMakeRect(180, y, 160, 30);
    [view addSubview:openBtn];

    y -= 50;
    NSTextField *note = [NSTextField wrappingLabelWithString:[self S:@"log_note"]];
    note.frame = NSMakeRect(20, y - 10, 440, 50);
    note.font = [NSFont systemFontOfSize:11];
    note.textColor = NSColor.secondaryLabelColor;
    [view addSubview:note];

    return view;
}

- (void)exportLogs:(id)sender {
    NSString *logPath = self.currentSettings[@"log_path"];
    if (!logPath) return;

    NSString *expandedPath = [logPath stringByExpandingTildeInPath];
    if (![[NSFileManager defaultManager] fileExistsAtPath:expandedPath]) {
        NSAlert *alert = [NSAlert new];
        alert.messageText = [self S:@"no_log_file"];
        alert.informativeText = [self S:@"no_log_desc"];
        [alert addButtonWithTitle:@"OK"];
        [alert runModal];
        return;
    }

    NSSavePanel *panel = [NSSavePanel savePanel];
    panel.nameFieldStringValue = @"cryptobar.log";
    panel.allowedContentTypes = @[[UTType typeWithFilenameExtension:@"log"]];

    if ([panel runModal] == NSModalResponseOK) {
        NSError *error = nil;
        [[NSFileManager defaultManager] copyItemAtPath:expandedPath toPath:panel.URL.path error:&error];
        if (error) {
            NSAlert *alert = [NSAlert new];
            alert.messageText = [self S:@"export_failed"];
            alert.informativeText = error.localizedDescription;
            [alert addButtonWithTitle:@"OK"];
            [alert runModal];
        }
    }
}

- (void)openLogFolder:(id)sender {
    NSString *logPath = self.currentSettings[@"log_path"];
    if (!logPath) return;
    NSString *folder = [[logPath stringByExpandingTildeInPath] stringByDeletingLastPathComponent];
    [[NSFileManager defaultManager] createDirectoryAtPath:folder withIntermediateDirectories:YES attributes:nil error:nil];
    [[NSWorkspace sharedWorkspace] openURL:[NSURL fileURLWithPath:folder]];
}

#pragma mark - Tab 4: Donate

- (NSView *)createDonateTab {
    NSView *view = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 480, 360)];

    CGFloat y = 290;

    NSTextField *title = [NSTextField labelWithString:[self S:@"support_title"]];
    title.frame = NSMakeRect(20, y, 300, 24);
    title.font = [NSFont boldSystemFontOfSize:14];
    [view addSubview:title];

    y -= 30;
    NSTextField *desc = [NSTextField wrappingLabelWithString:[self S:@"support_desc"]];
    desc.frame = NSMakeRect(20, y, 440, 30);
    desc.font = [NSFont systemFontOfSize:12];
    desc.textColor = NSColor.secondaryLabelColor;
    [view addSubview:desc];

    y -= 45;
    NSTextField *usdtIcon = [NSTextField labelWithString:@"💲"];
    usdtIcon.frame = NSMakeRect(20, y, 30, 24);
    usdtIcon.font = [NSFont systemFontOfSize:20];
    [view addSubview:usdtIcon];

    NSTextField *usdtTitle = [NSTextField labelWithString:@"USDT (TRC20)"];
    usdtTitle.frame = NSMakeRect(50, y + 2, 200, 20);
    usdtTitle.font = [NSFont boldSystemFontOfSize:13];
    [view addSubview:usdtTitle];

    y -= 8;
    NSTextField *networkLabel = [NSTextField labelWithString:[self S:@"network"]];
    networkLabel.frame = NSMakeRect(50, y, 250, 16);
    networkLabel.font = [NSFont systemFontOfSize:11];
    networkLabel.textColor = NSColor.secondaryLabelColor;
    [view addSubview:networkLabel];

    y -= 35;
    NSTextField *addrLabel = [NSTextField labelWithString:[self S:@"address"]];
    addrLabel.frame = NSMakeRect(20, y, 60, 18);
    addrLabel.font = [NSFont systemFontOfSize:11];
    [view addSubview:addrLabel];

    NSString *usdtAddr = self.currentSettings[@"usdt_address"];
    if (!usdtAddr) usdtAddr = @"";
    self.usdtField = [NSTextField labelWithString:usdtAddr];
    self.usdtField.frame = NSMakeRect(80, y, 310, 18);
    self.usdtField.font = [NSFont monospacedSystemFontOfSize:11 weight:NSFontWeightMedium];
    self.usdtField.selectable = YES;
    [view addSubview:self.usdtField];

    NSButton *copyBtn = [NSButton buttonWithTitle:[self S:@"copy"] target:self action:@selector(copyUSDT:)];
    copyBtn.frame = NSMakeRect(400, y - 2, 60, 24);
    copyBtn.font = [NSFont systemFontOfSize:11];
    [view addSubview:copyBtn];

    y -= 40;
    NSBox *separator = [[NSBox alloc] initWithFrame:NSMakeRect(20, y, 440, 1)];
    separator.boxType = NSBoxSeparator;
    [view addSubview:separator];

    y -= 25;
    NSTextField *note = [NSTextField wrappingLabelWithString:[self S:@"usdt_warning"]];
    note.frame = NSMakeRect(20, y - 20, 440, 40);
    note.font = [NSFont systemFontOfSize:11];
    note.textColor = NSColor.systemOrangeColor;
    [view addSubview:note];

    return view;
}

- (void)copyUSDT:(id)sender {
    NSString *addr = self.usdtField.stringValue;
    if (addr.length > 0) {
        [[NSPasteboard generalPasteboard] clearContents];
        [[NSPasteboard generalPasteboard] setString:addr forType:NSPasteboardTypeString];
    }
}

#pragma mark - Language Changed

- (void)languageChanged:(id)sender {
    [self settingChanged:sender];

    // Close and reopen the window so all text refreshes
    if (_settingsWindow) {
        [_settingsWindow close];
        _settingsWindow = nil;
    }
}

#pragma mark - Settings Changed

- (void)settingChanged:(id)sender {
    NSMutableDictionary *result = [NSMutableDictionary dictionary];

    double fontSizes[] = {10, 12, 14};
    result[@"font_size"] = @(fontSizes[self.fontSizePopup.indexOfSelectedItem]);

    NSArray *iconModes = @[@"logo", @"text", @"both"];
    result[@"icon_mode"] = iconModes[self.iconModePopup.indexOfSelectedItem];

    result[@"logo_color"] = (self.logoColorPopup.indexOfSelectedItem == 1) ? @"gray" : @"color";

    NSInteger langIdx = self.languagePopup.indexOfSelectedItem;
    if (langIdx >= 0 && langIdx < (NSInteger)_langCodes.count) {
        result[@"language"] = _langCodes[langIdx];
    }

    BOOL showChange = (self.showChangeSwitch.state == NSControlStateValueOn);
    BOOL compactName = (self.compactNameSwitch.state == NSControlStateValueOn);
    result[@"show_change"] = [NSNumber numberWithBool:showChange];
    result[@"compact_name"] = [NSNumber numberWithBool:compactName];

    NSInteger srcIdx = [self.sourceRadio selectedColumn];
    NSArray *sources = @[@"binance", @"htx", @"gateio"];
    if (srcIdx >= 0 && srcIdx < 3) {
        result[@"data_source"] = sources[srcIdx];
    }

    NSMutableDictionary *urls = [NSMutableDictionary dictionary];
    urls[@"binance"] = @{@"ws_url": self.bnWSField.stringValue ?: @"",
                         @"api_url": self.bnAPIField.stringValue ?: @""};
    urls[@"htx"] = @{@"ws_url": self.htxWSField.stringValue ?: @"",
                     @"api_url": self.htxAPIField.stringValue ?: @""};
    urls[@"gateio"] = @{@"ws_url": self.gateWSField.stringValue ?: @"",
                        @"api_url": self.gateAPIField.stringValue ?: @""};
    result[@"source_urls"] = urls;

    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:result options:0 error:nil];
    if (jsonData) {
        NSString *jsonStr = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
        settingsChanged(jsonStr.UTF8String);
    }
}

#pragma mark - Window delegate

- (void)windowWillClose:(NSNotification *)notification {
    _settingsWindow = nil;
}

@end

#pragma mark - C entry point

void showSettingsWindow(const char *jsonString) {
    NSString *jsonNSString = nil;
    if (jsonString) {
        jsonNSString = [NSString stringWithUTF8String:jsonString];
    }

    dispatch_async(dispatch_get_main_queue(), ^{
        if (_settingsWindow) {
            [_settingsWindow makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:YES];
            return;
        }

        NSDictionary *settings = nil;
        if (jsonNSString) {
            NSData *data = [jsonNSString dataUsingEncoding:NSUTF8StringEncoding];
            settings = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        }
        if (!settings) settings = @{};

        NSDictionary *strings = settings[@"strings"];
        NSString *windowTitle = strings[@"window_title"];
        if (!windowTitle) windowTitle = @"CryptoBar Settings";

        NSRect frame = NSMakeRect(0, 0, 540, 460);
        NSUInteger style = NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable;
        _settingsWindow = [[NSWindow alloc] initWithContentRect:frame
                                                     styleMask:style
                                                       backing:NSBackingStoreBuffered
                                                         defer:NO];
        _settingsWindow.title = windowTitle;
        [_settingsWindow center];

        SettingsWindowDelegate *delegate = [[SettingsWindowDelegate alloc] init];
        [delegate setupWithJSON:settings];
        _settingsWindow.delegate = delegate;
        objc_setAssociatedObject(_settingsWindow, "delegate", delegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

        NSTabView *tabView = [[NSTabView alloc] initWithFrame:NSMakeRect(0, 0, 540, 440)];
        tabView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;

        NSString *tab1Label = strings[@"tab_personalization"] ?: @"Personalization";
        NSString *tab2Label = strings[@"tab_datasource"] ?: @"Data Sources";
        NSString *tab3Label = strings[@"tab_privacy"] ?: @"Privacy";
        NSString *tab4Label = strings[@"tab_donate"] ?: @"Donate";

        NSTabViewItem *tab1 = [[NSTabViewItem alloc] initWithIdentifier:@"personalization"];
        tab1.label = tab1Label;
        tab1.view = [delegate createPersonalizationTab];
        [tabView addTabViewItem:tab1];

        NSTabViewItem *tab2 = [[NSTabViewItem alloc] initWithIdentifier:@"datasource"];
        tab2.label = tab2Label;
        tab2.view = [delegate createDataSourceTab];
        [tabView addTabViewItem:tab2];

        NSTabViewItem *tab3 = [[NSTabViewItem alloc] initWithIdentifier:@"privacy"];
        tab3.label = tab3Label;
        tab3.view = [delegate createPrivacyTab];
        [tabView addTabViewItem:tab3];

        NSTabViewItem *tab4 = [[NSTabViewItem alloc] initWithIdentifier:@"donate"];
        tab4.label = tab4Label;
        tab4.view = [delegate createDonateTab];
        [tabView addTabViewItem:tab4];

        delegate.tabView = tabView;
        _settingsWindow.contentView = tabView;

        [_settingsWindow makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
    });
}
