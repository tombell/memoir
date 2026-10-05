# Memoir

Command-line tools and API server for hosting track lists for DJ mixes.

## Request limits and validation

Tracklist POST and PATCH requests require `Content-Type: application/json`.
Media type parameters such as `charset=utf-8` are accepted. JSON bodies must
contain one complete value and may be at most 1 MiB, including whitespace.
Malformed or trailing JSON returns 400, unsupported content types return 415,
and oversized JSON bodies return 413. This limit does not apply to multipart
artwork uploads.

Tracklist validation returns 422 with errors keyed by field. Tracks retain the
array format `[name, artist, bpm, key, genre]`. Each row must have exactly five
strings. Name and artist are required and may have at most 256 characters,
genre at most 128, and key at most 8. BPM must be a finite number. Key and genre
may be empty; no particular musical-key notation is required. Artwork is
required and may have at most 256 characters. Row errors use keys such as
`tracks[0].bpm`.

List requests default to `page=1` and `per_page=10` when those parameters are
omitted. Supplied values must be positive integers, and `per_page` may be at
most 100. Page combinations whose offsets exceed 2147483647 return 400, as do
empty, malformed, or out-of-range values. Search and most-played tracks apply
these query checks but still use their existing limit-only behavior.
