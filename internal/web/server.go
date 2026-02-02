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

	// Create map for grouping
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
		StatusOrder  []string
		AllNotesJSON template.JS
	}{
		Columns:      grouped,
		StatusOrder:  []string{"TODO", "DOING", "DONE"},
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
    <title>Jira-like Idea Board</title>
    <style>
        /* Jira Dark Theme Variables */
        :root {
            --bg-color: #1d2125; /* Jira Dark Body */
            --column-bg: #161a1d; /* Jira Dark Column */
            --card-bg: #22272b; /* Jira Dark Card */
            --text-color: #b6c2cf; /* Jira Dark Text */
            --text-primary: #dcdfe4;
            --accent-color: #579dff; /* Jira Blue */
            --border-color: rgba(255, 255, 255, 0.08); /* Subtle borders */
            
            --todo-border: #dfe1e6;
            --doing-border: #0052cc;
            --done-border: #00875a;
            
            --panel-bg: #22272b; /* Side panel matches card bg or slightly different */
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--bg-color);
            background-image: none; /* Removed gradient */
            min-height: 100vh;
            color: var(--text-color);
            padding: 20px;
            padding-right: 20px; /* Base padding */
            transition: padding-right 0.3s ease;
            overflow-x: hidden;
        }

        /* Typography */
        h1 { 
            font-size: 1.5em; 
            color: var(--text-primary); 
            background: none; 
            -webkit-text-fill-color: initial;
            font-weight: 600;
        }
        
        header { text-align: left; padding: 20px 40px; margin-bottom: 20px; border-bottom: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: center; }
        
        .sub-header { color: #8c9bab; font-size: 0.9em; margin-top: 4px; }

        /* Board Layout */
        .board { 
            display: grid; 
            grid-template-columns: repeat(3, 1fr); /* Fixed 3 columns for TODO/DOING/DONE */
            gap: 16px; 
            max-width: 1600px; 
            margin: 0 auto; 
            align-items: start; 
            padding: 0 20px;
        }
        
        .column { 
            background: var(--column-bg); 
            border-radius: 8px; /* Jira rounded corners */
            padding: 12px; 
            min-height: 500px; 
            border: 1px solid var(--border-color);
        }
        
        .column-header { 
            padding-bottom: 12px; 
            margin-bottom: 8px; 
            font-weight: 600; 
            font-size: 0.85em; 
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: #8c9bab;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-bottom: none; /* Jira style headers don't have underline usually, but let's keep it clean */
        }
        
        .badge { 
            background: rgba(255,255,255,0.1); 
            color: var(--text-primary);
            padding: 2px 8px; 
            border-radius: 10px; 
            font-size: 0.9em; 
        }

        /* Cars */
        .note-card {
            background: var(--card-bg); 
            border-radius: 3px; /* Slightly sharper cards like Jira */
            padding: 12px; 
            margin-bottom: 8px;
            border: 1px solid rgba(255,255,255,0.05); /* Very subtle border */
            box-shadow: 0 1px 2px rgba(0,0,0,0.2);
            cursor: pointer; 
            transition: background 0.1s;
            color: var(--text-primary);
        }
        .note-card:hover { 
            background: #2c333a; /* Hover state */
        }
        .note-card.active { 
            box-shadow: 0 0 0 2px var(--accent-color); 
            background: #2c333a;
        }
        
        .note-content { 
            margin-bottom: 12px; 
            line-height: 1.4; 
            font-size: 0.95em;
            white-space: pre-wrap; 
            max-height: 100px; 
            overflow: hidden; 
            text-overflow: ellipsis; 
            display: -webkit-box; 
            -webkit-line-clamp: 3; 
            -webkit-box-orient: vertical; 
            pointer-events: none; 
        }
        
        .note-meta { 
            display: flex; 
            justify-content: space-between; 
            font-size: 0.75em; 
            color: #8c9bab; 
            pointer-events: none; 
            align-items: center;
        }
        
        .note-id {
            background: rgba(255,255,255,0.05);
            padding: 2px 4px;
            border-radius: 3px;
            font-family: monospace;
        }

        /* Status colors for cards */
        /* Jira cards usually don't have left border colors, they use labels. But let's keep the subtle left border for clarity? 
           Or remove it for cleaner look. Let's remove left border and use maybe an icon or just position. 
           User liked "Jira-like". Jira cards are simple. */
        
        /* Drag styles */
        .dragging { opacity: 0.5; transform: rotate(1deg); }
        .column.drag-over { background: rgba(87, 157, 255, 0.1); border: 1px dashed var(--accent-color); }

        /* Side Panel Styles (Jira-like) */
        .side-panel {
            position: fixed;
            top: 0;
            right: 0;
            width: 600px; /* Wider for detail view */
            height: 100vh;
            background: var(--panel-bg);
            border-left: 1px solid var(--border-color);
            box-shadow: -5px 0 30px rgba(0,0,0,0.5);
            transform: translateX(100%);
            transition: transform 0.25s cubic-bezier(0.2, 0, 0, 1);
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
            background: rgba(0,0,0,0.4);
            z-index: 900;
            opacity: 0;
            pointer-events: none;
            transition: opacity 0.2s;
        }
        
        .panel-overlay.show {
            opacity: 1;
            pointer-events: auto;
        }

        .panel-header {
            padding: 24px 32px 16px;
            /* No border bottom in modern Jira detail view usually, just spacing */
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
        }

        .panel-content {
            flex: 1;
            overflow-y: auto;
            padding: 0 32px 32px;
        }

        /* Breadcrumbs / Project ID */
        .note-breadcrumbs {
            font-size: 0.9em;
            color: #8c9bab;
            margin-bottom: 16px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .breadcrumb-link {
            cursor: pointer;
        }
        .breadcrumb-link:hover { text-decoration: underline; color: var(--accent-color); }

        .note-full-content {
            font-size: 1.1em;
            line-height: 1.6;
            margin-bottom: 40px;
            white-space: pre-wrap;
            color: var(--text-primary);
        }

        .section-title {
            font-size: 0.85em;
            font-weight: 700;
            color: var(--text-primary);
            margin-bottom: 16px;
        }

        /* Comments */
        .comment-list {
            display: flex;
            flex-direction: column;
            gap: 20px;
            margin-top: 24px;
        }
        
        .comment {
            display: flex;
            gap: 12px;
        }
        
        .comment-avatar {
            width: 32px;
            height: 32px;
            border-radius: 50%;
            background: #44546f;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            font-weight: bold;
            color: #1d2125;
            flex-shrink: 0;
        }

        .comment-body {
            flex: 1;
        }

        .comment-header {
            display: flex;
            align-items: baseline;
            gap: 8px;
            margin-bottom: 4px;
        }
        
        .comment-author { font-weight: 600; color: var(--text-primary); font-size: 0.95em; }
        .comment-date { color: #8c9bab; font-size: 0.85em; }
        .comment-text { color: var(--text-color); line-height: 1.5; font-size: 0.95em; white-space: pre-wrap; }

        .comment-input-wrapper {
            margin-top: 8px;
            border: 1px solid var(--border-color);
            background: var(--bg-color); /* Slightly lighter/darker input bg */
            border-radius: 3px;
            transition: box-shadow 0.2s;
        }
        .comment-input-wrapper:focus-within {
            box-shadow: 0 0 0 1px var(--accent-color);
            border-color: var(--accent-color);
        }
        
        .comment-input {
            width: 100%;
            background: transparent;
            border: none;
            color: var(--text-primary);
            min-height: 40px;
            padding: 12px;
            resize: vertical;
            font-family: inherit;
            display: block;
        }
        .comment-input:focus { outline: none; }
        
        .comment-actions {
            padding: 8px 12px;
            display: flex;
            justify-content: flex-end;
            gap: 8px;
            background: rgba(0,0,0,0.2); /* Footer of input */
        }

        /* Status Select (Jira Style Dropdown) */
        .status-badge-select {
            appearance: none;
            background-color: var(--bg-color);
            border: 1px solid transparent;
            color: var(--text-primary);
            padding: 6px 12px;
            border-radius: 3px;
            font-weight: 600;
            font-size: 0.85em;
            cursor: pointer;
            text-transform: uppercase;
            transition: background 0.1s;
        }
        .status-badge-select:hover { background-color: rgba(255,255,255,0.05); }
        .status-badge-select option { background-color: var(--panel-bg); color: var(--text-color); }
        
        /* Specific status colors for the dropdown button itself via JS/CSS mapping? 
           For simplicity, just bold text. */
        
        .btn-icon { background: none; border: none; font-size: 1.5em; color: #8c9bab; cursor: pointer; padding: 4px; border-radius: 3px; }
        .btn-icon:hover { background: rgba(255,255,255,0.1); color: var(--text-primary); }
        
        .btn-primary { 
            background: var(--accent-color); 
            color: #1d2125; /* Dark text on bright blue */
            border: none; 
            padding: 6px 12px; 
            border-radius: 3px; 
            font-weight: 600; 
            cursor: pointer; 
            font-size: 0.9em;
        }
        .btn-primary:hover { filter: brightness(1.1); }

        @media (max-width: 1024px) { 
            .board { grid-template-columns: 1fr; } 
            .column { min-height: auto; } 
            .side-panel { width: 100%; top: 60px; height: calc(100vh - 60px); border-radius: 12px 12px 0 0; }
        }
    </style>
</head>
<body>
    <header>
        <div>
            <h1>Kanban Board</h1>
            <div class="sub-header">All project updates</div>
        </div>
        <div style="font-size:0.9em; color:#8c9bab;">
            Drag to update • Click to view
        </div>
    </header>
    
    <div class="board">
        {{$notesMap := .Columns}}
        {{range $status := .StatusOrder}}
        {{$notes := index $notesMap $status}}
        <div class="column" id="{{$status}}-col" 
             ondrop="drop(event, '{{$status}}')" 
             ondragover="allowDrop(event)">
            <div class="column-header {{$status}}-header">
                <span>{{$status}}</span>
                <span class="badge">{{len $notes}}</span>
            </div>
            {{range $notes}}
            <div class="note-card" id="{{.ID}}" 
                 draggable="true" 
                 ondragstart="drag(event)" 
                 onclick="openPanel('{{.ID}}')"
                 data-status="{{.Status}}">
                <div class="note-content">{{.Content}}</div>
                <div class="note-meta">
                    <span class="note-id">IDEA-{{printf "%.4s" .ID}}</span>
                    <span>{{.CreatedAt.Format "Jan 02"}}</span>
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
                <span class="breadcrumb-link">Projects</span>
                <span>/</span>
                <span class="breadcrumb-link">Idea Board</span>
                <span>/</span>
                <span id="panelId" style="color: var(--text-primary);"></span>
            </div>
            <div style="display: flex; gap: 12px; align-items: center;">
                <select id="panelStatus" class="status-badge-select" onchange="updateNoteStatusFromPanel()">
                    <option value="TODO">TO DO</option>
                    <option value="DOING">IN PROGRESS</option>
                    <option value="DONE">DONE</option>
                </select>
                <div style="width:1px; height:20px; background:var(--border-color);"></div>
                <button class="btn-icon" onclick="closePanel()">✕</button>
            </div>
        </div>
        
        <div class="panel-content">
            <div id="panelContent" class="note-full-content"></div>
            
            <div class="section-title">Description</div>
            <div style="color:#8c9bab; font-size:0.9em; margin-bottom: 30px;">
                (No detailed description provided. Use the content above.)
            </div>

            <div class="section-title">
                Activity
            </div>
            
            <div style="display:flex; gap: 12px;">
                <div class="comment-avatar">U</div>
                <div style="flex:1;">
                    <div class="comment-input-wrapper">
                        <textarea id="newComment" class="comment-input" placeholder="Add a comment..." onkeydown="handleCommentKeydown(event)"></textarea>
                        <div class="comment-actions">
                            <span style="font-size:0.8em; color:#8c9bab; margin-right:auto; padding-top:6px;">Pro tip: press M to comment</span>
                            <button class="btn-primary" onclick="addComment()">Save</button>
                        </div>
                    </div>
                </div>
            </div>

            <div id="panelComments" class="comment-list">
                <!-- Comments injected -->
            </div>
        </div>
    </div>

    <script>
        // Inject data
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
            if (notesData[id]) notesData[id].status = status;
            
            if (currentNoteId === id) {
                 document.getElementById('panelStatus').value = status;
            }
        }

        // Panel Logic
        function openPanel(id) {
            if (event && event.type !== 'click') return;
            
            const note = notesData[id];
            if (!note) return;
            
            currentNoteId = id;
            
            document.querySelectorAll('.note-card').forEach(c => c.classList.remove('active'));
            document.getElementById(id).classList.add('active');

            // ID Formatting
            document.getElementById('panelId').textContent = 'IDEA-' + note.id.substring(0,4); 
            document.getElementById('panelContent').textContent = note.content;
            document.getElementById('panelStatus').value = note.status;
            
            renderComments(note.comments || []);
            
            document.getElementById('sidePanel').classList.add('open');
            document.getElementById('panelOverlay').classList.add('show');
        }

        function closePanel() {
            document.getElementById('sidePanel').classList.remove('open');
            document.getElementById('panelOverlay').classList.remove('show');
            document.querySelectorAll('.note-card').forEach(c => c.classList.remove('active'));
            currentNoteId = null;
        }

        function renderComments(comments) {
            const container = document.getElementById('panelComments');
            container.innerHTML = '';
            
            const sorted = comments.sort((a,b) => new Date(b.created_at) - new Date(a.created_at));

            sorted.forEach(c => {
                const div = document.createElement('div');
                div.className = 'comment';
                const date = new Date(c.created_at).toLocaleString(undefined, { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' });
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
                
                if (!notesData[currentNoteId].comments) notesData[currentNoteId].comments = [];
                notesData[currentNoteId].comments.push(comment);
                
                renderComments(notesData[currentNoteId].comments);
                input.value = '';
            } catch (err) {
                alert('In demo mode / Failed to add comment');
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
            
            const card = document.getElementById(currentNoteId);
            const col = document.getElementById(status + '-col');
            col.appendChild(card);
            card.setAttribute('data-status', status);
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') closePanel();
            if (e.key === 'm' && !['INPUT','TEXTAREA'].includes(document.activeElement.tagName) && currentNoteId) {
                e.preventDefault();
                document.getElementById('newComment').focus();
            }
        });
    </script>
</body>
</html>`
