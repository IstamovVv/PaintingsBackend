package cast

import (
	"reflect"
	"unsafe"
)

func StringToByteArray(str string) []byte {
	return (*[0x7fff0000]byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&str)).Data))[:len(str):len(str)]
}

func ByteArrayToString(arr []byte) string {
	return *(*string)(unsafe.Pointer(&arr))
}
