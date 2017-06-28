gcs-helper
==========

[![Build Status](https://travis-ci.org/NYTimes/gcs-helper.svg?branch=master)](https://travis-ci.org/NYTimes/gcs-helper)
[![codecov](https://codecov.io/gh/NYTimes/gcs-helper/branch/master/graph/badge.svg)](https://codecov.io/gh/NYTimes/gcs-helper)

gcs-helper is inspired by
[s3-helper](https://github.com/crunchyroll/evs-s3helper) and is used to provide
access to private GCS buckets.

It's intended to be deployed along with [Kaltura's
nginx-vod-module](https://github.com/kaltura/nginx-vod-module).

It also provides the needed support for the mapped mode (when using the proper
environment variables - ``GCS_HELPER_PROXY_PREFIX``, ``GCS_HELPER_MAP_PREFIX``
and ``GCS_HELPER_MAP_EXTENSIONS``).

Configuration
-------------

The following environment variables control the behavior of gcs-helper:

| Variable                  | Default value | Required | Description                                                                                                  |
| ------------------------- | ------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| GCS_HELPER_LISTEN         | :8080         | No       | Address to bind the server                                                                                   |
| GCS_HELPER_BUCKET_NAME    |               | Yes      | Name of the bucket                                                                                           |
| GCS_HELPER_LOG_LEVEL      | debug         | No       | Logging level                                                                                                |
| GCS_HELPER_PROXY_PREFIX   |               | No       | Prefix to use for the proxy binding. Required if running in map and proxy modes (example value: ``/proxy/``) |
| GCS_HELPER_MAP_PREFIX     |               | No       | Prefix to use for the map binding. Required if running in map and proxy modes (example value: ``/map/``)     |
| GCS_HELPER_MAP_EXTENSIONS |               | No       | Comma separated list of extensions to include in the mapping (example value: ``.mp4,.vtt,.srt``)             |
