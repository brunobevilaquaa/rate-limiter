# Rate Limiter

This is an usage example of a rate limiter in Go, which limits the number of requests per second based on a specific IP address or an access token. Follow the steps below to run the project and test the rate limiter.

## Prerequisites

- Docker and Docker Compose

## Setup

### 1. Clone the Repository

Clone the repository to your local machine:

```sh
git clone https://github.com/brunobevilaquaa/rate-limiter.git
cd rate-limiter
```

### 2. Docker Compose

Run the following command in the root directory of the project:

```sh
docker-compose -f docker-compose-example.yml up -d
```

This command will start a Redis instance and an Web Server as specified in the `docker-compose-example.yml` file.

## Example Token

Use the following example token for testing the rate limiter functionality:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwicmF0ZUxpbWl0ZXJUaW1lV2luZG93IjoiMTBzIiwicmF0ZUxpbWl0ZXJDcmVkaXRzUGVyVGltZVdpbmRvdyI6MTAsImlhdCI6MTUxNjIzOTAyMn0.2mMDJ4U1RyX1cjE362AjCY5v7LHVvcLERaP-EiKW1GI
```

This token is a valid JWT token with the following claims:

```json
{
  "sub": "1234567890",
  "name": "John Doe",
  "rateLimiterTimeWindow": "10s",
  "rateLimiterCreditsPerTimeWindow": 10,
  "iat": 1516239022
}
```

## Making Requests

### Request Without Token

To test the rate limiter based on IP address, make repeated requests to the server without including the token. For example:

```sh
curl -X GET http://localhost:8080/hello
```

If the number of requests exceeds the IP rate limit (default is 5 requests each 10 second), you will receive an HTTP 429 status code with the message:

```
you have reached the maximum number of requests or actions allowed within a certain time frame
```

### Request With Token

To test the rate limiter based on the access token, include the token in the header of your requests. For example:

```sh
curl -X GET http://localhost:8080/hello -H "API_KEY: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwicmF0ZUxpbWl0ZXJUaW1lV2luZG93IjoiMTBzIiwicmF0ZUxpbWl0ZXJDcmVkaXRzUGVyVGltZVdpbmRvdyI6MTAsImlhdCI6MTUxNjIzOTAyMn0.2mMDJ4U1RyX1cjE362AjCY5v7LHVvcLERaP-EiKW1GI"
```

If the number of requests exceeds the token rate limit (default is 10 requests each 10 second), you will receive an HTTP 429 status code with the same message.