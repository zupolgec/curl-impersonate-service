#include <curl/curl.h>
#include "curl_wrappers.h"

CURLcode _curl_easy_setopt_ptr(CURL *curl, CURLoption option, void *param) {
    return curl_easy_setopt(curl, option, param);
}

CURLcode _curl_easy_setopt_long(CURL *curl, CURLoption option, long param) {
    return curl_easy_setopt(curl, option, param);
}

CURLcode _curl_easy_getinfo_ptr(CURL *curl, CURLINFO info, void *param) {
    return curl_easy_getinfo(curl, info, param);
}
