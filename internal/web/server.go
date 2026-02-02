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

	case http.MethodPut: // Update status or content
		var req struct {
			ID      string           `json:"id"`
			Status  notes.NoteStatus `json:"status"`
			Content string           `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		success := false
		if req.Status != "" {
			if s.store.UpdateStatus(req.ID, req.Status) {
				success = true
			}
		}
		if req.Content != "" {
			if s.store.UpdateContent(req.ID, req.Content) {
				success = true
			}
		}

		if success {
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "Note not found or nothing to update", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCommentsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
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

	case http.MethodPut: // Edit comment
		var req struct {
			NoteID    string `json:"note_id"`
			CommentID string `json:"comment_id"`
			Content   string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if s.store.UpdateComment(req.NoteID, req.CommentID, req.Content) {
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		} else {
			http.Error(w, "Comment not found", http.StatusNotFound)
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
    <title>Jira-like Idea Board</title>
    <style>
        /* Jira Dark Theme Variables */
        :root {
            --bg-color: #1d2125;
            --column-bg: #161a1d;
            --card-bg: #22272b;
            --text-color: #b6c2cf;
            --text-primary: #dcdfe4;
            --accent-color: #579dff;
            --border-color: rgba(255, 255, 255, 0.08);
            --panel-bg: #22272b;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--bg-color);
            min-height: 100vh;
            color: var(--text-color);
            padding: 20px;
            padding-right: 20px;
            overflow-x: hidden;
        }

        h1 { font-size: 1.5em; color: var(--text-primary); font-weight: 600; }
        header { padding: 20px 40px; margin-bottom: 20px; border-bottom: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: center; }
        .sub-header { color: #8c9bab; font-size: 0.9em; margin-top: 4px; }

        .board { 
            display: grid; 
            grid-template-columns: repeat(3, 1fr); 
            gap: 16px; 
            max-width: 1600px; 
            margin: 0 auto; 
            align-items: start; 
            padding: 0 20px;
        }
        
        .column { 
            background: var(--column-bg); 
            border-radius: 8px; 
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
            color: #8c9bab; 
            display: flex; 
            justify-content: space-between; 
            align-items: center; 
        }
        
        .badge { 
            background: rgba(255,255,255,0.1); 
            color: var(--text-primary); 
            padding: 2px 8px; 
            border-radius: 10px; 
            font-size: 0.9em; 
        }

        .note-card {
            background: var(--card-bg); 
            border-radius: 3px; 
            padding: 12px; 
            margin-bottom: 8px;
            border: 1px solid rgba(255,255,255,0.05); 
            box-shadow: 0 1px 2px rgba(0,0,0,0.2);
            cursor: pointer; 
            transition: background 0.1s;
            color: var(--text-primary);
        }
        .note-card:hover { background: #2c333a; }
        .note-card.active { box-shadow: 0 0 0 2px var(--accent-color); background: #2c333a; }
        
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

        .dragging { opacity: 0.5; transform: rotate(1deg); }
        .column.drag-over { background: rgba(87, 157, 255, 0.1); border: 1px dashed var(--accent-color); }

        /* Side Panel */
        .side-panel {
            position: fixed;
            top: 0;
            right: 0;
            width: 600px; 
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
        
        .side-panel.open { transform: translateX(0); }
        .panel-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.4); z-index: 900; opacity: 0; pointer-events: none; transition: opacity 0.2s; }
        .panel-overlay.show { opacity: 1; pointer-events: auto; }

        .panel-header {
            padding: 24px 32px 16px;
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
        }

        .panel-content {
            flex: 1;
            overflow-y: auto;
            padding: 0 32px 32px;
        }

        .note-breadcrumbs { font-size: 0.9em; color: #8c9bab; margin-bottom: 16px; display: flex; align-items: center; gap: 8px; }
        .breadcrumb-link { cursor: pointer; }
        .breadcrumb-link:hover { text-decoration: underline; color: var(--accent-color); }

        .note-full-content {
            font-size: 1.1em;
            line-height: 1.6;
            margin-bottom: 40px;
            white-space: pre-wrap;
            color: var(--text-primary);
            cursor: pointer;
            padding: 8px;
            border-radius: 4px;
            border: 1px solid transparent;
        }
        .note-full-content:hover { background: rgba(255,255,255,0.05); }

        .editable-textarea {
            width: 100%;
            min-height: 150px;
            background: var(--bg-color);
            border: 2px solid var(--accent-color);
            color: var(--text-primary);
            padding: 12px;
            border-radius: 4px;
            font-size: 1.1em;
            line-height: 1.6;
            margin-bottom: 10px;
            resize: vertical;
        }

        .edit-actions { display: flex; gap: 8px; margin-bottom: 20px; }

        .section-title { font-size: 0.85em; font-weight: 700; color: var(--text-primary); margin-bottom: 16px; }

        .comment-list { display: flex; flex-direction: column; gap: 20px; margin-top: 24px; }
        .comment { display: flex; gap: 12px; }
        .comment-avatar { width: 32px; height: 32px; border-radius: 50%; background: #44546f; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: bold; color: #1d2125; flex-shrink: 0; }
        .comment-body { flex: 1; }
        .comment-header { display: flex; align-items: baseline; gap: 8px; margin-bottom: 4px; }
        .comment-author { font-weight: 600; color: var(--text-primary); font-size: 0.95em; }
        .comment-date { color: #8c9bab; font-size: 0.85em; }
        
        .comment-text { 
            color: var(--text-color); 
            line-height: 1.5; 
            font-size: 0.95em; 
            white-space: pre-wrap;
            cursor: pointer;
            padding: 4px;
            border-radius: 4px;
            border: 1px solid transparent;
        }
        .comment-text:hover { background: rgba(255,255,255,0.05); }

        .comment-input-wrapper {
            margin-top: 8px;
            border: 1px solid var(--border-color);
            background: var(--bg-color);
            border-radius: 3px;
            transition: box-shadow 0.2s;
        }
        .comment-input-wrapper:focus-within { box-shadow: 0 0 0 1px var(--accent-color); border-color: var(--accent-color); }
        .comment-input { width: 100%; background: transparent; border: none; color: var(--text-primary); min-height: 40px; padding: 12px; resize: vertical; font-family: inherit; display: block; }
        .comment-input:focus { outline: none; }
        .comment-actions { padding: 8px 12px; display: flex; justify-content: flex-end; gap: 8px; background: rgba(0,0,0,0.2); }

        .status-badge-select { appearance: none; background-color: var(--bg-color); border: 1px solid transparent; color: var(--text-primary); padding: 6px 12px; border-radius: 3px; font-weight: 600; font-size: 0.85em; cursor: pointer; text-transform: uppercase; transition: background 0.1s; }
        .status-badge-select:hover { background-color: rgba(255,255,255,0.05); }
        .status-badge-select option { background-color: var(--panel-bg); color: var(--text-color); }
        
        .btn-icon { background: none; border: none; font-size: 1.5em; color: #8c9bab; cursor: pointer; padding: 4px; border-radius: 3px; }
        .btn-icon:hover { background: rgba(255,255,255,0.1); color: var(--text-primary); }
        
        .btn-primary { background: var(--accent-color); color: #1d2125; border: none; padding: 6px 12px; border-radius: 3px; font-weight: 600; cursor: pointer; font-size: 0.9em; }
        .btn-primary:hover { filter: brightness(1.1); }
        .btn-secondary { background: rgba(255,255,255,0.1); color: var(--text-primary); border: none; padding: 6px 12px; border-radius: 3px; font-weight: 600; cursor: pointer; font-size: 0.9em; }
        .btn-secondary:hover { background: rgba(255,255,255,0.2); }

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
            Drag to update • Click to view • Click text to edit
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

    <div class="panel-overlay" id="panelOverlay" onclick="closePanel()"></div>

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
            <div id="descriptionContainer">
                <div id="panelContent" class="note-full-content" onclick="editDescription()" title="Click to edit"></div>
            </div>
            
            <div class="section-title">Activity</div>
            
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

            <div id="panelComments" class="comment-list"></div>
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

        // --- Drag & Drop ---
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

        // --- API Calls ---
        async function updateStatus(id, status) {
            await fetch('/api/notes', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id, status: status })
            });
            if (notesData[id]) notesData[id].status = status;
            if (currentNoteId === id) document.getElementById('panelStatus').value = status;
        }

        async function updateContent(id, content) {
            await fetch('/api/notes', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: id, content: content })
            });
            if (notesData[id]) notesData[id].content = content;
            
            // Update UI
            document.getElementById('panelContent').textContent = content;
            // Also update card in board
            const cardContent = document.querySelector('#' + id + ' .note-content');
            if (cardContent) cardContent.textContent = content;
        }

        async function updateComment(noteId, commentId, content) {
            await fetch('/api/notes/comments', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ note_id: noteId, comment_id: commentId, content: content })
            });
            
            // Update local data
            const note = notesData[noteId];
            const comment = note.comments.find(c => c.id === commentId);
            if (comment) comment.content = content;
            
            renderComments(note.comments);
        }

        // --- Panel Logic ---
        function openPanel(id) {
            if (event && event.type !== 'click') return;
            const note = notesData[id];
            if (!note) return;
            
            currentNoteId = id;
            document.querySelectorAll('.note-card').forEach(c => c.classList.remove('active'));
            document.getElementById(id).classList.add('active');

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
            // Revert any open edits
            const descContainer = document.getElementById('descriptionContainer');
            if (descContainer.querySelector('textarea')) {
                // If editing, reverting to view is handled by next openPanel, but safer to clean:
                descContainer.innerHTML = '<div id="panelContent" class="note-full-content" onclick="editDescription()" title="Click to edit"></div>';
            }
        }

        // --- Comment Logic ---
        function renderComments(comments) {
            const container = document.getElementById('panelComments');
            container.innerHTML = '';
            const sorted = (comments || []).sort((a,b) => new Date(b.created_at) - new Date(a.created_at));

            sorted.forEach(c => {
                const div = document.createElement('div');
                div.className = 'comment';
                div.id = 'comment-' + c.id;
                const date = new Date(c.created_at).toLocaleString(undefined, { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' });
                
                div.innerHTML = 
                    '<div class="comment-avatar">U</div>' +
                    '<div class="comment-body">' +
                        '<div class="comment-header">' +
                            '<span class="comment-author">User</span>' +
                            '<span class="comment-date">' + date + '</span>' +
                        '</div>' +
                        '<div class="comment-text" onclick="editComment(\'' + c.id + '\')" title="Click to edit">' + escapeHtml(c.content) + '</div>' +
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
                alert('Failed to add comment');
            }
        }

        // --- Editing Logic ---
        function editDescription() {
            const container = document.getElementById('descriptionContainer');
            const contentDiv = document.getElementById('panelContent');
            const currentText = contentDiv.textContent;
            
            container.innerHTML = 
                '<textarea id="editDescriptionInput" class="editable-textarea">' + escapeHtml(currentText) + '</textarea>' +
                '<div class="edit-actions">' +
                    '<button class="btn-primary" onclick="saveDescription()">Save</button>' +
                    '<button class="btn-secondary" onclick="cancelDescription(\'' + escapeJs(currentText) + '\')">Cancel</button>' +
                '</div>';
            
            document.getElementById('editDescriptionInput').focus();
        }

        function saveDescription() {
            const newText = document.getElementById('editDescriptionInput').value;
            updateContent(currentNoteId, newText);
            // Restore UI
            const container = document.getElementById('descriptionContainer');
            container.innerHTML = '<div id="panelContent" class="note-full-content" onclick="editDescription()" title="Click to edit">' + escapeHtml(newText) + '</div>';
        }

        function cancelDescription(originalText) {
            const container = document.getElementById('descriptionContainer');
            container.innerHTML = '<div id="panelContent" class="note-full-content" onclick="editDescription()" title="Click to edit">' + escapeHtml(originalText) + '</div>';
        }

        function editComment(commentId) {
            const commentDiv = document.getElementById('comment-' + commentId);
            const textDiv = commentDiv.querySelector('.comment-text');
            const currentText = textDiv.textContent;
            
            // Replace text div with input
            textDiv.parentNode.innerHTML += 
                '<div class="comment-edit-area" style="margin-top:8px;">' +
                    '<textarea class="editable-textarea" style="min-height:80px;">' + escapeHtml(currentText) + '</textarea>' +
                    '<div class="edit-actions">' +
                        '<button class="btn-primary" onclick="saveComment(\'' + commentId + '\', this)">Save</button>' +
                        '<button class="btn-secondary" onclick="renderComments(notesData[\'' + currentNoteId + '\'].comments)">Cancel</button>' +
                    '</div>' +
                '</div>';
            
            textDiv.style.display = 'none';
            commentDiv.querySelector('textarea').focus();
        }

        function saveComment(commentId, btn) {
            const textarea = btn.parentNode.parentNode.querySelector('textarea');
            const newContent = textarea.value;
            updateComment(currentNoteId, commentId, newContent);
        }

        // --- Helpers ---
        function handleCommentKeydown(e) { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) addComment(); }
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
        function escapeJs(text) {
            return text.replace(/'/g, "\\'").replace(/"/g, '\\"').replace(/\n/g, '\\n').replace(/\r/g, '');
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
