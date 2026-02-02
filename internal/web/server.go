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
	notesList := s.store.GetAll()

	// Group notes by status
	grouped := map[string][]notes.Note{
		"TODO":  {},
		"DOING": {},
		"DONE":  {},
	}

	for _, note := range notesList {
		status := string(note.Status)
		if status == "" {
			status = "TODO" // Default
		}
		grouped[status] = append(grouped[status], note)
	}

	tmpl := template.Must(template.New("home").Parse(homeHTML))
	if err := tmpl.Execute(w, grouped); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		notesList := s.store.GetAll()
		json.NewEncoder(w).Encode(notesList)

	case http.MethodDelete:
		id := strings.TrimPrefix(r.URL.Query().Get("id"), "")
		if s.store.Delete(id) {
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "Note not found", http.StatusNotFound)
		}

	case http.MethodPut: // Update status
		var req struct {
			ID     string           `json:"id"`
			Status notes.NoteStatus `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if s.store.UpdateStatus(req.ID, req.Status) {
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
    <title>💡 Idea Board</title>
    <style>
        :root {
            --bg-color: #1a1a2e;
            --card-bg: rgba(255, 255, 255, 0.08);
            --text-color: #e8e8e8;
            --accent-color: #00d4ff;
            --todo-border: #ff6b6b;
            --doing-border: #feca57;
            --done-border: #1dd1a1;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: var(--text-color);
            padding: 20px;
            overflow-x: hidden;
        }
        
        header {
            text-align: center;
            padding: 20px 0;
            margin-bottom: 30px;
        }
        
        h1 {
            font-size: 2.2em;
            background: linear-gradient(90deg, #00d4ff, #7b2cbf);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        
        .board {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            max-width: 1400px;
            margin: 0 auto;
            align-items: start;
        }
        
        .column {
            background: rgba(0, 0, 0, 0.2);
            border-radius: 12px;
            padding: 15px;
            min-height: 600px;
        }
        
        .column-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding-bottom: 15px;
            margin-bottom: 15px;
            border-bottom: 1px solid rgba(255,255,255,0.1);
            font-weight: bold;
            font-size: 1.1em;
        }
        
        .badge {
            background: rgba(255,255,255,0.1);
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
        }
        
        .todo-header { color: var(--todo-border); }
        .doing-header { color: var(--doing-border); }
        .done-header { color: var(--done-border); }
        
        .note-card {
            background: var(--card-bg);
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 15px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-left: 3px solid transparent;
            cursor: move;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        
        .note-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(0,0,0,0.3);
        }
        
        .note-card[data-status="TODO"] { border-left-color: var(--todo-border); }
        .note-card[data-status="DOING"] { border-left-color: var(--doing-border); }
        .note-card[data-status="DONE"] { border-left-color: var(--done-border); opacity: 0.8; }
        
        .note-content {
            margin-bottom: 10px;
            line-height: 1.5;
            white-space: pre-wrap;
        }
        
        .note-meta {
            display: flex;
            justify-content: space-between;
            font-size: 0.8em;
            color: #888;
        }
        
        .actions {
            display: flex;
            gap: 5px;
            margin-top: 10px;
            justify-content: flex-end;
        }
        
        .btn {
            background: rgba(255,255,255,0.1);
            border: none;
            color: #ccc;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition:  background 0.2s;
        }
        
        .btn:hover { background: rgba(255,255,255,0.2); color: white; }
        .btn-delete:hover { background: rgba(255, 59, 48, 0.3); color: #ff3b30; }
        
        /* Drag and Drop styles */
        .dragging {
            opacity: 0.5;
        }
        
        .column.drag-over {
            background: rgba(255,255,255,0.05);
        }
        
        @media (max-width: 768px) {
            .board { grid-template-columns: 1fr; }
            .column { min-height: auto; }
        }
    </style>
</head>
<body>
    <header>
        <h1>💡 Idea Board</h1>
        <p style="color: #888; margin-top: 5px;">Drag cards to update status</p>
    </header>
    
    <div class="board">
        <!-- TODO Column -->
        <div class="column" id="todo-col" ondrop="drop(event, 'TODO')" ondragover="allowDrop(event)">
            <div class="column-header todo-header">
                📝 TODO <span class="badge">{{len .TODO}}</span>
            </div>
            {{range .TODO}}
                {{template "card" .}}
            {{end}}
        </div>
        
        <!-- DOING Column -->
        <div class="column" id="doing-col" ondrop="drop(event, 'DOING')" ondragover="allowDrop(event)">
            <div class="column-header doing-header">
                🚀 IN PROGRESS <span class="badge">{{len .DOING}}</span>
            </div>
            {{range .DOING}}
                {{template "card" .}}
            {{end}}
        </div>
        
        <!-- DONE Column -->
        <div class="column" id="done-col" ondrop="drop(event, 'DONE')" ondragover="allowDrop(event)">
            <div class="column-header done-header">
                ✅ DONE <span class="badge">{{len .DONE}}</span>
            </div>
            {{range .DONE}}
                {{template "card" .}}
            {{end}}
        </div>
    </div>

    {{define "card"}}
    <div class="note-card" id="{{.ID}}" draggable="true" ondragstart="drag(event)" data-status="{{.Status}}">
        <div class="note-content">{{.Content}}</div>
        <div class="note-meta">
            <span>{{.CreatedAt.Format "01/02 15:04"}}</span>
            <span style="font-family: monospace;">#{{.ID}}</span>
        </div>
        <div class="actions">
            <button class="btn btn-delete" onclick="deleteNote('{{.ID}}')">🗑️</button>
        </div>
    </div>
    {{end}}
    
    <script>
        function allowDrop(ev) {
            ev.preventDefault();
            ev.currentTarget.classList.add('drag-over');
        }
        
        function drag(ev) {
            ev.dataTransfer.setData("text", ev.target.id);
            ev.target.classList.add('dragging');
        }
        
        function drop(ev, status) {
            ev.preventDefault();
            const documentCol = ev.currentTarget;
            documentCol.classList.remove('drag-over');
            
            const data = ev.dataTransfer.getData("text");
            const card = document.getElementById(data);
            
            // Move element visually immediately
            // Only append if dropping on column, not inside another card
            if (ev.target.classList.contains('column') || ev.target.closest('.column')) {
                 const col = ev.target.classList.contains('column') ? ev.target : ev.target.closest('.column');
                 col.appendChild(card);
            }
            
            card.classList.remove('dragging');
            card.setAttribute('data-status', status);
            
            // Update via API
            updateStatus(data, status);
        }
        
        // Remove drag-over style when leaving
        document.querySelectorAll('.column').forEach(col => {
            col.addEventListener('dragleave', () => col.classList.remove('drag-over'));
        });

        async function updateStatus(id, status) {
            try {
                await fetch('/api/notes', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id: id, status: status })
                });
                // Optional: reload to update counts, or update DOM dynamically
                // Location reload for simplicity to update counts
                setTimeout(() => location.reload(), 500); 
            } catch (err) {
                console.error('Failed to update status', err);
                alert('Failed to update status');
            }
        }

        async function deleteNote(id) {
            if (!confirm('Are you sure you want to delete this note?')) return;
            
            try {
                await fetch('/api/notes?id=' + id, { method: 'DELETE' });
                document.getElementById(id).remove();
                location.reload(); // Refresh to update counts
            } catch (err) {
                alert('Failed to delete note');
            }
        }
        
        // Auto refresh every 30 seconds to keep in sync
        setInterval(() => location.reload(), 30000);
    </script>
</body>
</html>`
