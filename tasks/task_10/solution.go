package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
	"uuid"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	newInMemoryTaskRepo := &inMemoryTaskRepo{}

	newInMemoryTaskRepo.clock = clock
	newInMemoryTaskRepo.tasks = make(map[string]Task)

	return newInMemoryTaskRepo
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	newTask := Task{}

	title = strings.TrimSpace(title)

	if title == "" {
		return newTask, ErrInvalidTitle
	}

	newTask.ID = uuid.New().String()
	newTask.Title = title
	newTask.Done = false
	newTask.UpdatedAt = r.clock.Now()

	r.mu.Lock()
	r.tasks[newTask.ID] = newTask
	r.mu.Unlock()

	return newTask, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, false
	}
	return task, true
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := []Task{}
	for _, task := range r.tasks {
		if task.Done == done {
			tasks = append(tasks, task)
		}
	}

	slices.SortFunc(tasks, func(i, j Task) int {
		if i.UpdatedAt.Equal(j.UpdatedAt) {
			return cmp.Compare(i.ID, j.ID)
		}
		return j.UpdatedAt.Compare(i.UpdatedAt)
	})

	return tasks
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()
	}

	r.tasks[id] = task
	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo}
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
			return
		case http.MethodGet:
			h.handleList(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	} else if strings.HasPrefix(path, "/tasks/") {
		rest := strings.TrimPrefix(path, "/tasks/")

		if rest == "" || strings.Contains(rest, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, rest)
			return
		case http.MethodPatch:
			h.handlePatch(w, r, rest)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	} else {
		http.NotFound(w, r)
	}
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {

	var title struct {
		Title *string `json:"title"`
	}

	err := decodeStrictJSON(r.Body, &title)
	if err != nil || title.Title == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.Create(*title.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()

	var done bool

	values, ok := param["done"]

	if ok {
		if len(values) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		value := values[0]
		switch value {
		case "true":
			done = true
		case "false":
			done = false
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	writeJSON(w, http.StatusOK, h.repo.List(done))
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {

	var taskIsDone struct {
		Done *bool `json:"done"`
	}

	err := decodeStrictJSON(r.Body, &taskIsDone)
	if err != nil || taskIsDone.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *taskIsDone.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, task)

}
func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return err
	}

	rawMsg := json.RawMessage{}

	err = decoder.Decode(&rawMsg)
	if err == io.EOF {
		return nil
	}
	return errors.New("unexpected data after JSON")
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
