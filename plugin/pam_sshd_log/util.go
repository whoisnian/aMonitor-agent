package main

/*
#cgo LDFLAGS: -lpam

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define PAM_SM_SESSION
#include <security/pam_modules.h>

char *get_user(pam_handle_t *pamh) {
	const char *user;
  if(PAM_SUCCESS != pam_get_user(pamh, &user, NULL)) return NULL;
  return strdup(user);
}

char *get_rhost(pam_handle_t *pamh) {
	const char *rhost;
  if(PAM_SUCCESS != pam_get_item(pamh, PAM_RHOST, (const void**)&rhost)) return NULL;
  return strdup(rhost);
}

char *get_auth_info(pam_handle_t *pamh) {
	const char *auth_info;
	auth_info = pam_getenv(pamh, "SSH_AUTH_INFO_0");
  return strdup(auth_info);
}
*/
import "C"
