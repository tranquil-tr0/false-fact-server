# False Fact Server

A simple Go API server ready for deployment on your VPS.

## Local Development

1. **Run the server locally:**
   ```bash
   go run main.go
   ```

2. **Build for production:**
   ```bash
   go build -o false-fact-server main.go
   ```

## VPS Deployment Instructions (Shared Environment)

### Prerequisites
- Your VPS with nest CLI configured
- Go binary built and uploaded to your VPS
- **Note**: This setup is optimized for shared VPS environments

### Step-by-Step Deployment

1. **Upload your files to the VPS:**
   ```bash
   # Upload the binary and deployment script
   scp false-fact-server deploy.sh name1@your-vps:/home/name1/
   ```

2. **SSH into your VPS:**
   ```bash
   ssh name1@your-vps
   ```

3. **Make scripts executable:**
   ```bash
   chmod +x false-fact-server deploy.sh
   ```

4. **Start your server (it will auto-find an available port):**
   ```bash
   ./deploy.sh start
   # OR specify a specific port:
   ./deploy.sh start 3001
   ```

5. **Note the port your server is using** (shown in the start output)

6. **Configure the subdomain (replace `api` and use the actual port):**
   ```bash
   nest caddy add api.name1.hackclub.app --proxy localhost:3001
   ```

7. **Reload Caddy:**
   ```bash
   systemctl --user reload caddy
   ```

### Shared VPS Considerations

- **Port Management**: The script automatically finds available ports to avoid conflicts
- **File Security**: PID and log files are stored in your home directory (`~/.false-fact-server.pid`)
- **Resource Sharing**: Be mindful of CPU/memory usage on shared systems
- **Port Range**: If default ports are taken, the script will try incrementally higher ports

### Server Management

- **Start server:** `./deploy.sh start` (auto-finds available port)
- **Start on specific port:** `./deploy.sh start 3001`
- **Stop server:** `./deploy.sh stop`
- **Restart server:** `./deploy.sh restart`
- **Restart on specific port:** `./deploy.sh restart 3002`
- **Check status:** `./deploy.sh status`
- **View logs:** `tail -f ~/false-fact-server.log`

## API Endpoints

Once deployed, your API will be available at: `http://api.name1.hackclub.app`

- `GET /` - Welcome message
- `GET /api` - Main API endpoint
- `POST /api` - POST endpoint
- `GET /health` - Health check

## Example API Usage

```bash
# Test your deployed API
curl http://api.name1.hackclub.app/
curl http://api.name1.hackclub.app/api
curl http://api.name1.hackclub.app/health
```

## Troubleshooting

1. **Check if server is running:**
   ```bash
   ./deploy.sh status
   ```

2. **View server logs:**
   ```bash
   tail -f server.log
   ```

3. **Check if port 3000 is in use:**
   ```bash
   netstat -tulpn | grep :3000
   ```

4. **Reload Caddy configuration:**
   ```bash
   systemctl --user reload caddy
   ```
