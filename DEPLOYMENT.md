# 🚀 Ubuntu/EC2 Deployment Guide

This guide shows how to deploy the Incident Commander Game to an Ubuntu server or AWS EC2 instance.

## 🎯 Quick Deployment (One Command)

For a fresh Ubuntu server, run this single command:

```bash
make deploy-ubuntu
```

This will:
- Install Go and dependencies
- Build optimized production binaries
- Install as a systemd service
- Configure firewall
- Start the service
- Show your public IP and access URLs

## 📋 Step-by-Step Deployment

### 1. **Prepare Ubuntu Server**

```bash
# Install dependencies
make ubuntu-deps

# If Go was just installed, reload your shell
source ~/.bashrc
```

### 2. **Build for Production**

```bash
# Build optimized WebAssembly and Linux binary
make build-ubuntu
```

### 3. **Choose Deployment Method**

#### Option A: Simple Background Process
```bash
# Start server in background
make run-daemon

# Check status
make status

# Stop server
make stop-daemon
```

#### Option B: Systemd Service (Recommended)
```bash
# Install as system service
make install-service

# Start service
make service-start

# Check status
make service-status

# View logs
make service-logs
```

### 4. **Configure Firewall (if needed)**

```bash
make setup-firewall
```

## 🌐 Access Your Game

After deployment, your game will be available at:
- **Game URL**: `http://YOUR_SERVER_IP:8080`
- **Health Check**: `http://YOUR_SERVER_IP:8080/health`

To find your server's public IP:
```bash
curl ifconfig.me
```

## 🔧 Service Management

### Systemd Service Commands
```bash
make service-start    # Start the service
make service-stop     # Stop the service  
make service-restart  # Restart the service
make service-status   # Check service status
make service-logs     # View real-time logs
```

### Daemon Process Commands
```bash
make run-daemon     # Start in background
make status         # Check daemon status
make stop-daemon    # Stop daemon process
```

## 📊 Monitoring & Health Checks

### Check Server Status
```bash
# Service status
make service-status

# Daemon status
make status

# Health endpoint
curl http://localhost:8080/health
```

### View Logs
```bash
# Systemd service logs
make service-logs

# Daemon logs (if using daemon mode)
tail -f incident-commander.log
```

## 🛠️ Troubleshooting

### Common Issues

**1. Port 8080 already in use**
```bash
# Find what's using port 8080
sudo lsof -i :8080

# Kill the process if needed
sudo kill $(sudo lsof -t -i:8080)
```

**2. Permission denied**
```bash
# Make binary executable
chmod +x incident-commander-server
```

**3. Go not found after installation**
```bash
# Reload shell environment
source ~/.bashrc

# Or manually add to PATH
export PATH=$PATH:/usr/local/go/bin
```

**4. Firewall blocking access**
```bash
# Configure UFW firewall
make setup-firewall

# Or manually open port
sudo ufw allow 8080/tcp
```

### Service Debugging
```bash
# Check service status
sudo systemctl status incident-commander.service

# View detailed logs
sudo journalctl -u incident-commander.service -n 50

# Restart service
make service-restart
```

## 🔐 Security Considerations

### Basic Security Setup
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Configure firewall
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 8080/tcp
```

### Optional: Use Reverse Proxy
For production, consider using nginx as a reverse proxy:

```bash
# Install nginx
sudo apt install -y nginx

# Configure nginx (create /etc/nginx/sites-available/incident-commander)
sudo tee /etc/nginx/sites-available/incident-commander > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/incident-commander /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 📁 File Structure

After deployment, you'll have these files:
```
incident-commander-game-no-instrumentation/
├── incident-commander-server          # Linux binary
├── incident-commander.pid             # Daemon PID file
├── incident-commander.log             # Daemon logs
├── web/                               # Web assets
│   ├── index.html                     # Game HTML
│   ├── images/o11y_alert.png          # Game sprite
│   └── static/                        # WebAssembly files
│       ├── game.wasm                  # Game logic
│       └── wasm_exec.js               # Go WASM runtime
└── /etc/systemd/system/incident-commander.service  # Service file
```

## 🧹 Cleanup

To remove deployment files:
```bash
# Stop and remove service
make service-stop
sudo systemctl disable incident-commander.service
sudo rm /etc/systemd/system/incident-commander.service
sudo systemctl daemon-reload

# Clean deployment files
make clean-deploy
```

## 🎮 Testing Deployment

After deployment, test these endpoints:

```bash
# Health check
curl http://your-server-ip:8080/health

# Game page (should return HTML)
curl -I http://your-server-ip:8080/

# WebAssembly file (should exist)
curl -I http://your-server-ip:8080/static/game.wasm
```

## 📞 Support

If you encounter issues:

1. **Check logs**: `make service-logs` or `tail -f incident-commander.log`
2. **Verify build**: `make build-ubuntu`
3. **Test locally**: `make run-ubuntu`
4. **Check firewall**: `sudo ufw status`
5. **Verify Go installation**: `go version`

---

**Ready to play!** 🎮 Your Incident Commander game should now be running on Ubuntu and accessible from any web browser!