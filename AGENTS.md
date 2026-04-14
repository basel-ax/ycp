# AGENTS.md

A simple, open format for guiding coding agents on the YouTube Stream Comments Processor project.

## Project Overview
The YouTube Stream Comments Processor is a Golang console application that reads comments from a YouTube stream (mock or real via YouTube Data API v3), processes them based on a configuration file, and interacts with Redis to track letter counts. The application checks for double letters/symbols in comments, compares them with a FINAL_COMMENT string, and updates counts in Redis. When a count exceeds REDIS_COUNT, it resets the count, increases the total limit, and prints the triggering letter. The application displays statistics and handles various edge cases such as time limits, command limits, and specific final comments.

## Build and Test Commands
- Install Go dependencies: `go mod tidy`
- Format code: `go fmt ./...`
- Run all tests: `go test -v ./...`
- Run tests for specific package: `go test -v ./package`
- Build application: `go build -o ycp`
- Run in development mode (prints to console): `./ycp -dev`
- Run in normal mode (logs to file): `./ycp`

## Code Style Guidelines
- Use `go fmt` for formatting (run `go fmt ./...` before committing)
- Group imports: standard library, then third-party, then local packages
- Prefer explicit error handling: check errors, don't use `_` when action is needed
- Use meaningful, descriptive names for variables and functions
- Functions should do one thing well and be reasonably sized (<50 lines ideal)
- Use interfaces for dependency injection (especially for Redis, Logger components)
- Comment only when necessary (complex logic, non-obvious behavior)
- Follow Go naming conventions: MixedCaps, no underscores for exported names
- Keep functions focused and avoid deep nesting
- Use struct composition over inheritance patterns
- Handle errors close to where they occur when possible

## Testing Instructions
- Test files named *_test.go and colocated with source code
- Redis integration tests use alicebob/miniredis/v2 for mock server
- Aim for table-driven tests when testing multiple test cases
- Mock external dependencies (Redis, HTTP clients) in unit tests
- Integration tests verify real Redis interaction when applicable

## Project Structure
- main.go: Application entry point containing core logic and YouTube integration
- config/: Configuration loading and validation functionality
- redis/: Redis client wrapper with interface abstraction
- ui/: Console display functions (home screen, final screen, clear console)

## YouTube Integration Notes
- When STREAM_URL is empty: Uses built-in mock comment stream for testing
- When STREAM_URL is set and YOUTUBE_API_KEY provided: Uses YouTube Data API v3 for real stream
- Supported YouTube URL formats:
  - https://www.youtube.com/watch?v=VIDEO_ID
  - https://youtu.be/VIDEO_ID
  - https://www.youtube.com/live/VIDEO_ID
  - Just the VIDEO_ID
- Fallback to mock on any API error or configuration issue
- Video ID extraction handles various URL formats robustly

## Configuration
- Loaded from .env file via github.com/joho/godotenv package
- All configuration variables have sensible defaults
- Config validation ensures TotalLimit, TimeLimit, and RedisCount are > 0
- StreamURL: YouTube stream URL or video ID (empty = mock mode)
- YouTubeAPIKey: YouTube Data API v3 key (required for real stream integration)
- APIConnection: Reserved for future API integration features