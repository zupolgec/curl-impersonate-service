#ifndef CURL_WRAPPERS_H
#define CURL_WRAPPERS_H

// Forward declarations of curl types we need
// This avoids including curl.h directly in CGO context
typedef void CURL;
typedef int CURLcode;
typedef int CURLoption;
typedef int CURLINFO;

// Common CURLcode values
#define CURLE_OK 0
#define CURLE_OPERATION_TIMEDOUT 28
#define CURLE_COULDNT_RESOLVE_HOST 6
#define CURLE_SSL_CONNECT_ERROR 35
#define CURLE_PEER_FAILED_VERIFICATION 60

// Common CURLoption values
#define CURLOPT_URL 10002
#define CURLOPT_CUSTOMREQUEST 10036
#define CURLOPT_HTTPHEADER 10023
#define CURLOPT_POSTFIELDS 10015
#define CURLOPT_POSTFIELDSIZE 60
#define CURLOPT_FOLLOWLOCATION 52
#define CURLOPT_TIMEOUT 13
#define CURLOPT_WRITEFUNCTION 20011
#define CURLOPT_WRITEDATA 10001
#define CURLOPT_HEADERFUNCTION 20079
#define CURLOPT_HEADERDATA 10029

// Common CURLINFO values
#define CURLINFO_RESPONSE_CODE 0x200002
#define CURLINFO_EFFECTIVE_URL 0x100001
#define CURLINFO_TOTAL_TIME 0x300003
#define CURLINFO_NAMELOOKUP_TIME 0x300004
#define CURLINFO_CONNECT_TIME 0x300005
#define CURLINFO_STARTTRANSFER_TIME 0x300006

// Slist type for headers
struct curl_slist {
    char *data;
    struct curl_slist *next;
};

// Core curl functions
CURL *curl_easy_init(void);
void curl_easy_cleanup(CURL *curl);
CURLcode curl_easy_perform(CURL *curl);
const char *curl_easy_strerror(CURLcode code);
struct curl_slist *curl_slist_append(struct curl_slist *list, const char *string);
void curl_slist_free_all(struct curl_slist *list);

// Declaration for curl-impersonate function
int curl_easy_impersonate(CURL *curl, const char *target, int default_headers);

// Wrappers for variadic functions - these are called from Go
CURLcode _curl_easy_setopt_ptr(CURL *curl, CURLoption option, void *param);
CURLcode _curl_easy_setopt_long(CURL *curl, CURLoption option, long param);
CURLcode _curl_easy_getinfo_ptr(CURL *curl, CURLINFO info, void *param);

#endif
