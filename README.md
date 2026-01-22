# YouTube Stream Comments Processor

## Overview
This is a Golang console application that reads comments from a YouTube stream, processes them based on a configuration file, and interacts with Redis to track letter counts. The application checks for double letters/symbols in comments, compares them with the FINAL_COMMENT, and updates counts in Redis. If a count exceeds REDIS_COUNT, it resets the count, increases the total limit, and prints the letter. The application displays statistics and handles various edge cases such as time limits, command limits, and specific final comments.

## Features
- Read comments from a YouTube stream.
- Process comments based on a configuration file.
- Track button presses in Redis.
- Display statistics and handle edge cases.
- Support for both real API and mock data.

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

## Prerequisites
- Golang 1.19 or higher
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

In development mode, comments are printed directly to the console instead of being logged to a file. This is useful for debugging and development purposes.

## Configuration
The application can be configured using the `.env` file. Here are the available configuration options:

- **Total Limit**: Set the total limit on transmitted commands (`TOTAL_LIMIT=100`).
- **Time Limit**: Set the time limit for completion in seconds (`TIME_LIMIT=3600`).
- **Final Comment**: Set the FINAL_COMMENT to trigger early termination (`FINAL_COMMENT="exit"`).
- **API Connection**: Set the API connection details (`API_CONNECTION=""`). If empty, the application will use mock data.
- **Redis Connection**: Set the Redis connection details (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).
- **Redis Count**: Set the threshold for resetting letter counts (`REDIS_COUNT=5`).

## Testing
The application includes comprehensive auto tests covering configuration loading, comment processing, Redis integration, and comment reading.

To run the tests, use the following command:
```bash
go test -v
```

The tests use a mini Redis server for integration testing and mock data for comment processing.

## Troubleshooting
- **Redis Connection Issues**: Ensure Redis is running and the connection details in the `.env` file are correct.
- **Dependency Issues**: Ensure all dependencies are installed using `go mod tidy`.
- **Configuration Errors**: Ensure the `.env` file is correctly formatted and all required fields are set.

## License
This project is licensed under the MIT License.
