package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"unsafe"
)

/*
#cgo LDFLAGS: -lpam

#include <stdlib.h>

#define PAM_SM_SESSION
#include <security/pam_appl.h>

char *string_from_argv(int i, char **argv);
char *get_username(pam_handle_t *pamh);
char *get_remote_host(pam_handle_t *pamh);
char *get_auth_info(pam_handle_t *pamh);
char *get_service_name(pam_handle_t *pamh);
*/
import "C"

type sshdInfo struct {
	Username   string
	RemoteHost string
	AuthInfo   string
}

func sliceFromArgv(argc C.int, argv **C.char) []string {
	res := make([]string, 0, argc)
	for i := 0; i < int(argc); i++ {
		cStr := C.string_from_argv(C.int(i), argv)
		defer C.free(unsafe.Pointer(cStr))
		res = append(res, C.GoString(cStr))
	}
	return res
}

//export pam_sm_open_session
func pam_sm_open_session(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	cServiceName := C.get_service_name(pamh)
	if cServiceName != nil {
		defer C.free(unsafe.Pointer(cServiceName))
	}
	if "sshd" != strings.TrimSpace(C.GoString(cServiceName)) {
		return C.PAM_IGNORE
	}

	var sshd sshdInfo
	cUsername := C.get_username(pamh)
	if cUsername != nil {
		defer C.free(unsafe.Pointer(cUsername))
	}
	sshd.Username = strings.TrimSpace(C.GoString(cUsername))

	cRemoteHost := C.get_remote_host(pamh)
	if cRemoteHost != nil {
		defer C.free(unsafe.Pointer(cRemoteHost))
	}
	sshd.RemoteHost = strings.TrimSpace(C.GoString(cRemoteHost))

	cAuthInfo := C.get_auth_info(pamh)
	if cAuthInfo != nil {
		defer C.free(unsafe.Pointer(cAuthInfo))
	}
	sshd.AuthInfo = strings.TrimSpace(C.GoString(cAuthInfo))

	addr := "127.0.0.1:8001"
	for _, arg := range sliceFromArgv(argc, argv) {
		opt := strings.Split(arg, "=")
		switch opt[0] {
		case "addr":
			addr = opt[1]
		}
	}

	content, _ := json.Marshal(sshd)
	resp, err := http.Post("http://"+addr+"?category=sshdInfo", "application/json", bytes.NewBuffer(content))
	if err == nil {
		defer resp.Body.Close()
	}

	return C.PAM_IGNORE
}

//export pam_sm_close_session
func pam_sm_close_session(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	return C.PAM_IGNORE
}

func main() {}
