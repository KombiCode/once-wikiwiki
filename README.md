# Once WikiWiki

A simple wiki written in Go using SQLite.

## Features

- SQLite-backed page storage (database at `/storage/wiki.db`)
- `[[Wiki Links]]` syntax for linking pages
- `/up` health check endpoint (returns 200 OK)
- Dockerized for easy deployment

## Running with Docker

```bash
docker build -t once-wikiwiki .
docker run -p 8080:8080 -v $(pwd)/storage:/storage once-wikiwiki
```

Or use Docker Compose:

```bash
docker-compose up
```

## Usage

- Visit `http://localhost:8080` for the home page
- Navigate to `/view/PageName` to view a page
- Navigate to `/edit/PageName` to edit a page
- Use `[[PageName]]` syntax in page content to create links

## Health Check

```bash
curl http://localhost:8080/up
```
