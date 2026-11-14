# Changelog

All notable changes to the Go SDK. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow [Semantic Versioning](https://semver.org/). Releases are tagged `sdk/vX.Y.Z`; see [Releasing](README.md#releasing).

## [Unreleased]

## [0.1.0]

### Added

- `New` with options `WithClientCredentials`, `WithToken`, `WithHTTPClient` and `WithPollInterval`.
- `Health`, `AuditLogs`, and `Instances.Create`, `Get`, `List`, `All`, `Update`, `Delete`, `Credentials`, `Logs`, `Query` and `WaitUntilRunning`.
- `APIError` and sentinel errors for every API error code, for `errors.Is`.
- `dbaastest`, an in-memory fake of the API for tests.
- A `User-Agent` header with the SDK version.
