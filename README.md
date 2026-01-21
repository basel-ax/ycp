# YouTube Stream Comments Processor

## Overview
This is a Golang console application that reads comments from a YouTube stream, processes them based on a configuration file, and interacts with Redis to track button presses. The application displays statistics and handles various edge cases such as time limits, command limits, and specific final comments.

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

## Running the Application
To start the application, run the following command:
```bash
./ycp
```

## Configuration
The application can be configured using the `.env` file. Here are the available configuration options:

- **Button Codes**: Define button codes and their corresponding words (e.g., `BUTTON_WW=w`).
- **Total Limit**: Set the total limit on transmitted commands (`TOTAL_LIMIT=100`).
- **Time Limit**: Set the time limit for completion in seconds (`TIME_LIMIT=3600`).
- **Final Comment**: Set the FINAL_COMMENT to trigger early termination (`FINAL_COMMENT="exit"`).
- **API Connection**: Set the API connection details (`API_CONNECTION=""`). If empty, the application will use mock data.
- **Redis Connection**: Set the Redis connection details (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).

## Testing
To run the tests, use the following command:
```bash
go test -v
```

## Troubleshooting
- **Redis Connection Issues**: Ensure Redis is running and the connection details in the `.env` file are correct.
- **Dependency Issues**: Ensure all dependencies are installed using `go mod tidy`.
- **Configuration Errors**: Ensure the `.env` file is correctly formatted and all required fields are set.

## License
This project is licensed under the MIT License.
