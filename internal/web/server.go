package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/applejobs/telegram-remote-controller/internal/notes"
)

// Server provides the web UI for notes
type Server struct {
	store *notes.Store
	port  int
}

// NewServer creates a new web server
func NewServer(store *notes.Store, port int) *Server {
	return &Server{
		store: store,
		port:  port,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/api/notes", s.handleAPI)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Web UI starting on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	notes := s.store.GetAll()

	tmpl := template.Must(template.New("home").Parse(homeHTML))
	if err := tmpl.Execute(w, notes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		notes := s.store.GetAll()
		json.NewEncoder(w).Encode(notes)

	case http.MethodDelete:
		id := strings.TrimPrefix(r.URL.Query().Get("id"), "")
		if s.store.Delete(id) {
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "Note not found", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

const homeHTML = `<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>💡 Ideas - Telegram Notes</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: #e8e8e8;
            padding: 20px;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        
        header {
            text-align: center;
            padding: 40px 0;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 20px;
            margin-bottom: 30px;
            backdrop-filter: blur(10px);
        }
        
        h1 {
            font-size: 2.5em;
            background: linear-gradient(90deg, #00d4ff, #7b2cbf);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 10px;
        }
        
        .subtitle {
            color: #888;
            font-size: 1.1em;
        }
        
        .stats {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-top: 20px;
        }
        
        .stat {
            background: rgba(255, 255, 255, 0.1);
            padding: 10px 25px;
            border-radius: 30px;
            font-size: 0.9em;
        }
        
        .notes-grid {
            display: grid;
            gap: 20px;
        }
        
        .note-card {
            background: rgba(255, 255, 255, 0.08);
            border-radius: 16px;
            padding: 25px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            transition: transform 0.3s, box-shadow 0.3s;
        }
        
        .note-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 40px rgba(0, 212, 255, 0.2);
        }
        
        .note-content {
            font-size: 1.1em;
            line-height: 1.6;
            margin-bottom: 15px;
            white-space: pre-wrap;
        }
        
        .note-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 0.85em;
            color: #888;
        }
        
        .note-time {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        
        .note-id {
            font-family: monospace;
            background: rgba(255, 255, 255, 0.1);
            padding: 3px 8px;
            border-radius: 5px;
        }
        
        .empty-state {
            text-align: center;
            padding: 60px;
            color: #666;
        }
        
        .empty-state h2 {
            font-size: 3em;
            margin-bottom: 20px;
        }
        
        .tip {
            background: rgba(0, 212, 255, 0.1);
            border-left: 4px solid #00d4ff;
            padding: 15px 20px;
            border-radius: 0 10px 10px 0;
            margin-top: 30px;
            font-size: 0.95em;
        }
        
        code {
            background: rgba(255, 255, 255, 0.15);
            padding: 2px 8px;
            border-radius: 4px;
            font-family: 'SF Mono', Monaco, monospace;
        }
        
        @media (max-width: 600px) {
            body { padding: 10px; }
            h1 { font-size: 1.8em; }
            .note-card { padding: 20px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>💡 Ideas</h1>
            <p class="subtitle">Your thoughts, captured via Telegram</p>
            <div class="stats">
                <div class="stat">📝 {{len .}} ideas</div>
                <div class="stat">⚡ Auto-refresh</div>
            </div>
        </header>
        
        <div class="notes-grid">
            {{if .}}
                {{range .}}
                <div class="note-card">
                    <div class="note-content">{{.Content}}</div>
                    <div class="note-meta">
                        <div class="note-time">
                            🕐 {{.CreatedAt.Format "2006-01-02 15:04"}}
                        </div>
                        <div class="note-id">{{.ID}}</div>
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty-state">
                    <h2>📭</h2>
                    <p>No ideas yet. Start capturing your thoughts!</p>
                </div>
            {{end}}
        </div>
        
        <div class="tip">
            <strong>💬 How to add ideas:</strong><br>
            Send <code>/notes your idea here</code> to your Telegram bot
        </div>
    </div>
    
    <script>
        // Auto refresh every 10 seconds
        setTimeout(() => location.reload(), 10000);
    </script>
</body>
</html>`
