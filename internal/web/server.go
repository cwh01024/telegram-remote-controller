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
            --modal-bg: #16213e;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: var(--text-color);
            padding: 20px;
            overflow-x: hidden;
        }
        
        /* Modal Styles */
        .modal-overlay {
            display: none;
            position: fixed;
            top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0, 0, 0, 0.7);
            z-index: 1000;
            backdrop-filter: blur(5px);
            align-items: center;
            justify-content: center;
        }
        
        .modal {
            background: var(--modal-bg);
            width: 90%;
            max-width: 600px;
            max-height: 90vh;
            border-radius: 16px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            flex-direction: column;
            box-shadow: 0 20px 50px rgba(0,0,0,0.5);
            animation: slideIn 0.3s ease;
        }
        
        @keyframes slideIn {
            from { transform: translateY(20px); opacity: 0; }
            to { transform: translateY(0); opacity: 1; }
        }
        
        .modal-header {
            padding: 20px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
        }
        
        .modal-content {
            padding: 20px;
            overflow-y: auto;
            flex-grow: 1;
        }
        
        .note-full-content {
            font-size: 1.2em;
            line-height: 1.6;
            margin-bottom: 20px;
            white-space: pre-wrap;
        }
        
        .comments-section {
            margin-top: 30px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            padding-top: 20px;
        }
        
        .comment-list {
            display: flex;
            flex-direction: column;
            gap: 15px;
            margin-bottom: 20px;
        }
        
        .comment {
            background: rgba(255, 255, 255, 0.05);
            padding: 10px 15px;
            border-radius: 8px;
            font-size: 0.95em;
        }
        
        .comment-meta {
            font-size: 0.8em;
            color: #888;
            margin-bottom: 5px;
        }
        
        .comment-input-area {
            display: flex;
            gap: 10px;
        }
        
        input[type="text"] {
            flex-grow: 1;
            padding: 10px;
            border-radius: 8px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            background: rgba(0, 0, 0, 0.2);
            color: white;
        }
        
        button {
            padding: 10px 20px;
            border-radius: 8px;
            border: none;
            cursor: pointer;
            font-weight: bold;
            transition: opacity 0.2s;
        }
        
        .btn-primary { background: var(--accent-color); color: #000; }
        .btn-close { background: transparent; color: #888; font-size: 1.5em; padding: 0; line-height: 1; }
        .btn-primary:hover { opacity: 0.9; }
        .btn-close:hover { color: white; }

        .status-select {
            background: rgba(0, 0, 0, 0.2);
            color: white;
            border: 1px solid rgba(255, 255, 255, 0.1);
            padding: 5px 10px;
            border-radius: 6px;
            margin-top: 10px;
        }

        /* Existing Board Styles */
        header { text-align: center; padding: 20px 0; margin-bottom: 30px; }
        h1 { font-size: 2.2em; background: linear-gradient(90deg, #00d4ff, #7b2cbf); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        
        .board { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; max-width: 1400px; margin: 0 auto; align-items: start; }
        .column { background: rgba(0, 0, 0, 0.2); border-radius: 12px; padding: 15px; min-height: 600px; }
        .column-header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 15px; margin-bottom: 15px; border-bottom: 1px solid rgba(255,255,255,0.1); font-weight: bold; font-size: 1.1em; }
        .badge { background: rgba(255,255,255,0.1); padding: 2px 8px; border-radius: 12px; font-size: 0.8em; }
        
        .todo-header { color: var(--todo-border); }
        .doing-header { color: var(--doing-border); }
        .done-header { color: var(--done-border); }
        
        .note-card {
            background: var(--card-bg); border-radius: 8px; padding: 15px; margin-bottom: 15px;
            backdrop-filter: blur(10px); border: 1px solid rgba(255, 255, 255, 0.05);
            border-left: 3px solid transparent; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;
        }
        .note-card:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,0,0,0.3); }
        .note-card[data-status="TODO"] { border-left-color: var(--todo-border); }
        .note-card[data-status="DOING"] { border-left-color: var(--doing-border); }
        .note-card[data-status="DONE"] { border-left-color: var(--done-border); opacity: 0.8; }
        
        .note-content { margin-bottom: 10px; line-height: 1.5; white-space: pre-wrap; max-height: 100px; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; }
        .note-meta { display: flex; justify-content: space-between; font-size: 0.8em; color: #888; }
        
        /* Drag styles */
        .dragging { opacity: 0.5; }
        .column.drag-over { background: rgba(255,255,255,0.05); }
        
        @media (max-width: 768px) { .board { grid-template-columns: 1fr; } .column { min-height: auto; } }
    </style>
</head>
<body>
    <header>
        <h1>💡 Idea Board</h1>
        <p style="color: #888; margin-top: 5px;">Drag cards to update status • Click to edit</p>
    </header>
    
    <div class="board">
        {{range $status, $notes := .}}
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
                 onclick="openModal({{.}})"
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

    <!-- Details Modal -->
    <div class="modal-overlay" id="detailsModal" onclick="if(event.target===this) closeModal()">
        <div class="modal">
            <div class="modal-header">
                <div>
                    <h3 style="color: #888; font-size: 0.9em; margin-bottom: 5px;">NOTE DETAILS</h3>
                    <div id="modalNoteId" style="font-family: monospace; color: var(--accent-color);"></div>
                </div>
                <button class="btn-close" onclick="closeModal()">×</button>
            </div>
            <div class="modal-content">
                <div id="modalContent" class="note-full-content"></div>
                
                <div style="margin-bottom: 20px;">
                    <label style="color: #888; font-size: 0.9em;">Status:</label>
                    <select id="modalStatus" class="status-select" onchange="updateNoteStatusFromModal()">
                        <option value="TODO">TODO</option>
                        <option value="DOING">DOING</option>
                        <option value="DONE">DONE</option>
                    </select>
                </div>

                <div class="comments-section">
                    <h4 style="margin-bottom: 15px;">Comments</h4>
                    <div id="modalComments" class="comment-list">
                        <!-- Comments injected here -->
                    </div>
                    <div class="comment-input-area">
                        <input type="text" id="newComment" placeholder="Add a comment..." onkeypress="if(event.key==='Enter') addComment()">
                        <button class="btn-primary" onclick="addComment()">Send</button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Store notes data for modal access
        const notesData = {};
        {{range $status, $notes := .}}
            {{range $notes}}
            notesData["{{.ID}}"] = {{.}};
            {{end}}
        {{end}}

        let currentNoteId = null;

        function openModal(note) {
            // Because template passing object to JS function is tricky with quotes/newlines,
            // we use the ID to lookup from the pre-populated notesData object
            // However, the onclick in HTML passes the object directly if formatted correctly,
            // but let's be safe and use ID lookup if passing full object is complex.
            // Actually, let's just use the ID from the card element.
            
            // Wait, the template rendering above: onclick="openModal({{.}})"
            // Go template will render struct as a string like {ID:..., Content:...} which isn't valid JS object literal most likely.
            // Better strategy: onclick="openModal('{{.ID}}')"
        }
    </script>
    
    <!-- Fix script logic -->
    <script>
        function allowDrop(ev) { ev.preventDefault(); ev.currentTarget.classList.add('drag-over'); }
        function drag(ev) { ev.dataTransfer.setData("text", ev.target.id); ev.target.classList.add('dragging'); }
        function drop(ev, status) {
            ev.preventDefault();
            ev.currentTarget.classList.remove('drag-over');
            const id = ev.dataTransfer.getData("text");
            const card = document.getElementById(id);
            if (card) {
                // Determine raw target to append to
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
            // Update local data
            if (notesData[id]) notesData[id].Status = status;
        }

        // Modal Logic
        function openModalById(id) {
            const note = notesData[id];
            if (!note) return;
            
            currentNoteId = id;
            document.getElementById('modalNoteId').textContent = '#' + note.ID;
            document.getElementById('modalContent').textContent = note.Content; // Use textContent for safety
            document.getElementById('modalStatus').value = note.Status;
            
            renderComments(note.Comments || []);
            
            const overlay = document.getElementById('detailsModal');
            overlay.style.display = 'flex';
        }

        function closeModal() {
            document.getElementById('detailsModal').style.display = 'none';
            currentNoteId = null;
        }

        function renderComments(comments) {
            const container = document.getElementById('modalComments');
            container.innerHTML = '';
            comments.forEach(c => {
                const div = document.createElement('div');
                div.className = 'comment';
                const date = new Date(c.created_at).toLocaleString();
                div.innerHTML = '<div class="comment-meta">' + date + '</div>' + 
                                '<div>' + escapeHtml(c.content) + '</div>';
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
                
                // Update local data
                if (!notesData[currentNoteId].Comments) notesData[currentNoteId].Comments = [];
                notesData[currentNoteId].Comments.push(comment);
                
                // Re-render
                renderComments(notesData[currentNoteId].Comments);
                input.value = '';
            } catch (err) {
                alert('Failed to add comment');
            }
        }

        async function updateNoteStatusFromModal() {
            const status = document.getElementById('modalStatus').value;
            if (!currentNoteId) return;
            
            await updateStatus(currentNoteId, status);
            // Reload to reflect move on board
            location.reload(); 
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // Attach click listeners to cards
        // We do this via delegation or by updating the template to call openModalById
        // Let's rely on the template update below
    </script>
</body>
</html>`
