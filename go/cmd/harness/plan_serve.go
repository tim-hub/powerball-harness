package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)


//go:embed plan_web/dist
var planWebDist embed.FS

// ---------------------------------------------------------------------------
// Serve command
// ---------------------------------------------------------------------------

func runPlanServe(args []string) {
	port := 8888
	openBrowser := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			i++
			if i < len(args) {
				if p, err := strconv.Atoi(args[i]); err == nil {
					port = p
				}
			}
		case "--open":
			openBrowser = true
		}
	}

	planFile, err := planPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	registerPlanAPI(mux, planFile)

	// SPA fallback: serve static files, fall back to index.html
	distFS, err := fs.Sub(planWebDist, "plan_web/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: embed error: %v\n", err)
		os.Exit(1)
	}
	mux.Handle("/", spaHandler{fs: http.FS(distFS)})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("harness plan-cli serve listening on http://localhost%s\n", addr)

	if openBrowser {
		go func() {
			time.Sleep(200 * time.Millisecond)
			openURL(fmt.Sprintf("http://localhost%s", addr))
		}()
	}

	srv := &http.Server{Addr: addr, Handler: mux}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "serve: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// SPA handler — serves static files, falls back to index.html
// ---------------------------------------------------------------------------

type spaHandler struct {
	fs http.FileSystem
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Don't intercept API routes.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}
	// Try to serve the exact file.
	f, err := h.fs.Open(r.URL.Path)
	if err == nil {
		f.Close()
		http.FileServer(h.fs).ServeHTTP(w, r)
		return
	}
	// Fall back to index.html for client-side routing (SPA deep-links).
	r2 := *r
	r2.URL.Path = "/"
	http.FileServer(h.fs).ServeHTTP(w, &r2)
}

// ---------------------------------------------------------------------------
// REST API
// ---------------------------------------------------------------------------

func registerPlanAPI(mux *http.ServeMux, planFile string) {
	mux.HandleFunc("/api/phases", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiGetPhases(w, r, planFile)
		case http.MethodPost:
			apiPostPhase(w, r, planFile)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/phases/", func(w http.ResponseWriter, r *http.Request) {
		// /api/phases/:id
		// /api/phases/:id/tasks
		// /api/phases/:id/archive
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/phases/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		phaseID, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "invalid phase id", http.StatusBadRequest)
			return
		}
		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				apiGetPhase(w, r, planFile, phaseID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		switch parts[1] {
		case "tasks":
			if r.Method == http.MethodPost {
				apiPostTask(w, r, planFile, phaseID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		case "archive":
			if r.Method == http.MethodPost {
				apiArchivePhase(w, r, planFile, phaseID)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		if taskID == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			apiGetTask(w, r, planFile, taskID)
		case http.MethodPatch:
			apiPatchTask(w, r, planFile, taskID)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/comments/", func(w http.ResponseWriter, r *http.Request) {
		targetID := strings.TrimPrefix(r.URL.Path, "/api/comments/")
		if targetID == "" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		apiPostComment(w, r, planFile, targetID)
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// ── handlers ─────────────────────────────────────────────────────────────────

func apiGetPhases(w http.ResponseWriter, r *http.Request, planFile string) {
	setCORSHeaders(w)
	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, p.Phases)
}

func apiGetPhase(w http.ResponseWriter, r *http.Request, planFile string, phaseID int) {
	setCORSHeaders(w)
	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, ph := range p.Phases {
		if ph.ID == phaseID {
			writeJSON(w, ph)
			return
		}
	}
	http.NotFound(w, r)
}

func apiGetTask(w http.ResponseWriter, r *http.Request, planFile string, taskID string) {
	setCORSHeaders(w)
	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, ph := range p.Phases {
		for _, t := range ph.Tasks {
			if t.ID == taskID {
				writeJSON(w, t)
				return
			}
		}
	}
	http.NotFound(w, r)
}

func apiPatchTask(w http.ResponseWriter, r *http.Request, planFile string, taskID string) {
	setCORSHeaders(w)
	var fields map[string]string
	if err := decodeBody(r, &fields); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	found := false
	for pi := range p.Phases {
		for ti := range p.Phases[pi].Tasks {
			if p.Phases[pi].Tasks[ti].ID == taskID {
				t := &p.Phases[pi].Tasks[ti]
				if v, ok := fields["status"]; ok {
					t.Status = v
				}
				if v, ok := fields["urgency"]; ok {
					t.Urgency = v
				}
				if v, ok := fields["importance"]; ok {
					t.Importance = v
				}
				if v, ok := fields["statusHash"]; ok {
					t.StatusHash = v
				}
				if v, ok := fields["blockedReason"]; ok {
					t.BlockedReason = v
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		http.NotFound(w, r)
		return
	}

	if err := SavePlans(planFile, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiPostPhase(w http.ResponseWriter, r *http.Request, planFile string) {
	setCORSHeaders(w)
	var req struct {
		Title      string `json:"title"`
		Goal       string `json:"goal"`
		Urgency    string `json:"urgency"`
		Importance string `json:"importance"`
	}
	if err := decodeBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Title == "" || req.Goal == "" {
		http.Error(w, "title and goal required", http.StatusBadRequest)
		return
	}
	if req.Urgency == "" {
		req.Urgency = "medium"
	}
	if req.Importance == "" {
		req.Importance = "medium"
	}

	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	maxID := 0
	for _, ph := range p.Phases {
		if ph.ID > maxID {
			maxID = ph.ID
		}
	}
	ph := Phase{
		ID:         maxID + 1,
		Title:      req.Title,
		Created:    time.Now().Format("2006-01-02"),
		Goal:       req.Goal,
		Status:     "active",
		Urgency:    req.Urgency,
		Importance: req.Importance,
		Comments:   []Comment{},
		Tasks:      []Task{},
	}
	p.Phases = append([]Phase{ph}, p.Phases...)

	if err := SavePlans(planFile, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, ph)
}

func apiPostTask(w http.ResponseWriter, r *http.Request, planFile string, phaseID int) {
	setCORSHeaders(w)
	var req struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		DoD            string   `json:"dod"`
		Depends        []string `json:"depends"`
		Urgency        string   `json:"urgency"`
		Importance     string   `json:"importance"`
		QualityMarkers []string `json:"qualityMarkers"`
	}
	if err := decodeBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Name == "" || req.DoD == "" {
		http.Error(w, "name and dod required", http.StatusBadRequest)
		return
	}
	if req.Urgency == "" {
		req.Urgency = "medium"
	}
	if req.Importance == "" {
		req.Importance = "medium"
	}
	if req.Depends == nil {
		req.Depends = []string{}
	}
	if req.QualityMarkers == nil {
		req.QualityMarkers = []string{}
	}

	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	phIdx := -1
	for i, ph := range p.Phases {
		if ph.ID == phaseID {
			phIdx = i
			break
		}
	}
	if phIdx == -1 {
		http.NotFound(w, r)
		return
	}

	seq := len(p.Phases[phIdx].Tasks) + 1
	task := Task{
		ID:             fmt.Sprintf("%d.%d", phaseID, seq),
		Name:           req.Name,
		Description:    req.Description,
		DoD:            req.DoD,
		Depends:        req.Depends,
		Status:         "cc:TODO",
		Urgency:        req.Urgency,
		Importance:     req.Importance,
		QualityMarkers: req.QualityMarkers,
		Comments:       []Comment{},
	}
	p.Phases[phIdx].Tasks = append(p.Phases[phIdx].Tasks, task)

	if err := SavePlans(planFile, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, task)
}

func apiArchivePhase(w http.ResponseWriter, r *http.Request, planFile string, phaseID int) {
	setCORSHeaders(w)
	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	found := false
	for i := range p.Phases {
		if p.Phases[i].ID == phaseID {
			p.Phases[i].Status = "archived"
			found = true
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	if err := SavePlans(planFile, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiPostComment(w http.ResponseWriter, r *http.Request, planFile string, targetID string) {
	setCORSHeaders(w)
	var req struct {
		Text       string `json:"text"`
		Author     string `json:"author"`
		AuthorName string `json:"authorName"`
	}
	if err := decodeBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.Text == "" {
		http.Error(w, "text required", http.StatusBadRequest)
		return
	}
	if req.Author == "" {
		req.Author = "human"
	}

	p, err := LoadPlans(planFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := Comment{
		ID:         uuid.New().String(),
		Author:     req.Author,
		AuthorName: req.AuthorName,
		At:         time.Now().UTC().Format(time.RFC3339),
		Text:       req.Text,
	}

	found := false
	if strings.Contains(targetID, ".") {
		for pi := range p.Phases {
			for ti := range p.Phases[pi].Tasks {
				if p.Phases[pi].Tasks[ti].ID == targetID {
					p.Phases[pi].Tasks[ti].Comments = append(p.Phases[pi].Tasks[ti].Comments, c)
					found = true
				}
			}
		}
	} else {
		phID, err := strconv.Atoi(targetID)
		if err != nil {
			http.Error(w, "invalid target id", http.StatusBadRequest)
			return
		}
		for i := range p.Phases {
			if p.Phases[i].ID == phID {
				p.Phases[i].Comments = append(p.Phases[i].Comments, c)
				found = true
			}
		}
	}

	if !found {
		http.NotFound(w, r)
		return
	}

	if err := SavePlans(planFile, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, c)
}

// ---------------------------------------------------------------------------
// Browser launcher
// ---------------------------------------------------------------------------

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

