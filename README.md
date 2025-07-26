## false-fact-server

the backend for the false fact extension

### Endpoints

- `/` — root, says hi
- `/api` — main API endpoint
- `/health` — health check

### Control Script

`server-control.sh` start/stop/restart/check status of the server:

```sh
./server-control.sh start
./server-control.sh stop
./server-control.sh restart
./server-control.sh status
```

### Environment Variables

Uses the following environment variables:

- `GEMINI_API_KEY` — api key for gemini ai
- `PORT` — port number to run the server (default: 3088)

Create a `.env` file in the project root to set these values.

