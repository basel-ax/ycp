# AI Agent Instructions for YouTube Stream Comments Processor

## Overview
This document provides instructions for an AI Agent to understand, set up, and run the YouTube Stream Comments Processor application.

## Project Description
The YouTube Stream Comments Processor is a Golang console application that reads comments from a YouTube stream, processes them based on a configuration file, and interacts with Redis to track button presses. The application displays statistics and handles various edge cases such as time limits, command limits, and specific final comments.

## Setup Instructions

### Prerequisites
1. **Golang**: Ensure Golang is installed on your system. You can download it from [Golang's official website](https://golang.org/dl/).
2. **Redis**: Ensure Redis is installed and running. You can download it from [Redis's official website](https://redis.io/download).
3. **Dependencies**: The project uses the following dependencies:
   - `github.com/joho/godotenv` for loading environment variables.
   - `github.com/go-redis/redis/v8` for Redis integration.
   - `github.com/alicebob/miniredis/v2` for testing Redis integration.

### Installation Steps
1. **Clone the Repository**:
   ```bash
   git clone https://github.com/basel-ax/ycp.git
   cd ycp
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Configure the Application**:
   - Copy the `example.env` file to `.env`:
     ```bash
     cp example.env .env
     ```
   - Edit the `.env` file to set your desired configuration values.

4. **Build the Application**:
   ```bash
   go build -o ycp
   ```

## Running the Application

### Start the Application
To start the application, run the following command:
```bash
./ycp
```

### Application Flow
1. **Home Screen**: The application will display a home screen with the configured buttons and parameters.
2. **Start Processing**: Press Enter to clear the console and start reading comments from the stream.
3. **Processing Comments**: The application will process comments from the stream, update Redis with button counts, and log comments.
4. **Final Screen**: The application will display a final statistics screen when the total limit, time limit, or FINAL_COMMENT condition is met.

### Configuration Options
The application can be configured using the `.env` file. Here are the available configuration options:

- **Button Codes**: Define button codes and their corresponding words (e.g., `BUTTON_WW=w`).
- **Total Limit**: Set the total limit on transmitted commands (`TOTAL_LIMIT=100`).
- **Time Limit**: Set the time limit for completion in seconds (`TIME_LIMIT=3600`).
- **Final Comment**: Set the FINAL_COMMENT to trigger early termination (`FINAL_COMMENT="exit"`).
- **API Connection**: Set the API connection details (`API_CONNECTION=""`). If empty, the application will use mock data.
- **Redis Connection**: Set the Redis connection details (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).

### Example Configuration
Here is an example configuration file (`example.env`):
```env
# Button codes and their corresponding words
BUTTON_WW=w
BUTTON_HH=h
BUTTON_AA=a
BUTTON_TT=t
BUTTON_SPACE= 
BUTTON_DOT=.
BUTTON_QUESTION=?
BUTTON_EXCLAMATION=!

# Total limit on transmitted commands
TOTAL_LIMIT=100

# Time limit for completion (in seconds)
TIME_LIMIT=3600

# FINAL_COMMENT to trigger early termination
FINAL_COMMENT="exit"

# API connection details (leave empty to use mock)
API_CONNECTION=""

# Redis connection details
REDIS_HOST="localhost"
REDIS_PORT="6379"
REDIS_PASSWORD=""
REDIS_DB=0
```

## Testing the Application

### Running Tests
To run the tests, use the following command:
```bash
go test -v
```

### Test Cases
The application includes the following test cases:
1. **Config Loading**: Tests the loading and parsing of the configuration file.
2. **Comment Processing**: Tests the processing of comments and updating of statistics.
3. **Redis Integration**: Tests the interaction with Redis for storing and retrieving button counts.
4. **Comment Reading**: Tests the reading of comments from the stream.

## Troubleshooting

### Common Issues
1. **Redis Connection Issues**: Ensure Redis is running and the connection details in the `.env` file are correct.
2. **Dependency Issues**: Ensure all dependencies are installed using `go mod tidy`.
3. **Configuration Errors**: Ensure the `.env` file is correctly formatted and all required fields are set.

### Debugging
- **Logs**: The application logs comments and events to `comments.log`. Check this file for debugging information.
- **Console Output**: The application provides detailed console output during execution. Use this to identify issues.

## Conclusion
This document provides comprehensive instructions for setting up, running, and testing the YouTube Stream Comments Processor application. Follow these instructions to ensure a smooth and successful deployment.