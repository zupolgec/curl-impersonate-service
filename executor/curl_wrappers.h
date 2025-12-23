#ifndef CURL_WRAPPERS_H
#define CURL_WRAPPERS_H

// Shield variadic functions from CGO
#define curl_easy_setopt(...) dummy_setopt
#define curl_easy_getinfo(...) dummy_getinfo
#include <curl/curl.h>
#undef curl_easy_setopt
#undef curl_easy_getinfo

// Declaration for curl-impersonate function
int curl_easy_impersonate(CURL *curl, const char *target, int default_headers);

// Wrappers for variadic functions
CURLcode _curl_easy_setopt_ptr(CURL *curl, CURLoption option, void *param);
CURLcode _curl_easy_setopt_long(CURL *curl, CURLoption option, long param);
CURLcode _curl_easy_getinfo_ptr(CURL *curl, CURLINFO info, void *param);

#endif
