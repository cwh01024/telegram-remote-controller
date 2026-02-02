package notes

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// NoteStatus represents the status of a note
type NoteStatus string

const (
	StatusTodo       NoteStatus = "TODO"
	StatusInProgress NoteStatus = "DOING"
	StatusDone       NoteStatus = "DONE"
)

// Note represents a single idea/note
type Note struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	Status    NoteStatus `json:"status"`
	Tags      []string   `json:"tags,omitempty"`
}

// Store manages notes persistence
type Store struct {
	filePath string
	notes    []Note
	mutex    sync.RWMutex
}

// NewStore creates a new notes store
func NewStore() *Store {
	notesDir := "/Users/applejobs/.gemini/antigravity/scratch/telegram-agent-controller/notes"
	os.MkdirAll(notesDir, 0755)

	store := &Store{
		filePath: filepath.Join(notesDir, "ideas.json"),
		notes:    []Note{},
	}

	// Load existing notes
	store.load()

	return store
}

// Add adds a new note
func (s *Store) Add(content string, tags ...string) Note {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	note := Note{
		ID:        generateID(),
		Content:   content,
		CreatedAt: time.Now(),
		Status:    StatusTodo,
		Tags:      tags,
	}

	s.notes = append(s.notes, note)
	s.save()

	log.Printf("Added note: %s", note.ID)
	return note
}

// UpdateStatus updates the status of a note
func (s *Store) UpdateStatus(id string, status NoteStatus) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, note := range s.notes {
		if note.ID == id {
			s.notes[i].Status = status
			s.save()
			log.Printf("Updated note %s status to %s", id, status)
			return true
		}
	}
	return false
}

// GetAll returns all notes sorted by creation time (newest first)
func (s *Store) GetAll() []Note {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Make a copy and sort
	notes := make([]Note, len(s.notes))
	copy(notes, s.notes)

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].CreatedAt.After(notes[j].CreatedAt)
	})

	return notes
}

// Delete removes a note by ID
func (s *Store) Delete(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, note := range s.notes {
		if note.ID == id {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			s.save()
			return true
		}
	}
	return false
}

// Count returns the number of notes
func (s *Store) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.notes)
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error loading notes: %v", err)
		}
		return
	}

	if err := json.Unmarshal(data, &s.notes); err != nil {
		log.Printf("Error parsing notes: %v", err)
	}
}

func (s *Store) save() {
	data, err := json.MarshalIndent(s.notes, "", "  ")
	if err != nil {
		log.Printf("Error marshaling notes: %v", err)
		return
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("Error saving notes: %v", err)
	}
}

func generateID() string {
	return time.Now().Format("20060102-150405")
}
