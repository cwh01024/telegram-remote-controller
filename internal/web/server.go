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
	http.HandleFunc("/api/notes/comments", s.handleCommentsAPI)

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

	notesJSON, err := json.Marshal(notesList)
	if err != nil {
		log.Printf("Error marshaling notes: %v", err)
		notesJSON = []byte("[]")
	}

	data := struct {
		Columns      map[string][]notes.Note
		AllNotesJSON template.JS
	}{
		Columns:      grouped,
		AllNotesJSON: template.JS(notesJSON),
	}

	tmpl := template.Must(template.New("home").Parse(homeHTML))
	if err := tmpl.Execute(w, data); err != nil {
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

func (s *Server) handleCommentsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NoteID  string `json:"note_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if comment, ok := s.store.AddComment(req.NoteID, req.Content); ok {
		json.NewEncoder(w).Encode(comment)
	} else {
		http.Error(w, "Note not found", http.StatusNotFound)
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
            --panel-bg: #16213e;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: var(--text-color);
            padding: 20px;
            padding-right: 20px; /* Base padding */
            transition: padding-right 0.3s ease;
            overflow-x: hidden;
        }

        /* Side Panel Styles (Jira-like) */
        .side-panel {
            position: fixed;
            top: 0;
            right: 0;
            width: 450px;
            height: 100vh;
            background: var(--panel-bg);
            border-left: 1px solid rgba(255, 255, 255, 0.1);
            box-shadow: -5px 0 30px rgba(0,0,0,0.5);
            transform: translateX(100%);
            transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
            z-index: 1000;
            display: flex;
            flex-direction: column;
        }
        
        .side-panel.open {
            transform: translateX(0);
        }
        
        /* Overlay for mobile or focus */
        .panel-overlay {
            position: fixed;
            top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.3);
            z-index: 900;
            opacity: 0;
            pointer-events: none;
            transition: opacity 0.3s;
        }
        
        .panel-overlay.show {
            opacity: 1;
            pointer-events: auto;
        }

        .panel-header {
            padding: 20px 24px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: rgba(0,0,0,0.1);
        }

        .panel-content {
            flex: 1;
            overflow-y: auto;
            padding: 24px;
        }

        .panel-actions {
            padding: 16px 24px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            background: rgba(0,0,0,0.1);
            display: flex;
            justify-content: flex-end;
        }

        .note-breadcrumbs {
            font-size: 0.85em;
            color: #888;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            gap: 5px;
        }

        .note-full-content {
            font-size: 1.1em;
            line-height: 1.6;
            margin-bottom: 30px;
            white-space: pre-wrap;
            color: #fff;
        }

        .section-title {
            font-size: 0.8em;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #888;
            margin-bottom: 12px;
            font-weight: 600;
        }

        /* Comments */
        .comment-list {
            display: flex;
            flex-direction: column;
            gap: 16px;
            margin-bottom: 24px;
        }
        
        .comment {
            display: flex;
            gap: 12px;
        }
        
        .comment-avatar {
            width: 32px;
            height: 32px;
            border-radius: 50%;
            background: linear-gradient(135deg, #00d4ff, #7b2cbf);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            font-weight: bold;
            color: white;
            flex-shrink: 0;
        }

        .comment-body {
            flex: 1;
        }

        .comment-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 4px;
            font-size: 0.9em;
        }
        
        .comment-author { font-weight: 500; color: #ccc; }
        .comment-date { color: #666; font-size: 0.85em; }
        .comment-text { color: #e8e8e8; line-height: 1.5; font-size: 0.95em; white-space: pre-wrap; }

        .comment-input-wrapper {
            background: rgba(255,255,255,0.05);
            border-radius: 8px;
            padding: 12px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        
        .comment-input {
            width: 100%;
            background: transparent;
            border: none;
            color: white;
            min-height: 60px;
            resize: vertical;
            font-family: inherit;
            margin-bottom: 10px;
        }
        .comment-input:focus { outline: none; }

        /* Board Styles */
        header { text-align: center; padding: 20px 0; margin-bottom: 30px; }
        h1 { font-size: 2.2em; background: linear-gradient(90deg, #00d4ff, #7b2cbf); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        
        .board { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; max-width: 1400px; margin: 0 auto; align-items: start; transition: padding-right 0.3s; }
        .column { background: rgba(0, 0, 0, 0.2); border-radius: 12px; padding: 15px; min-height: 600px; }
        .column-header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 15px; margin-bottom: 15px; border-bottom: 1px solid rgba(255,255,255,0.1); font-weight: bold; font-size: 1.1em; }
        .badge { background: rgba(255,255,255,0.1); padding: 2px 8px; border-radius: 12px; font-size: 0.8em; }
        
        .todo-header { color: var(--todo-border); }
        .doing-header { color: var(--doing-border); }
        .done-header { color: var(--done-border); }
        
        .note-card {
            background: var(--card-bg); border-radius: 8px; padding: 15px; margin-bottom: 15px;
            backdrop-filter: blur(10px); border: 1px solid rgba(255, 255, 255, 0.05);
            border-left: 3px solid transparent; cursor: pointer; transition: all 0.2s;
        }
        .note-card:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,0,0,0.3); background: rgba(255,255,255,0.12); }
        .note-card.active { border-color: var(--accent-color); background: rgba(255,255,255,0.15); box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.2); }
        
        .note-card[data-status="TODO"] { border-left-color: var(--todo-border); }
        .note-card[data-status="DOING"] { border-left-color: var(--doing-border); }
        .note-card[data-status="DONE"] { border-left-color: var(--done-border); opacity: 0.8; }
        
        .note-content { margin-bottom: 10px; line-height: 1.5; white-space: pre-wrap; max-height: 100px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; pointer-events: none; }
        .note-meta { display: flex; justify-content: space-between; font-size: 0.8em; color: #888; pointer-events: none; }
        
        /* Status Select */
        .status-badge-select {
            background: transparent;
            border: 1px solid rgba(255,255,255,0.2);
            color: #ddd;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            cursor: pointer;
        }
        .status-badge-select:hover { background: rgba(255,255,255,0.1); }
        
        .btn-icon { background: none; border: none; font-size: 1.2em; color: #888; cursor: pointer; padding: 5px; border-radius: 50%; transition: all 0.2s; }
        .btn-icon:hover { background: rgba(255,255,255,0.1); color: white; }
        .btn-primary { background: var(--accent-color); color: #0f3460; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 600; cursor: pointer; }
        .btn-primary:hover { opacity: 0.9; }

        /* Drag styles */
        .dragging { opacity: 0.5; }
        .column.drag-over { background: rgba(255,255,255,0.05); }
        
        @media (max-width: 768px) { 
            .board { grid-template-columns: 1fr; } 
            .column { min-height: auto; } 
            .side-panel { width: 100%; top: 50px; height: calc(100vh - 50px); border-top-left-radius: 20px; border-top-right-radius: 20px; }
        }
    </style>
</head>
<body>
    <header>
        <h1>💡 Idea Board</h1>
        <p style="color: #888; margin-top: 5px;">Drag to move • Click to view details</p>
    </header>
    
    <div class="board">
        {{range $status, $notes := .Columns}}
        <div class="column" id="{{$status}}-col" 
             ondrop="drop(event, '{{$status}}')" 
             ondragover="allowDrop(event)">
            <div class="column-header {{$status}}-header">
                {{$status}} <span class="badge">{{len $notes}}</span>
            </div>
            {{range $notes}}
            <div class="note-card" id="{{.ID}}" 
                 draggable="true" 
                 ondragstart="drag(event)" 
                 onclick="openPanel('{{.ID}}')"
                 data-status="{{.Status}}">
                <div class="note-content">{{.Content}}</div>
                <div class="note-meta">
                    <span>{{.CreatedAt.Format "01/02 15:04"}}</span>
                    <span style="font-family: monospace;">#{{.ID}}</span>
                </div>
            </div>
            {{end}}
        </div>
        {{end}}
    </div>

    <!-- Overlay -->
    <div class="panel-overlay" id="panelOverlay" onclick="closePanel()"></div>

    <!-- Side Panel -->
    <div class="side-panel" id="sidePanel">
        <div class="panel-header">
            <div class="note-breadcrumbs">
                <span id="panelId"></span>
                <span>/</span>
                <span id="panelDate"></span>
            </div>
            <div style="display: flex; gap: 10px; align-items: center;">
                <select id="panelStatus" class="status-badge-select" onchange="updateNoteStatusFromPanel()">
                    <option value="TODO">TODO</option>
                    <option value="DOING">DOING</option>
                    <option value="DONE">DONE</option>
                </select>
                <button class="btn-icon" onclick="closePanel()">✕</button>
            </div>
        </div>
        
        <div class="panel-content">
            <div class="section-title">Description</div>
            <div id="panelContent" class="note-full-content"></div>
            
            <div class="section-title">
                Activity 
                <span class="badge" id="commentCount">0</span>
            </div>
            
            <div class="comment-input-wrapper">
                <textarea id="newComment" class="comment-input" placeholder="Add a comment..." onkeydown="handleCommentKeydown(event)"></textarea>
                <div style="text-align: right;">
                    <button class="btn-primary" onclick="addComment()">Comment</button>
                </div>
            </div>

            <div id="panelComments" class="comment-list">
                <!-- Comments injected here -->
            </div>
        </div>
    </div>

    <script>
        // Store notes data for access
        const allNotesList = {{.AllNotesJSON}};
        const notesData = {};
        if (allNotesList) {
            allNotesList.forEach(n => notesData[n.id] = n);
        }

        let currentNoteId = null;

        // Drag & Drop
        function allowDrop(ev) { ev.preventDefault(); ev.currentTarget.classList.add('drag-over'); }
        function drag(ev) { ev.dataTransfer.setData("text", ev.target.id); ev.target.classList.add('dragging'); }
        function drop(ev, status) {
            ev.preventDefault();
            ev.currentTarget.classList.remove('drag-over');
            const id = ev.dataTransfer.getData("text");
            const card = document.getElementById(id);
            if (card) {
                let target = ev.target;
                while (!target.classList.contains('column')) { target = target.parentElement; }
                target.appendChild(card);
                card.classList.remove('dragging');
                card.setAttribute('data-status', status);
                updateStatus(id, status);
            }
        }
        document.querySelectorAll('.column').forEach(col => col.addEventListener('dragleave', () => col.classList.remove('drag-over')));

        async function updateStatus(id, status) {
            await fetch('/api/notes', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id, status: status })
            });
            if (notesData[id]) notesData[id].Status = status;
            
            // If panel is open for this note, update status select
            if (currentNoteId === id) {
                 document.getElementById('panelStatus').value = status;
            }
        }

        // Panel Logic
        function openPanel(id) {
            // Prevent event bubbling if triggered from drag
            if (event && event.type !== 'click') return;
            
            const note = notesData[id];
            if (!note) {
                console.error("Note not found:", id);
                return;
            }
            
            currentNoteId = id;
            
            // Update Active Card UI
            document.querySelectorAll('.note-card').forEach(c => c.classList.remove('active'));
            document.getElementById(id).classList.add('active');

            // Populate Panel
            document.getElementById('panelId').textContent = 'TEST-' + note.ID.split('-')[1]; // Simulate Jira ID
            document.getElementById('panelDate').textContent = new Date(note.CreatedAt).toLocaleString();
            document.getElementById('panelContent').textContent = note.Content;
            document.getElementById('panelStatus').value = note.Status;
            
            renderComments(note.Comments || []);
            
            // Show Panel
            document.getElementById('sidePanel').classList.add('open');
            document.getElementById('panelOverlay').classList.add('show');
            
            // Adjust body size? Optional: for split view feel
            // document.body.style.paddingRight = '450px';
        }

        function closePanel() {
            document.getElementById('sidePanel').classList.remove('open');
            document.getElementById('panelOverlay').classList.remove('show');
            document.querySelectorAll('.note-card').forEach(c => c.classList.remove('active'));
            currentNoteId = null;
            // document.body.style.paddingRight = '20px';
        }

        function renderComments(comments) {
            const container = document.getElementById('panelComments');
            document.getElementById('commentCount').textContent = comments.length;
            container.innerHTML = '';
            
            // Show newest at bottom? or top? Jira usually newest at bottom of list but input is at top? 
            // Let's do newest at bottom.
            comments.forEach(c => {
                const div = document.createElement('div');
                div.className = 'comment';
                const date = new Date(c.created_at).toLocaleString();
                div.innerHTML = 
                    '<div class="comment-avatar">U</div>' +
                    '<div class="comment-body">' +
                        '<div class="comment-header">' +
                            '<span class="comment-author">User</span>' +
                            '<span class="comment-date">' + date + '</span>' +
                        '</div>' +
                        '<div class="comment-text">' + escapeHtml(c.content) + '</div>' +
                    '</div>';
                container.appendChild(div);
            });
        }

        async function addComment() {
            const input = document.getElementById('newComment');
            const content = input.value.trim();
            if (!content || !currentNoteId) return;
            
            try {
                const res = await fetch('/api/notes/comments', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ note_id: currentNoteId, content: content })
                });
                const comment = await res.json();
                
                if (!notesData[currentNoteId].Comments) notesData[currentNoteId].Comments = [];
                notesData[currentNoteId].Comments.push(comment);
                
                renderComments(notesData[currentNoteId].Comments);
                input.value = '';
            } catch (err) {
                alert('Failed to add comment');
            }
        }
        
        function handleCommentKeydown(e) {
            if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                addComment();
            }
        }

        async function updateNoteStatusFromPanel() {
            const status = document.getElementById('panelStatus').value;
            if (!currentNoteId) return;
            
            await updateStatus(currentNoteId, status);
            
            // Move card in DOM to correct column
            const card = document.getElementById(currentNoteId);
            const col = document.getElementById(status + '-col');
            col.appendChild(card);
            card.setAttribute('data-status', status);
            
            // Don't reload, smooth!
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        // Escape key to close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closePanel();
        });
    </script>
</body>
</html>`
