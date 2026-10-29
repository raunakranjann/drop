#!/usr/bin/env bash
set -e

echo "📦 Installing drop (Universal Linux File Sharing)..."

# Build binary
go build -o drop main.go

# Install to system path
sudo mv drop /usr/local/bin/drop
sudo chmod +x /usr/local/bin/drop

# Allow default port in firewall if ufw is present
if command -v ufw >/dev/null 2>&1; then
    sudo ufw allow 8989/tcp >/dev/null 2>&1 || true
fi

echo "✅ Installed successfully! Run 'drop <file>' or 'drop -r' to start."
