package menuet

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers

#import <Cocoa/Cocoa.h>

#ifndef __SETTINGS_H__
#import "settings.h"
#endif

*/
import "C"

import (
	"unsafe"
)

// SettingsCallback is called when settings change from the ObjC window
var SettingsCallback func(jsonStr string)

// ShowSettings opens the native settings window with given JSON config
func ShowSettings(jsonStr string) {
	cstr := C.CString(jsonStr)
	defer C.free(unsafe.Pointer(cstr))
	C.showSettingsWindow(cstr)
}

//export settingsChanged
func settingsChanged(jsonCStr *C.char) {
	jsonStr := C.GoString(jsonCStr)
	if SettingsCallback != nil {
		SettingsCallback(jsonStr)
	}
}
