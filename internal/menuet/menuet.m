#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

#import "NSImage+Resize.h"
#import "menuet.h"

void itemClicked(const char *);
void notificationRespond(const char *, const char *);
const char *children(const char *);
void menuClosed(const char *);
bool hideStartup();
bool runningAtStartup();
void toggleStartup();
void shutdownWait();

NSStatusItem *_statusItem;

@interface MenuetMenu : NSMenu <NSMenuDelegate>

@property(nonatomic, copy) NSString *unique;
@property(nonatomic, assign) BOOL root;
@property(nonatomic, assign) BOOL open;

@end

@implementation MenuetMenu
- (id)init {
	self = [super init];
	if (self) {
		self.delegate = self;
		self.autoenablesItems = false;
	}
	return self;
}

- (void)refreshVisibleMenus {
	if (!self.open) {
		return;
	}
	[self menuWillOpen:self];
	for (NSMenuItem *item in self.itemArray) {
		MenuetMenu *menu = (MenuetMenu *)item.submenu;
		if (menu != NULL) {
			[menu refreshVisibleMenus];
		}
	}
}

- (void)populate:(NSArray *)items {
	for (int i = 0; i < items.count; i++) {
		NSMenuItem *item = nil;
		if (i < self.numberOfItems) {
			item = [self itemAtIndex:i];
		}
		NSDictionary *dict = [items objectAtIndex:i];
		NSString *type = dict[@"Type"];
		if ([type isEqualTo:@"separator"]) {
			if (!item || !item.isSeparatorItem) {
				[self insertItem:[NSMenuItem separatorItem] atIndex:i];
			}
			continue;
		}
		NSString *unique = dict[@"Unique"];
		NSString *text = dict[@"Text"];
		NSString *imageName = dict[@"Image"];
		NSNumber *fontSize = dict[@"FontSize"];
		NSNumber *fontWeight = dict[@"FontWeight"];
		BOOL state = [dict[@"State"] boolValue];
		BOOL hasChildren = [dict[@"HasChildren"] boolValue];
		BOOL clickable = [dict[@"Clickable"] boolValue];
		if (!item || item.isSeparatorItem) {
			item =
				[self insertItemWithTitle:@"" action:nil keyEquivalent:@"" atIndex:i];
		}
		NSMutableDictionary *attributes = [NSMutableDictionary new];
		float size = fontSize.floatValue;
		if (fontSize == 0) {
			size = 14;
		}
		attributes[NSFontAttributeName] =
			[NSFont monospacedDigitSystemFontOfSize:size
			 weight:fontWeight.floatValue];
		item.attributedTitle =
			[[NSMutableAttributedString alloc] initWithString:text
			 attributes:attributes];
		item.target = self;
		if (clickable) {
			item.action = @selector(press:);
			item.representedObject = unique;
		} else {
			item.action = nil;
			item.representedObject = nil;
		}
		if (state) {
			item.state = NSControlStateValueOn;
		} else {
			item.state = NSControlStateValueOff;
		}
		if (hasChildren) {
			if (!item.submenu) {
				item.submenu = [MenuetMenu new];
			}
			MenuetMenu *menu = (MenuetMenu *)item.submenu;
			menu.unique = unique;
		} else if (item.submenu) {
			item.submenu = nil;
		}
		item.enabled = clickable || hasChildren;
		item.image = [NSImage imageFromName:imageName withHeight:16];
	}
	while (self.numberOfItems > items.count) {
		[self removeItemAtIndex:self.numberOfItems - 1];
	}
}

// The documentation says not to make changes here, but it seems to work.
// submenuAction does not appear to be called, and menuNeedsUpdate is only
// called once per tracking session.
- (void)menuWillOpen:(MenuetMenu *)menu {
	if (self.root) {
		// For the root menu, we generate a new unique every time it's opened. Go
		// handles all other unique generation.
		self.unique = [[[[NSProcessInfo processInfo] globallyUniqueString]
		                substringFromIndex:51] stringByAppendingString:@":root"];
	}
	const char *str = children(self.unique.UTF8String);
	NSArray *items = @[];
	if (str != NULL) {
		items = [NSJSONSerialization
		         JSONObjectWithData:[[NSString stringWithUTF8String:str]
		                             dataUsingEncoding:NSUTF8StringEncoding]
		         options:0
		         error:nil];
		free((char *)str);
	}
	if (self.root) {
		items = [items arrayByAddingObjectsFromArray:@[
				 @{@"Type" : @"separator",
				   @"Clickable" : @YES},
		]];
		if (!hideStartup()) {
			items = [items arrayByAddingObjectsFromArray:@[
					@{@"Text" : @"Start at Login",
					@"Clickable" : @YES},
			]];
		}
		items = [items arrayByAddingObjectsFromArray:@[
				 @{@"Text" : @"Quit",
				   @"Clickable" : @YES},
		]];
	}
	[self populate:items];
	if (self.root) {
		NSMenuItem *item = nil;
		if (!hideStartup()) {
			item = [self itemAtIndex:items.count - 2];
			item.action = @selector(toggleStartup:);
			if (runningAtStartup()) {
				item.state = NSControlStateValueOn;
			} else {
				item.state = NSControlStateValueOff;
			}
		}
		item = [self itemAtIndex:items.count - 1];
		item.action = @selector(prepareShutdown:);
	}
	self.open = YES;
}

- (void)menuDidClose:(MenuetMenu *)menu {
	self.open = NO;
	menuClosed(self.unique.UTF8String);
}

- (void)press:(id)sender {
	NSString *callback = [sender representedObject];
	itemClicked(callback.UTF8String);
}

- (void)toggleStartup:(id)sender {
	toggleStartup();
}

- (void)prepareShutdown:(id)sender {
	shutdownWait();
	[NSApp terminate: nil];
}

@end

@interface MenuetAppDelegate : NSObject <NSApplicationDelegate, NSMenuDelegate, NSUserNotificationCenterDelegate>

@end

static NSImage *makeGrayscale(NSImage *src) {
	NSSize size = src.size;
	NSImage *gray = [[NSImage alloc] initWithSize:size];
	[gray lockFocus];
	CIImage *ci = [[CIImage alloc] initWithData:[src TIFFRepresentation]];
	CIFilter *filter = [CIFilter filterWithName:@"CIColorMonochrome"];
	[filter setValue:ci forKey:kCIInputImageKey];
	[filter setValue:[CIColor colorWithRed:0.7 green:0.7 blue:0.7] forKey:@"inputColor"];
	[filter setValue:@1.0 forKey:@"inputIntensity"];
	CIImage *output = filter.outputImage;
	if (output) {
		NSCIImageRep *rep = [NSCIImageRep imageRepWithCIImage:output];
		[rep drawInRect:NSMakeRect(0, 0, size.width, size.height)];
	}
	[gray unlockFocus];
	return gray;
}

void setState(const char *jsonString) {
	NSDictionary *state = [NSJSONSerialization
	                       JSONObjectWithData:[[NSString stringWithUTF8String:jsonString]
	                                           dataUsingEncoding:NSUTF8StringEncoding]
	                       options:0
	                       error:nil];
	dispatch_async(dispatch_get_main_queue(), ^{
		CGFloat fontSize = 14;
		NSNumber *fontSizeNum = state[@"FontSize"];
		if (fontSizeNum != nil && [fontSizeNum floatValue] > 0) {
			fontSize = [fontSizeNum floatValue];
		}
		NSDictionary *textAttrs = @{
			NSFontAttributeName :
				[NSFont monospacedDigitSystemFontOfSize:fontSize
				 weight:NSFontWeightRegular],
			NSBaselineOffsetAttributeName : @0.0
		};

		NSArray *segments = state[@"Segments"];
		if (segments != nil && [segments count] > 0) {
			NSMutableAttributedString *rich = [[NSMutableAttributedString alloc] init];
			for (NSDictionary *seg in segments) {
				NSString *imgURL = seg[@"image"];
				NSString *text = seg[@"text"];
				BOOL grayscale = [seg[@"grayscale"] boolValue];
				if (imgURL != nil && imgURL.length > 0) {
					NSImage *img = [NSImage imageFromName:imgURL withHeight:fontSize];
					if (img != nil) {
						if (grayscale) {
							img = makeGrayscale(img);
						}
						NSTextAttachment *attachment = [[NSTextAttachment alloc] init];
						attachment.image = img;
						CGFloat yOffset = (fontSize - img.size.height) / 2.0 - 1.0;
						attachment.bounds = CGRectMake(0, yOffset, img.size.width, img.size.height);
						NSAttributedString *imgStr = [NSAttributedString attributedStringWithAttachment:attachment];
						[rich appendAttributedString:imgStr];
					}
				}
				if (text != nil && text.length > 0) {
					NSAttributedString *textStr = [[NSAttributedString alloc]
					                               initWithString:text
					                               attributes:textAttrs];
					[rich appendAttributedString:textStr];
				}
			}
			_statusItem.button.attributedTitle = rich;
			_statusItem.button.image = nil;
		} else {
			NSString *title = state[@"Title"];
			if (title == nil) title = @"";
			_statusItem.button.attributedTitle = [[NSAttributedString alloc]
			                                      initWithString:title
			                                      attributes:textAttrs];
			NSString *imageName = state[@"Image"];
			NSImage *image = [NSImage imageFromName:imageName withHeight:22];
			_statusItem.button.image = image;
			if (image) {
				_statusItem.button.image.template = true;
				_statusItem.button.imagePosition = NSImageLeft;
			}
		}
	});
}

void menuChanged() {
        dispatch_async(dispatch_get_main_queue(), ^{
		MenuetMenu *menu = (MenuetMenu *)_statusItem.menu;
		[menu refreshVisibleMenus];
	});
}

void createAndRunApplication() {
        [NSAutoreleasePool new];
        NSApplication *a = NSApplication.sharedApplication;
        MenuetAppDelegate *d = [MenuetAppDelegate new];
        [a setDelegate:d];
        [[NSUserNotificationCenter defaultUserNotificationCenter] setDelegate:d];
        [a setActivationPolicy:NSApplicationActivationPolicyAccessory];
        _statusItem = [[NSStatusBar systemStatusBar]
                       statusItemWithLength:NSVariableStatusItemLength];
        MenuetMenu *menu = [MenuetMenu new];
        menu.root = true;
        _statusItem.menu = menu;
        [a run];
}

@implementation MenuetAppDelegate

- (NSApplicationTerminateReply)applicationShouldTerminate:
        (NSApplication *)sender {
        return NSTerminateNow;
}

- (void)userNotificationCenter:(NSUserNotificationCenter *)center didActivateNotification:(NSUserNotification *)notification {
        if (notification.activationType == NSUserNotificationActivationTypeReplied) {
                NSString* userResponse = notification.response.string;
                notificationRespond(notification.identifier.UTF8String, userResponse.UTF8String);
	} else {
                notificationRespond(notification.identifier.UTF8String, @"".UTF8String);
	}
}

@end
