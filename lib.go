package main

// static char* get_p(char **pp, int i)
// {
//	    return pp[i];
// }
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
)

func renderData(data interface{}) *C.char {
	ret := map[string]interface{}{"data": data}
	result, _ := json.Marshal(ret)
	return C.CString(string(result))
}

func renderError(err error) *C.char {
	pc, fn, line, _ := runtime.Caller(1)
	error := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
	ret := map[string]interface{}{"error": error}
	result, _ := json.Marshal(ret)
	return C.CString(string(result))
}

//export init_
func init_() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//export say_hello_
func say_hello_(name *C.char) {
	_name := C.GoString(name)
	log.Println("hello,world", _name)
}
