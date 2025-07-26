## false-fact-server

the backend for the false fact extension

### Endpoints

- POST `/analyze/article` — for articles `{ "content": "content", "title": "Title", "url": "something.com", "last_edited": "2025-07-25T18:05:27.849Z" }`
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

