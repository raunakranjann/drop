package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/skip2/go-qrcode"
)

const receiveHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Drop — Upload to Linux</title>
  <style>
    :root {
      --bg: #090d16;
      --card-bg: rgba(22, 30, 49, 0.75);
      --accent: #38bdf8;
      --accent-glow: rgba(56, 189, 248, 0.35);
      --text: #f8fafc;
      --muted: #94a3b8;
      --border: rgba(255, 255, 255, 0.08);
      --success: #34d399;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
    body {
      background: radial-gradient(circle at top, #1e293b, var(--bg));
      color: var(--text);
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      padding: 16px;
    }
    .card {
      background: var(--card-bg);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid var(--border);
      border-radius: 24px;
      padding: 32px 24px;
      width: 100%;
      max-width: 400px;
      text-align: center;
      box-shadow: 0 20px 40px rgba(0,0,0,0.6);
    }
    .icon {
      width: 64px;
      height: 64px;
      background: rgba(56, 189, 248, 0.1);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 16px;
      color: var(--accent);
    }
    h2 { font-size: 1.5rem; font-weight: 700; margin-bottom: 6px; letter-spacing: -0.5px; }
    p { color: var(--muted); font-size: 0.9rem; margin-bottom: 24px; }
    .dropzone {
      border: 2px dashed rgba(56, 189, 248, 0.3);
      border-radius: 16px;
      padding: 28px 16px;
      cursor: pointer;
      background: rgba(15, 23, 42, 0.5);
      transition: all 0.2s ease;
    }
    .dropzone.dragover {
      border-color: var(--accent);
      background: rgba(56, 189, 248, 0.08);
      transform: scale(1.02);
    }
    input[type=file] { display: none; }
    .btn {
      display: inline-block;
      margin-top: 12px;
      padding: 10px 20px;
      background: var(--accent);
      color: #030712;
      border-radius: 10px;
      font-size: 0.9rem;
      font-weight: 600;
    }
    .progress-container { display: none; margin-top: 24px; }
    .progress-bar {
      height: 8px;
      background: rgba(255, 255, 255, 0.08);
      border-radius: 999px;
      overflow: hidden;
      margin-top: 8px;
    }
    .progress-fill {
      height: 100%;
      width: 0%;
      background: var(--accent);
      box-shadow: 0 0 12px var(--accent-glow);
      transition: width 0.1s;
    }
    .status-text {
      font-size: 0.85rem;
      margin-top: 10px;
      color: var(--muted);
      word-break: break-all;
    }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">
      <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
        <polyline points="17 8 12 3 7 8"></polyline>
        <line x1="12" y1="3" x2="12" y2="15"></line>
      </svg>
    </div>
    <h2>Drop to Linux</h2>
    <p>Upload files directly to your machine</p>
    
    <div class="dropzone" id="dz" onclick="document.getElementById('fi').click()">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2">
        <path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"></path>
        <path d="M12 12v9"></path>
        <path d="m16 16-4-4-4 4"></path>
      </svg>
      <div style="font-size: 0.85rem; margin-top: 8px; color: #cbd5e1;">Tap to select or drop file</div>
      <span class="btn">Choose File</span>
      <input type="file" id="fi" onchange="handleUpload(this.files[0])">
    </div>

    <div class="progress-container" id="pc">
      <div style="display:flex; justify-content:space-between; font-size:0.8rem;">
        <span id="fname">Uploading...</span>
        <span id="pct">0%</span>
      </div>
      <div class="progress-bar"><div class="progress-fill" id="pf"></div></div>
      <div class="status-text" id="st">Streaming file over local link...</div>
    </div>
  </div>

  <script>
    const dz = document.getElementById('dz');
    ['dragenter', 'dragover'].forEach(n => dz.addEventListener(n, e => { e.preventDefault(); dz.classList.add('dragover'); }));
    ['dragleave', 'drop'].forEach(n => dz.addEventListener(n, e => { e.preventDefault(); dz.classList.remove('dragover'); }));
    dz.addEventListener('drop', e => { if (e.dataTransfer.files.length) handleUpload(e.dataTransfer.files[0]); });

    function handleUpload(file) {
      if (!file) return;
      document.getElementById('dz').style.display = 'none';
      document.getElementById('pc').style.display = 'block';
      document.getElementById('fname').innerText = file.name;

      const fd = new FormData();
      fd.append('file', file);

      const xhr = new XMLHttpRequest();
      xhr.open('POST', '/upload', true);
      xhr.upload.onprogress = e => {
        if (e.lengthComputable) {
          const p = Math.round((e.loaded / e.total) * 100);
          document.getElementById('pf').style.width = p + '%';
          document.getElementById('pct').innerText = p + '%';
        }
      };
      xhr.onload = () => {
        if (xhr.status === 200) {
          document.getElementById('st').innerText = '✅ Transfer complete!';
          document.getElementById('st').style.color = '#34d399';
        } else {
          document.getElementById('st').innerText = '❌ Transfer failed';
          document.getElementById('st').style.color = '#f87171';
        }
      };
      xhr.send(fd);
    }
  </script>
</body>
</html>`

func getOutboundIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err == nil {
		defer conn.Close()
		return conn.LocalAddr().(*net.UDPAddr).IP.String()
	}
	return "127.0.0.1"
}

func main() {
	receiveMode := flag.Bool("r", false, "Receive files from mobile device")
	onlineMode := flag.Bool("online", false, "Use public HTTPS Cloudflare tunnel")
	destDir := flag.String("dir", ".", "Directory to save received files")
	port := flag.Int("port", 8989, "Server port")
	flag.Parse()

	// 1. Establish TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		fmt.Printf("⚠️ Port %d is currently occupied by an old process.\n", *port)
		fmt.Printf("Run this command to clear it: sudo fuser -k %d/tcp\n", *port)
		os.Exit(1)
	}

	var baseURL string
	var tunnelCmd *exec.Cmd

	// Set up process cleanup helper
	cleanupProcesses := func() {
		if tunnelCmd != nil && tunnelCmd.Process != nil {
			_ = tunnelCmd.Process.Kill()
		}
	}

	// Trap Ctrl+C (SIGINT) and kill signals (SIGTERM)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		fmt.Println("\n🛑 Stopping drop and releasing port...")
		cleanupProcesses()
		_ = listener.Close()
		os.Exit(0)
	}()

	if *onlineMode {
		fmt.Println("⚡ Establishing secure HTTPS tunnel...")
		tunnelCmd = exec.Command("cloudflared", "tunnel", "--no-autoupdate", "--url", fmt.Sprintf("http://127.0.0.1:%d", *port))
		stderr, err := tunnelCmd.StderrPipe()
		if err != nil {
			fmt.Println("Tunnel error:", err)
			cleanupProcesses()
			os.Exit(1)
		}
		if err := tunnelCmd.Start(); err != nil {
			fmt.Println("cloudflared error:", err)
			cleanupProcesses()
			os.Exit(1)
		}

		re := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if match := re.FindString(line); match != "" {
				baseURL = match
				break
			}
		}
		if baseURL == "" {
			fmt.Println("Failed to retrieve tunnel URL.")
			cleanupProcesses()
			os.Exit(1)
		}
	} else {
		baseURL = fmt.Sprintf("http://%s:%d", getOutboundIP(), *port)
	}

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	// Helper to perform a graceful zero-hang shutdown
	gracefulShutdown := func() {
		go func() {
			time.Sleep(1 * time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
			cleanupProcesses()
			os.Exit(0)
		}()
	}

	// ===========================
	// MODE 1: RECEIVE FILES
	// ===========================
	if *receiveMode {
		absDir, _ := filepath.Abs(*destDir)
		_ = os.MkdirAll(absDir, 0755)

		qr, _ := qrcode.New(baseURL+"/", qrcode.Medium)
		fmt.Println("\n" + qr.ToSmallString(false))
		fmt.Println("📥 MODE: Receive from Mobile")
		fmt.Printf("💾 Saving to: %s\n", absDir)
		fmt.Printf("🔗 Link:      %s/\n", baseURL)
		fmt.Println("📱 Scan QR with phone to upload...")

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(receiveHTML))
		})

		var once sync.Once
		mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			mr, err := r.MultipartReader()
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil || part.FileName() == "" {
					continue
				}

				savePath := filepath.Join(absDir, filepath.Base(part.FileName()))
				dest, err := os.Create(savePath)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				fmt.Printf("\n📥 Receiving: %s\n", filepath.Base(savePath))
				bar := progressbar.DefaultBytes(-1, "Receiving")
				_, _ = io.Copy(io.MultiWriter(dest, bar), part)
				dest.Close()

				fmt.Printf("\n✅ Saved to: %s\n", savePath)
			}

			w.WriteHeader(http.StatusOK)
			once.Do(gracefulShutdown)
		})

		_ = server.Serve(listener)
		return
	}

	// ===========================
	// MODE 2: SEND FILE
	// ===========================
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("  drop <file>            Send over local Wi-Fi / Hotspot")
		fmt.Println("  drop -online <file>    Send over global Cloudflare HTTPS")
		fmt.Println("  drop -r [-dir <path>]  Receive files to directory")
		cleanupProcesses()
		os.Exit(1)
	}

	filePath := args[0]
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("File not found: %s\n", filePath)
		cleanupProcesses()
		os.Exit(1)
	}
	fileName := filepath.Base(filePath)
	downloadURL := fmt.Sprintf("%s/download", baseURL)

	qr, _ := qrcode.New(downloadURL, qrcode.Medium)
	fmt.Println("\n" + qr.ToSmallString(false))
	fmt.Printf("📦 File:  %s (%.2f MB)\n", fileName, float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("🔗 Link:  %s\n", downloadURL)
	fmt.Println("📱 Scan with phone to download...")

	var once sync.Once
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", strings.ReplaceAll(fileName, "\"", "")))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Read error", 500)
			return
		}
		defer file.Close()

		fmt.Printf("\n📱 Connected by %s. Streaming:\n", r.RemoteAddr)

		bar := progressbar.NewOptions64(
			fileInfo.Size(),
			progressbar.OptionSetDescription("Sending"),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowCount(),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		)

		_, _ = io.Copy(io.MultiWriter(w, bar), file)

		once.Do(gracefulShutdown)
	})

	_ = server.Serve(listener)
}
