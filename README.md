## false-fact-server

the backend for the false fact extension

### Endpoints

- `/` — root, says hi
- `/api` — main API endpoint
- `/health` — health check

### Control Script

`server-control.sh` start/stop/restart/check status of the server:

```sh
./server-control.sh start (optional: [port])
./server-control.sh stop
./server-control.sh restart (optional: [port])
./server-control.sh status
```

