# Key-Value Store Web App

A lightweight key-value store web application built with Go, Gin, SQLite, and Tailwind CSS, running in a Docker container.

## Features

- CRUD operations for key-value pairs
- Search functionality
- Sorting by columns
- Infinite scroll pagination
- Responsive dark-mode UI
- Dockerized deployment
- Data persistence
- Bulk operations (generate dummy data, truncate DB)

## Requirements

- Docker 20.10+
- Docker Compose 2.0+

## Installation

1. Clone the repository:
```bash
git clone https://github.com/karearl/key-value-store.git
cd key-value-store
```

2. Build and start the container:
```bash
docker-compose up --build
```

3. Access the application at:  
`http://localhost:8080`

## Data Persistence

The application uses 2 persistence methods (choose one):

### 1. Docker Volume (Recommended for production)
- Data is stored in a Docker volume named `kvstore-data`
- Volume is automatically created on first run
- To delete all data: `docker-compose down -v`

### 2. Local Directory (Recommended for development)
1. Create a `data` directory in the project root:
```bash
mkdir data
```

2. Modify `docker-compose.yml`:
```yaml
volumes:
  - ./data:/app/data
```

## Configuration

Environment variables (optional, add to `docker-compose.yml`):
```yaml
environment:
  - GIN_MODE=release
  - PORT=8080
```

## Usage

### Docker Commands
- Start: `docker-compose up -d`
- Stop: `docker-compose down`
- Rebuild: `docker-compose up --build -d`
- View logs: `docker-compose logs -f`

### API Endpoints
| Method | Endpoint                    | Description                 |
|--------|-----------------------------|-----------------------------|
| GET    | /api/entries                | List entries                |
| POST   | /api/entries                | Create entry                |
| GET    | /api/entries/{id}           | Get single entry            |
| PUT    | /api/entries/{id}           | Update entry                |
| DELETE | /api/entries/{id}           | Delete entry                |
| POST   | /api/entries/generate-dummy | Generate 1000 dummy entries |
| POST   | /api/entries/truncate       | Delete all entries          |

## Development

1. Install Go 1.21+ and Node.js 16+
2. Install Tailwind CSS:
```bash
npm install -D tailwindcss
npx tailwindcss -i ./static/css/input.css -o ./static/css/styles.css --watch
```

3. Run the Go application:
```bash
go run main.go
```

## License

MIT License