#ifndef CURL_WRAPPERS_H
#define CURL_WRAPPERS_H

#include <curl/curl.h>

CURLcode _curl_easy_setopt_ptr(CURL *curl, CURLoption option, void *param);
CURLcode _curl_easy_setopt_long(CURL *curl, CURLoption option, long param);
CURLcode _curl_easy_getinfo_ptr(CURL *curl, CURLINFO info, void *param);

#endif
