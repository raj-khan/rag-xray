# Kestrel API Reference

Kestrel is Orbita Labs' public API for satellite telemetry. This page covers authentication, limits and errors.

## Authentication

Every request needs an API key in the `X-Kestrel-Key` header. API keys expire and must be rotated every 90 days. Keys are created and revoked in the Kestrel console.

## Rate limits

Standard plan customers can send 600 requests per minute. Enterprise plan customers can send 3000 requests per minute. Limits apply per API key, not per account.

When a limit is exceeded the API returns HTTP 429 with error code KST-429 and a `Retry-After` header that says how many seconds to wait.

## Error codes

- KST-400: the request body is invalid. The response explains which field is wrong.
- KST-401: the API key is missing, expired or revoked.
- KST-429: the rate limit was exceeded.
- KST-503: the API is in scheduled maintenance. Maintenance windows are Sundays from 02:00 to 04:00 UTC.

## Retries

Clients should retry KST-429 and KST-503 errors with exponential backoff, starting at 2 seconds and doubling each time, with at most 5 retries. Other errors should not be retried.

## Pagination

List endpoints return at most 100 items per page. Use the `next_cursor` value from the response to fetch the next page.
