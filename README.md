# Schedule Job Server

[![Docker Image Build With Push](https://github.com/schedule-job/schedule-job-server/actions/workflows/docker-image-build-push.yml/badge.svg)](https://github.com/schedule-job/schedule-job-server/actions/workflows/docker-image-build-push.yml) [![Docker Pulls](https://img.shields.io/docker/pulls/sotaneum/schedule-job-server?logoColor=fff&logo=docker)](https://hub.docker.com/r/sotaneum/schedule-job-server) [![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/schedule-job/schedule-job-server?logo=go&logoColor=fff)](https://go.dev/) [![Docker Image Size (tag)](https://img.shields.io/docker/image-size/sotaneum/schedule-job-server/latest?logoColor=fff&logo=docker)](https://hub.docker.com/r/sotaneum/schedule-job-server) [![postgresql](https://img.shields.io/badge/14_or_higher-blue?logo=postgresql&logoColor=fff&label=PostgreSQL)](https://www.postgresql.org/)

## Auth

- OAuth 방식을 지원합니다.

### OAuth 2.0 Github Apps

```go
// Example for Github App
oauth.Core.AddProvider("github", &oauth.Github{
  ClientId:             ":Client ID:",
  ClientSecret:         ":Client Secret:",
  RedirectUrl:          ":Callback URL:",
  GithubAccessTokenAPI: "https://github.com/login/oauth/access_token",
  GithubUserAPI:        "https://api.github.com/user",
  GithubLoginUrl:       "https://github.com/login/oauth/authorize",
})
```
