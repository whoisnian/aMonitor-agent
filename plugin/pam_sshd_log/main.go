package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"unsafe"
)

/*
#cgo LDFLAGS: -lpam

#include <stdlib.h>

#define PAM_SM_SESSION
#include <security/pam_appl.h>

char *get_user(pam_handle_t *pamh);
char *get_rhost(pam_handle_t *pamh);
char *get_auth_info(pam_handle_t *pamh);
*/
import "C"

type sshdInfo struct {
	User     string
	RHost    string
	AuthInfo string
}

//export pam_sm_open_session
func pam_sm_open_session(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	var sshd sshdInfo
	cUser := C.get_user(pamh)
	if cUser != nil {
		defer C.free(unsafe.Pointer(cUser))
	}
	sshd.User = C.GoString(cUser)

	cRHost := C.get_rhost(pamh)
	if cRHost != nil {
		defer C.free(unsafe.Pointer(cRHost))
	}
	sshd.RHost = C.GoString(cRHost)

	cAuthInfo := C.get_auth_info(pamh)
	if cAuthInfo != nil {
		defer C.free(unsafe.Pointer(cAuthInfo))
	}
	sshd.AuthInfo = C.GoString(cAuthInfo)

	jsonValue, _ := json.Marshal(sshd)
	http.Post("http://127.0.0.1:8001", "application/json", bytes.NewBuffer(jsonValue))

	return C.PAM_IGNORE
}

//export pam_sm_close_session
func pam_sm_close_session(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

func main() {}
