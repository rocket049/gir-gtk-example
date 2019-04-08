//+build cwrap

package main

import "C"

import "unsafe"

//export TreeRowCall
func TreeRowCall(v, p, c unsafe.Pointer) {
	win1.RowActivated(v, p, c)
}
