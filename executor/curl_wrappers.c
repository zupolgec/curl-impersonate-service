// This file is compiled separately and needs the REAL curl functions,
// not the shielded versions. Include curl.h directly WITHOUT the shielding.
#include <curl/curl.h>

CURLcode _curl_easy_setopt_ptr(CURL *curl, CURLoption option, void *param) {
    return curl_easy_setopt(curl, option, param);
}

CURLcode _curl_easy_setopt_long(CURL *curl, CURLoption option, long param) {
    return curl_easy_setopt(curl, option, param);
}

CURLcode _curl_easy_getinfo_ptr(CURL *curl, CURLINFO info, void *param) {
    return curl_easy_getinfo(curl, info, param);
}
