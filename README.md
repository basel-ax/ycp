# YouTube Stream Comments Processor

## Overview
This is a Golang console application that reads comments from a YouTube stream, processes them based on a configuration file, and interacts with Redis to track letter counts. The application checks for double letters/symbols in comments, compares them with the FINAL_COMMENT, and updates counts in Redis. If a count exceeds REDIS_COUNT, it resets the count, increases the total limit, and prints the letter. The application displays statistics and handles various edge cases such as time limits, command limits, and specific final comments.

## Features
- Read comments from a YouTube stream or use built-in mock data
- Process comments based on a configuration file
- Track button presses in Redis
- **ANSI art final screen** — random image from `graphics/` displayed on the left 2/3 with statistics on the right 1/3
- **Clean production startup** — no console messages in normal mode; all data logged to file
- Support for both real YouTube API and mock data
- Graceful shutdown via ESC key, SIGINT/SIGTERM, time limit, or total limit

## Prerequisites
- Golang
- Redis
- YouTube API (optional, for real API integration)

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/basel-ax/ycp.git
   cd ycp
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure the application:
   - Copy the `example.env` file to `.env`:
     ```bash
     cp example.env .env
     ```
   - Edit the `.env` file to set your desired configuration values.

4. Build the application:
   ```bash
   go build -o ycp
   ```

5. (Optional) Run tests:
   ```bash
   go test -v ./...
   ```

## Prerequisites
- Go 1.25.3 or higher
- Redis server running (for production use)

## Running the Application

### Normal Mode
To start the application in normal mode (logs to file), run the following command:
```bash
./ycp
```

### Development Mode
To start the application in development mode (prints comments to console), run the following command:
```bash
./ycp -dev
```

In development mode, comments are printed directly to the console instead of being logged to a file, and startup log messages are visible. This is useful for debugging and development purposes.

### Production Mode
In normal mode (`./ycp`), the console starts completely empty — no startup messages are shown. All comment data is logged to a timestamped file. The final statistics screen renders a random ANSI graphic from the `graphics/` directory alongside the statistics.

## Configuration
The application can be configured using the `.env` file. Here are the available configuration options:

- **Total Limit**: Set the total limit on transmitted commands (`TOTAL_LIMIT=100`).
- **Time Limit**: Set the time limit for completion in seconds (`TIME_LIMIT=3600`).
- **Final Comment**: Set the FINAL_COMMENT to trigger early termination (`FINAL_COMMENT=exit`).
- **Stream URL**: Set the YouTube stream URL or video ID to read real comments (`STREAM_URL=`). When empty, uses mock data.
- **YouTube API Key**: Set your YouTube Data API v3 key for real stream integration (`YOUTUBE_API_KEY=`). Optional - required only for live stream comments.
- **Redis Connection**: Set the Redis connection details (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).
- **Redis Count**: Set the threshold for resetting letter counts (`REDIS_COUNT=5`).

## Final Statistics Screen
When processing completes (by reaching the total limit, time limit, or detecting the final comment), the application displays a final statistics screen:

- **ANSI Graphic**: A random image from the `graphics/` directory is rendered as ANSI art on the left 2/3 of the terminal.
- **Statistics Panel**: Comments read, letters typed, commands sent, and triggered letters (sorted alphabetically) are displayed on the right 1/3.
- **Fallback Mode**: If no graphics directory or images exist, statistics are shown in plain text.

Add JPG or PNG images to the `graphics/` directory to customize the final screen.

### YouTube Integration Modes

**Mock Mode (Default)**: When `STREAM_URL` is empty, the application uses built-in mock comments for testing.

**Live Stream Mode**: To read real comments from a YouTube live stream:
1. Get a YouTube Data API v3 key from Google Cloud Console
2. Set `STREAM_URL` to a YouTube live video URL or ID (e.g., `https://www.youtube.com/watch?v=VIDEO_ID` or just `VIDEO_ID`)
3. Set `YOUTUBE_API_KEY` to your API key

Supported URL formats:
- `https://www.youtube.com/watch?v=VIDEO_ID`
- `https://youtu.be/VIDEO_ID`
- `https://www.youtube.com/live/VIDEO_ID`
- Just the video ID (`VIDEO_ID`)

## Testing
The application includes comprehensive auto tests covering configuration loading, comment processing, Redis integration, and comment reading.

To run the tests, use the following command:
```bash
go test -v ./...
```

The tests use a mini Redis server (miniredis) for integration testing and mock data for comment processing. Press Ctrl+C at any time to gracefully shut down.

## Troubleshooting
- **Redis Connection Issues**: Ensure Redis is running and the connection details in the `.env` file are correct.
- **Dependency Issues**: Ensure all dependencies are installed using `go mod tidy`.
- **Configuration Errors**: Ensure the `.env` file is correctly formatted and all required fields are set.

## License
This project is licensed under the MIT License.
