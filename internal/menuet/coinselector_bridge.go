package menuet

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

#ifndef __COINSELECTOR_H__
#import "coinselector.h"
#endif

*/
import "C"

import (
	"unsafe"
)

var CoinSelectionCallback func(jsonStr string)

func ShowCoinSelector(jsonStr string) {
	cstr := C.CString(jsonStr)
	defer C.free(unsafe.Pointer(cstr))
	C.showCoinSelectorWindow(cstr)
}

//export coinSelectionChanged
func coinSelectionChanged(jsonCStr *C.char) {
	jsonStr := C.GoString(jsonCStr)
	if CoinSelectionCallback != nil {
		CoinSelectionCallback(jsonStr)
	}
}
