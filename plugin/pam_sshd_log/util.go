package main

/*
#include <string.h>

#define PAM_SM_SESSION
#include <security/pam_modules.h>

char *string_from_argv(int i, char **argv) {
  return strdup(argv[i]);
}

char *get_username(pam_handle_t *pamh) {
  const char *username;
  if(PAM_SUCCESS != pam_get_user(pamh, &username, NULL)) return NULL;
  return strdup(username);
}

char *get_remote_host(pam_handle_t *pamh) {
  const char *remote_host;
  if(PAM_SUCCESS != pam_get_item(pamh, PAM_RHOST, (const void**)&remote_host)) return NULL;
  return strdup(remote_host);
}

char *get_auth_info(pam_handle_t *pamh) {
  const char *auth_info;
  auth_info = pam_getenv(pamh, "SSH_AUTH_INFO_0");
  return strdup(auth_info);
}

char *get_service_name(pam_handle_t *pamh) {
  const char *service_name;
  if(PAM_SUCCESS != pam_get_item(pamh, PAM_SERVICE, (const void**)&service_name)) return NULL;
  return strdup(service_name);
}
*/
import "C"
