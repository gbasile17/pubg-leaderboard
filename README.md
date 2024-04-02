```
# PUBG Leaderboard Data Manager

This application serves as a data manager for PUBG (PlayerUnknown's Battlegrounds) leaderboard statistics. It provides services to load leaderboard data into Redis, backup data to MinIO, and restore from MinIO.

## Features

- Load current PUBG leaderboard data into Redis.
- Periodically refresh leaderboard data from the PUBG API.
- Backup leaderboard data to MinIO object storage.
- Restore leaderboard data from MinIO into Redis.
- Provide a RESTful API to interact with the service.

## Prerequisites

- Go (version 1.15 or higher)
- Redis server
- MinIO server
- Access to PUBG API with an API key

## Installation

1. Clone the repository to your local machine:
   ```shell
   git clone https://github.com/your-username/pubg-leaderboard.git
   ```
2. Navigate into the project directory:
   ```shell
   cd pubg-leaderboard
   ```
3. Build the application:
   ```shell
   go build -o pubg-leaderboard
   ```

## Docker Build
   ```shell
   docker build -t pubg-leaderboard:latest .
   ```

## Configuration

Before running the application, configure the necessary environment variables or `config.json` file with the following settings:

- `REDIS_ADDR`: The address of your Redis server.
- `MINIO_ENDPOINT`: The endpoint for your MinIO server.
- `MINIO_ACCESS_KEY`: Your MinIO access key.
- `MINIO_SECRET_KEY`: Your MinIO secret key.
- `PUBG_API_KEY`: Your API key for the PUBG API.

## Running the Application

After configuration, you can start the application by running:

```shell
./pubg-leaderboard
```

## API Endpoints

The application exposes the following RESTful endpoints:

- `GET /ping`: Health check for the application.
- `GET /redis-ping`: Check the connection to the Redis server.
- `GET /current-season`: Get the current PUBG season data.
- `GET /current-leaderboard`: Get the current PUBG leaderboard.
- `GET /player-stats/:playerID`: Get specific stats for a player by their ID.
- `POST /backup-leaderboard`: Backup the current leaderboard to MinIO.
- `POST /restore-leaderboard`: Restore the leaderboard from a MinIO backup.

## Contributing

If you'd like to contribute to the project, please fork the repository and use a feature branch. Pull requests are warmly welcome.


## Contact

Feel free to contact project maintainers for any inquiries (gregymann@gmail.com).
```
