## false-fact-server

the backend for the false fact extension

### Run

releases, unzip the tar file, then you can use the control script after setting your .env (instructions for both below) (for linux amd64)
you can run on Windows easily by having the .env file, ignore the script, and you can run the win binary
other platforms build

### Endpoints

- POST `/analyze/article` - for articles - `{ "content": "content", "title": "Title", "url": "something.com", "last_edited": "2025-07-25T18:05:27.849Z" }` is the format
- POST `/analyze/text/short` - for short text - `{ "content": "content" }` is the format
- POST `/analyze/text/long` - for long text - `{ "content": "content" }` is the format
- `/health` - health check

### Environment Variables

Uses the following environment variables:

- `GEMINI_API_KEY` - api key for gemini ai
- `MODEL` - select the model used (Gemini or Pollinations)
- `PORT` - port number to run the server

Create a `.env` file in the project root to set these values.

### Example systemd File

at false-fact-server.service.example