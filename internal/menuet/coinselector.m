#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#import "coinselector.h"

void coinSelectionChanged(const char *jsonString);

static NSWindow *_coinWindow = nil;

@interface CoinSelectorDelegate : NSObject <NSWindowDelegate, NSTableViewDataSource, NSTableViewDelegate, NSSearchFieldDelegate>
@property(nonatomic, strong) NSArray *allCoins;
@property(nonatomic, strong) NSArray *filteredCoins;
@property(nonatomic, strong) NSMutableSet *selectedPairs;
@property(nonatomic, strong) NSTableView *tableView;
@property(nonatomic, strong) NSSearchField *searchField;
@property(nonatomic, strong) NSTextField *countLabel;
@property(nonatomic, strong) NSDictionary *strings;
@end

@implementation CoinSelectorDelegate

- (void)setupWithJSON:(NSDictionary *)data {
    self.strings = data[@"strings"] ?: @{};

    NSArray *coins = data[@"coins"] ?: @[];
    self.allCoins = coins;
    self.filteredCoins = coins;

    NSArray *selected = data[@"selected_pairs"] ?: @[];
    self.selectedPairs = [NSMutableSet setWithArray:selected];
}

- (NSString *)S:(NSString *)key {
    NSString *val = self.strings[key];
    return val ? val : key;
}

- (void)updateFilter:(NSString *)query {
    if (!query || query.length == 0) {
        self.filteredCoins = self.allCoins;
    } else {
        NSString *q = [query uppercaseString];
        NSMutableArray *result = [NSMutableArray array];
        for (NSDictionary *coin in self.allCoins) {
            NSString *symbol = [coin[@"symbol"] uppercaseString] ?: @"";
            NSString *name = [coin[@"name"] uppercaseString] ?: @"";
            if ([symbol containsString:q] || [name containsString:q]) {
                [result addObject:coin];
            }
        }
        self.filteredCoins = result;
    }
    [self.tableView reloadData];
    [self updateCountLabel];
}

- (void)updateCountLabel {
    NSInteger total = self.allCoins.count;
    NSInteger showing = self.filteredCoins.count;
    NSInteger selected = self.selectedPairs.count;
    self.countLabel.stringValue = [NSString stringWithFormat:@"%@ %ld / %ld    %@ %ld",
        [self S:@"cs_showing"], (long)showing, (long)total,
        [self S:@"cs_selected"], (long)selected];
}

#pragma mark - NSTableViewDataSource

- (NSInteger)numberOfRowsInTableView:(NSTableView *)tableView {
    return (NSInteger)self.filteredCoins.count;
}

#pragma mark - NSTableViewDelegate

- (NSView *)tableView:(NSTableView *)tableView viewForTableColumn:(NSTableColumn *)tableColumn row:(NSInteger)row {
    if (row < 0 || row >= (NSInteger)self.filteredCoins.count) return nil;

    NSDictionary *coin = self.filteredCoins[row];
    NSString *colId = tableColumn.identifier;

    if ([colId isEqualToString:@"check"]) {
        NSButton *check = [NSButton checkboxWithTitle:@"" target:self action:@selector(checkClicked:)];
        check.tag = row;
        NSString *pair = coin[@"pair"] ?: @"";
        check.state = [self.selectedPairs containsObject:pair] ? NSControlStateValueOn : NSControlStateValueOff;
        return check;
    }

    if ([colId isEqualToString:@"symbol"]) {
        NSTextField *tf = [NSTextField labelWithString:coin[@"symbol"] ?: @""];
        tf.font = [NSFont boldSystemFontOfSize:12];
        return tf;
    }

    if ([colId isEqualToString:@"name"]) {
        NSTextField *tf = [NSTextField labelWithString:coin[@"name"] ?: @""];
        tf.font = [NSFont systemFontOfSize:12];
        tf.textColor = NSColor.secondaryLabelColor;
        return tf;
    }

    if ([colId isEqualToString:@"pair"]) {
        NSTextField *tf = [NSTextField labelWithString:coin[@"pair"] ?: @""];
        tf.font = [NSFont monospacedSystemFontOfSize:10 weight:NSFontWeightRegular];
        tf.textColor = NSColor.tertiaryLabelColor;
        return tf;
    }

    return nil;
}

- (CGFloat)tableView:(NSTableView *)tableView heightOfRow:(NSInteger)row {
    return 28;
}

- (void)checkClicked:(NSButton *)sender {
    NSInteger row = sender.tag;
    if (row < 0 || row >= (NSInteger)self.filteredCoins.count) return;

    NSDictionary *coin = self.filteredCoins[row];
    NSString *pair = coin[@"pair"] ?: @"";

    if (sender.state == NSControlStateValueOn) {
        [self.selectedPairs addObject:pair];
    } else {
        [self.selectedPairs removeObject:pair];
    }

    [self updateCountLabel];
    [self notifyGo];
}

- (void)notifyGo {
    NSMutableArray *selected = [NSMutableArray array];
    for (NSDictionary *coin in self.allCoins) {
        NSString *pair = coin[@"pair"] ?: @"";
        if ([self.selectedPairs containsObject:pair]) {
            [selected addObject:coin];
        }
    }

    NSDictionary *result = @{@"selected": selected};
    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:result options:0 error:nil];
    if (jsonData) {
        NSString *jsonStr = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
        coinSelectionChanged(jsonStr.UTF8String);
    }
}

#pragma mark - NSSearchFieldDelegate

- (void)controlTextDidChange:(NSNotification *)notification {
    NSSearchField *field = notification.object;
    [self updateFilter:field.stringValue];
}

#pragma mark - Window delegate

- (void)windowWillClose:(NSNotification *)notification {
    _coinWindow = nil;
}

@end

#pragma mark - C entry point

void showCoinSelectorWindow(const char *jsonString) {
    NSString *jsonNSString = nil;
    if (jsonString) {
        jsonNSString = [NSString stringWithUTF8String:jsonString];
    }

    dispatch_async(dispatch_get_main_queue(), ^{
        if (_coinWindow) {
            [_coinWindow makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:YES];
            return;
        }

        NSDictionary *data = nil;
        if (jsonNSString) {
            NSData *d = [jsonNSString dataUsingEncoding:NSUTF8StringEncoding];
            data = [NSJSONSerialization JSONObjectWithData:d options:0 error:nil];
        }
        if (!data) data = @{};

        CoinSelectorDelegate *delegate = [[CoinSelectorDelegate alloc] init];
        [delegate setupWithJSON:data];

        NSDictionary *strings = data[@"strings"] ?: @{};
        NSString *windowTitle = strings[@"cs_title"] ?: @"Select Coins";

        NSRect frame = NSMakeRect(0, 0, 520, 500);
        NSUInteger style = NSWindowStyleMaskTitled | NSWindowStyleMaskClosable |
                           NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable;
        _coinWindow = [[NSWindow alloc] initWithContentRect:frame
                                                  styleMask:style
                                                    backing:NSBackingStoreBuffered
                                                      defer:NO];
        _coinWindow.title = windowTitle;
        _coinWindow.minSize = NSMakeSize(400, 350);
        [_coinWindow center];
        _coinWindow.delegate = delegate;
        objc_setAssociatedObject(_coinWindow, "delegate", delegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

        NSView *content = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 520, 500)];

        // Search field
        NSSearchField *search = [[NSSearchField alloc] initWithFrame:NSMakeRect(15, 460, 490, 28)];
        search.placeholderString = strings[@"cs_search_placeholder"] ?: @"Search coins...";
        search.delegate = delegate;
        search.autoresizingMask = NSViewWidthSizable | NSViewMinYMargin;
        delegate.searchField = search;
        [content addSubview:search];

        // Count label
        NSTextField *countLabel = [NSTextField labelWithString:@""];
        countLabel.frame = NSMakeRect(15, 435, 490, 18);
        countLabel.font = [NSFont systemFontOfSize:11];
        countLabel.textColor = NSColor.secondaryLabelColor;
        countLabel.autoresizingMask = NSViewWidthSizable | NSViewMinYMargin;
        delegate.countLabel = countLabel;
        [content addSubview:countLabel];

        // Scroll view with table
        NSScrollView *scrollView = [[NSScrollView alloc] initWithFrame:NSMakeRect(15, 10, 490, 420)];
        scrollView.hasVerticalScroller = YES;
        scrollView.autohidesScrollers = YES;
        scrollView.borderType = NSBezelBorder;
        scrollView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;

        NSTableView *table = [[NSTableView alloc] initWithFrame:scrollView.bounds];
        table.dataSource = delegate;
        table.delegate = delegate;
        table.rowHeight = 28;
        table.usesAlternatingRowBackgroundColors = YES;
        table.columnAutoresizingStyle = NSTableViewLastColumnOnlyAutoresizingStyle;

        NSTableColumn *checkCol = [[NSTableColumn alloc] initWithIdentifier:@"check"];
        checkCol.title = @"";
        checkCol.width = 30;
        checkCol.minWidth = 30;
        checkCol.maxWidth = 30;
        [table addTableColumn:checkCol];

        NSString *symbolHeader = strings[@"cs_col_symbol"] ?: @"Symbol";
        NSTableColumn *symbolCol = [[NSTableColumn alloc] initWithIdentifier:@"symbol"];
        symbolCol.title = symbolHeader;
        symbolCol.width = 80;
        symbolCol.minWidth = 60;
        [table addTableColumn:symbolCol];

        NSString *nameHeader = strings[@"cs_col_name"] ?: @"Name";
        NSTableColumn *nameCol = [[NSTableColumn alloc] initWithIdentifier:@"name"];
        nameCol.title = nameHeader;
        nameCol.width = 180;
        nameCol.minWidth = 100;
        [table addTableColumn:nameCol];

        NSString *pairHeader = strings[@"cs_col_pair"] ?: @"Pair";
        NSTableColumn *pairCol = [[NSTableColumn alloc] initWithIdentifier:@"pair"];
        pairCol.title = pairHeader;
        pairCol.width = 120;
        pairCol.minWidth = 80;
        [table addTableColumn:pairCol];

        delegate.tableView = table;
        scrollView.documentView = table;
        [content addSubview:scrollView];

        _coinWindow.contentView = content;
        [delegate updateCountLabel];

        [_coinWindow makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:YES];
    });
}
