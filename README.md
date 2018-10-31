# gcs-helper

[![Build Status](https://travis-ci.org/NYTimes/gcs-helper.svg?branch=master)](https://travis-ci.org/NYTimes/gcs-helper)
[![codecov](https://codecov.io/gh/NYTimes/gcs-helper/branch/master/graph/badge.svg)](https://codecov.io/gh/NYTimes/gcs-helper)

gcs-helper is inspired by
[s3-helper](https://github.com/crunchyroll/evs-s3helper) and is used to provide
access to private GCS buckets.

It was designed to be used with [Kaltura's
nginx-vod-module](https://github.com/kaltura/nginx-vod-module), but it can work
stand-alone too, specially the proxy feature.

Specific to nginx-vod-module, gcs-helper provides support for the mapped mode
(when using the proper environment variables - ``GCS_HELPER_PROXY_PREFIX``,
``GCS_HELPER_MAP_PREFIX`` and ``GCS_HELPER_MAP_REGEX_FILTER``).

## Configuration

The following environment variables control the behavior of gcs-helper:

| Variable                         | Default value | Required | Description                                                                                                  |
| -------------------------------- | ------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GCS_HELPER_LISTEN                | :8080         | No       | Address to bind the server                                                                                                                                               |
| GCS_HELPER_BUCKET_NAME           |               | Yes      | Name of the bucket                                                                                                                                                       |
| GCS_HELPER_LOG_LEVEL             | debug         | No       | Logging level                                                                                                                                                           |
| GCS_HELPER_PROXY_PREFIX          |               | No       | Prefix to use for the proxy binding. Required if running in map and proxy modes (example value: ``/proxy/``)                                                        |
| GCS_HELPER_PROXY_TIMEOUT         | 10s           | No       | Defines the maximum time in serving the proxy requests, this is a hard timeout and includes retries                                                                    |
| GCS_HELPER_MAP_PREFIX            |               | No       | Prefix to use for the map binding. Required if running in map and proxy modes (example value: ``/map/``)                                                                |
| GCS_HELPER_MAP_REGEX_FILTER      |               | No       | A regular expression that is used to deliver only those files that match the specified naming convention (example value: ``\d{3,4}p(\.mp4\|[a-z0-9_-]{37}\.(vtt\|srt))$``) |
| GCS_HELPER_EXTRA_RESOURCES_TOKEN |               |          | Token to be used as query string parameter on the map location to pass extra resources to the mapping                                                                  |
| GCS_HELPER_MAP_EXTRA_PREFIXES    |               | No       | Comma separated list of prefixes that allow gcs-helper to lookup files in different paths                                                                              |
| GCS_HELPER_MAP_EXTENSION_SPLIT   | false         | No       | Boolean flag that indicates whether extensions in the path should be stripped from the prefix and used as a suffix                                                     |

The are also some configuration variables for network communication with Google
Cloud Storage API:

| Variable                     | Default value | Required | Description                                                                                                  |
| ---------------------------- | ------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| GCS_CLIENT_TIMEOUT           | 2s            | No       | Hard timeout on requests that gcs-helper sends to the Google Storage API                                     |
| GCS_CLIENT_IDLE_CONN_TIMEOUT | 120s          | No       | Maximum duration of idle connections between gcs-helper and the Google Storage API                           |
| GCS_CLIENT_MAX_IDLE_CONNS    | 10            | No       | Maximum number of idle connections to keep open. This doesn't control the maximum number of connections      |

### GCS_HELPER_PROXY_TIMEOUT x GCS_CLIENT_TIMEOUT

The timeout configuration is mainly controlled by two environment variables:
``GCS_HELPER_PROXY_TIMEOUT`` and ``GCS_CLIENT_TIMEOUT``. The
``GCS_HELPER_PROXY_TIMEOUT`` controls how long requests to gcs-helper can take,
and ``GCS_CLIENT_TIMEOUT`` controls how long requests from gcs-helper to
Google's API can take. Since gcs-helper automatically retries on failures, the
number of retries is roughly the value of ``GCS_HELPER_PROXY_TIMEOUT`` divided
by the value of ``GCS_CLIENT_TIMEOUT``.

### GCS_HELPER_EXTRA_RESOURCES_TOKEN

The extra resources token is the query string parameter that the mapping location
will look for to add external resources to the JSON output.

For example, giving that you set `GCS_HELPER_EXTRA_RESOURCES_TOKEN` to `extras`,
you'll be able to add videos and captions files that are in different bucket
by calling the map location with `?extras=/bucket-1/file.mp4,/bucket-2/pt-br.vtt`.
