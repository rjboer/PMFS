package main

import (
	"embed"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	PMFS "github.com/rjboer/PMFS"
)

// server implements a tiny REST interface for PMFS using only the
// standard library. It mirrors the endpoints described in
// examples/webinterface/architecture.md and requirements.md.
type server struct {
	db   *PMFS.Database
	mu   sync.Mutex
	subs map[int][]chan struct{}
}

//go:embed index.html
var staticFS embed.FS

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dir := flag.String("dir", "webdata", "data directory")
	flag.Parse()

	db, err := PMFS.LoadSetup(*dir)
	if err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}
	s := &server{db: db, subs: make(map[int][]chan struct{})}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := staticFS.ReadFile("index.html")
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(b)
	})
	mux.HandleFunc("/products", s.handleProducts)
	mux.HandleFunc("/products/", s.handleProducts)
	mux.HandleFunc("/projects/", s.handleProjects)
	mux.HandleFunc("/requirements/", s.handleRequirements)

	log.Printf("PMFS web interface listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// roleFromRequest extracts the user's role from the X-Role header.
func roleFromRequest(r *http.Request) string {
	role := r.Header.Get("X-Role")
	switch role {
	case "viewer", "editor", "admin":
		return role
	default:
		return ""
	}
}

// authorize ensures the user has the required permission for the action.
func authorize(w http.ResponseWriter, r *http.Request, action string) bool {
	role := roleFromRequest(r)
	if role == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	switch action {
	case "read":
		return true
	case "write":
		if role == "editor" || role == "admin" {
			return true
		}
	case "delete":
		if role == "admin" {
			return true
		}
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}

// respondJSON writes v as JSON to w.
func respondJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// helper functions to locate products and projects ---------------------------------

func (s *server) findProduct(id int) (*PMFS.ProductType, error) {
	for i := range s.db.Products {
		if s.db.Products[i].ID == id {
			return &s.db.Products[i], nil
		}
	}
	return nil, fmt.Errorf("product not found")
}

func (s *server) findProject(id int) (*PMFS.ProjectType, error) {
	for i := range s.db.Products {
		for j := range s.db.Products[i].Projects {
			if s.db.Products[i].Projects[j].ID == id {
				prd := &s.db.Products[i]
				prj := &prd.Projects[j]
				prj.ProductID = prd.ID
				if err := prj.Load(); err != nil {
					return nil, err
				}
				return prj, nil
			}
		}
	}
	return nil, fmt.Errorf("project not found")
}

func (s *server) findProjectByRequirement(rid int) (*PMFS.ProjectType, *PMFS.Requirement, error) {
	for i := range s.db.Products {
		for j := range s.db.Products[i].Projects {
			prj := &s.db.Products[i].Projects[j]
			prj.ProductID = s.db.Products[i].ID
			if err := prj.Load(); err != nil {
				return nil, nil, err
			}
			for k := range prj.D.Requirements {
				if prj.D.Requirements[k].ID == rid {
					return prj, &prj.D.Requirements[k], nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("requirement not found")
}

// notifySubscribers broadcasts a change event for a project.
func (s *server) notifySubscribers(projectID int) {
	s.mu.Lock()
	subs := s.subs[projectID]
	s.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// ------------------------- Product handlers ---------------------------------------

func (s *server) handleProducts(w http.ResponseWriter, r *http.Request) {
	if !authorize(w, r, map[string]string{http.MethodGet: "read", http.MethodPost: "write", http.MethodPut: "write", http.MethodDelete: "delete"}[r.Method]) {
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/products")
	path = strings.Trim(path, "/")
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, s.db.Products)
		case http.MethodPost:
			var pd PMFS.ProductData
			if err := json.NewDecoder(r.Body).Decode(&pd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := s.db.NewProduct(pd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]int{"id": id})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	segs := strings.Split(path, "/")
	id, err := strconv.Atoi(segs[0])
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	prd, err := s.findProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if len(segs) == 1 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, prd)
		case http.MethodPut:
			var pd PMFS.ProductData
			if err := json.NewDecoder(r.Body).Decode(&pd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			prd.Name = pd.Name
			if err := s.db.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, prd)
		case http.MethodDelete:
			// remove product
			for i := range s.db.Products {
				if s.db.Products[i].ID == id {
					s.db.Products = append(s.db.Products[:i], s.db.Products[i+1:]...)
					break
				}
			}
			if err := s.db.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if segs[1] == "projects" {
		s.handleProjectsForProduct(w, r, prd, segs[2:])
		return
	}
	http.NotFound(w, r)
}

func (s *server) handleProjectsForProduct(w http.ResponseWriter, r *http.Request, prd *PMFS.ProductType, segs []string) {
	if len(segs) == 0 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, prd.Projects)
		case http.MethodPost:
			var pd PMFS.ProjectData
			if err := json.NewDecoder(r.Body).Decode(&pd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := prd.NewProject(pd)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := s.db.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, map[string]int{"id": id})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	id, err := strconv.Atoi(segs[0])
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	prj, err := prd.Project(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(segs) == 1 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, prj)
		case http.MethodPut:
			var pd PMFS.ProjectData
			if err := json.NewDecoder(r.Body).Decode(&pd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if pd.Name != "" {
				prj.Name = pd.Name
			}
			if pd.Scope != "" {
				prj.D.Scope = pd.Scope
			}
			if err := prj.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := s.db.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, prj)
		case http.MethodDelete:
			for i := range prd.Projects {
				if prd.Projects[i].ID == id {
					prd.Projects = append(prd.Projects[:i], prd.Projects[i+1:]...)
					break
				}
			}
			if err := s.db.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	http.NotFound(w, r)
}

// ------------------------- Project-level handlers ---------------------------------

func (s *server) handleProjects(w http.ResponseWriter, r *http.Request) {
	if !authorize(w, r, map[string]string{http.MethodGet: "read", http.MethodPost: "write", http.MethodPut: "write", http.MethodDelete: "delete"}[r.Method]) {
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	segs := strings.Split(strings.Trim(path, "/"), "/")
	if segs[0] == "" {
		http.NotFound(w, r)
		return
	}
	prid, err := strconv.Atoi(segs[0])
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	prj, err := s.findProject(prid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if len(segs) == 1 {
		if r.Method == http.MethodGet {
			respondJSON(w, prj)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch segs[1] {
	case "requirements":
		s.handleRequirementsForProject(w, r, prj, segs[2:])
	case "export":
		s.handleProjectExport(w, r, prj, segs[2:])
	case "import":
		s.handleProjectImport(w, r, prj, segs[2:])
	case "design":
		if r.Method == http.MethodGet {
			respondJSON(w, prj.D.Intelligence)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	case "struct":
		if len(segs) == 2 {
			s.handleProjectStruct(w, r, prj)
			return
		}
		if len(segs) == 3 && segs[2] == "subscribe" {
			s.handleProjectSubscribe(w, r, prj)
			return
		}
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) handleProjectStruct(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType) {
	q := r.URL.Query()
	depth, _ := strconv.Atoi(q.Get("depth"))
	status := q.Get("status")
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	reqs := prj.D.Requirements
	if status != "" {
		filtered := reqs[:0]
		for _, r := range reqs {
			ok := false
			switch status {
			case "active":
				ok = r.Condition.Active
			case "deleted":
				ok = r.Condition.Deleted
			case "proposed":
				ok = r.Condition.Proposed
			default:
				ok = true
			}
			if ok {
				filtered = append(filtered, r)
			}
		}
		reqs = filtered
	}
	if depth > 0 {
		filtered := reqs[:0]
		for _, r := range reqs {
			if r.Level <= depth {
				filtered = append(filtered, r)
			}
		}
		reqs = filtered
	}
	start := (page - 1) * pageSize
	if start > len(reqs) {
		reqs = []PMFS.Requirement{}
	} else {
		end := start + pageSize
		if end > len(reqs) {
			end = len(reqs)
		}
		reqs = reqs[start:end]
	}

	type response struct {
		Product      int                `json:"product"`
		ProductName  string             `json:"product_name"`
		Project      int                `json:"project"`
		ProjectName  string             `json:"project_name"`
		Requirements []PMFS.Requirement `json:"requirements"`
		Attachments  []PMFS.Attachment  `json:"attachments"`
	}
	prdName := ""
	for _, p := range s.db.Products {
		if p.ID == prj.ProductID {
			prdName = p.Name
		}
	}
	resp := response{
		Product:      prj.ProductID,
		ProductName:  prdName,
		Project:      prj.ID,
		ProjectName:  prj.Name,
		Requirements: reqs,
		Attachments:  prj.D.Attachments,
	}
	respondJSON(w, resp)
}

func (s *server) handleProjectSubscribe(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType) {
	if f, ok := w.(http.Flusher); ok {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		ch := make(chan struct{}, 1)
		s.mu.Lock()
		s.subs[prj.ID] = append(s.subs[prj.ID], ch)
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			subs := s.subs[prj.ID]
			for i := range subs {
				if subs[i] == ch {
					s.subs[prj.ID] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
			s.mu.Unlock()
		}()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ch:
				fmt.Fprintf(w, "event: update\ndata: %d\n\n", time.Now().Unix())
				f.Flush()
			}
		}
	} else {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
	}
}

func (s *server) handleProjectExport(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType, segs []string) {
	if len(segs) == 0 {
		http.NotFound(w, r)
		return
	}
	switch segs[0] {
	case "excel":
		tmp := filepath.Join(os.TempDir(), fmt.Sprintf("project_%d.xlsx", prj.ID))
		if err := prj.ExportExcel(tmp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmp)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=project.xlsx")
		f, _ := os.Open(tmp)
		defer f.Close()
		io.Copy(w, f)
	case "struct":
		s.handleProjectStruct(w, r, prj)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=project.csv")
		cw := csv.NewWriter(w)
		header := []string{"ID", "Name", "Description", "Priority", "Status"}
		cw.Write(header)
		for _, req := range prj.D.Requirements {
			cw.Write([]string{
				strconv.Itoa(req.ID),
				req.Name,
				req.Description,
				strconv.Itoa(req.Priority),
				req.Status,
			})
		}
		cw.Flush()
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=project.pdf")
		// Very small PDF with requirement names.
		var content strings.Builder
		content.WriteString("Project: " + prj.Name + "\n")
		for _, r := range prj.D.Requirements {
			content.WriteString(fmt.Sprintf("- %s\n", r.Name))
		}
		pdf := minimalPDF(content.String())
		w.Write(pdf)
	default:
		http.NotFound(w, r)
	}
}

func minimalPDF(text string) []byte {
	// Extremely small PDF generator for demonstration.
	text = strings.ReplaceAll(text, "(", "[")
	text = strings.ReplaceAll(text, ")", "]")
	stream := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", strings.ReplaceAll(text, "\n", "\\n"))
	pdf := fmt.Sprintf("%%PDF-1.1\n1 0 obj<<>>endobj\n2 0 obj<< /Length %d >>stream\n%s\nendstream\nendobj\n3 0 obj<< /Type /Page /Parent 4 0 R /Contents 2 0 R >>endobj\n4 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n5 0 obj<< /Type /Catalog /Pages 4 0 R >>endobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000034 00000 n \n0000000%d 00000 n \n0000000%d 00000 n \n0000000%d 00000 n \ntrailer<< /Size 6 /Root 5 0 R >>\nstartxref\n0\n%%EOF", len(stream), stream, 34+len(stream)+39, 34+len(stream)+39+41, 34+len(stream)+39+41+44)
	return []byte(pdf)
}

func (s *server) handleProjectImport(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType, segs []string) {
	if len(segs) == 0 || segs[0] != "excel" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	tmp := filepath.Join(os.TempDir(), header.Filename)
	out, err := os.Create(tmp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		out.Close()
		return
	}
	out.Close()
	defer os.Remove(tmp)

	if err := prj.ImportExcel(tmp, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := prj.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.notifySubscribers(prj.ID)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleRequirementsForProject(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType, segs []string) {
	switch r.Method {
	case http.MethodGet:
		if len(segs) == 0 {
			respondJSON(w, prj.D.Requirements)
			return
		}
	}
	if len(segs) == 0 {
		if r.Method == http.MethodPost {
			var req PMFS.Requirement
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := prj.AddRequirement(req); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.notifySubscribers(prj.ID)
			respondJSON(w, req)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(segs[0])
	if err != nil {
		http.Error(w, "invalid requirement id", http.StatusBadRequest)
		return
	}
	var req *PMFS.Requirement
	for i := range prj.D.Requirements {
		if prj.D.Requirements[i].ID == id {
			req = &prj.D.Requirements[i]
			break
		}
	}
	if req == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if len(segs) == 1 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, req)
		case http.MethodPut:
			var upd PMFS.Requirement
			if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			upd.ID = req.ID
			*req = upd
			if err := prj.Save(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.notifySubscribers(prj.ID)
			respondJSON(w, req)
		case http.MethodDelete:
			prj.DeleteRequirementByID(id)
			s.notifySubscribers(prj.ID)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	http.NotFound(w, r)
}

// --------------------- Requirement-level endpoints --------------------------------

func (s *server) handleRequirements(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/requirements/")
	segs := strings.Split(strings.Trim(path, "/"), "/")
	if len(segs) == 0 || segs[0] == "" {
		http.NotFound(w, r)
		return
	}
	rid, err := strconv.Atoi(segs[0])
	if err != nil {
		http.Error(w, "invalid requirement id", http.StatusBadRequest)
		return
	}
	prj, req, err := s.findProjectByRequirement(rid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(segs) == 1 {
		switch r.Method {
		case http.MethodGet:
			respondJSON(w, req)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	switch segs[1] {
	case "attachments":
		s.handleRequirementAttachments(w, r, prj, req, segs[2:])
	case "analyze":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pass, ans, err := req.Analyze("system", "clarity-form-1")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Condition.AIanalyzed = true
		_ = prj.Save()
		respondJSON(w, map[string]interface{}{"pass": pass, "answer": ans})
	case "suggestions":
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		reqs, err := req.SuggestOthers(prj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.notifySubscribers(prj.ID)
		respondJSON(w, reqs)
	default:
		http.NotFound(w, r)
	}
}

func projectInputDir(prj *PMFS.ProjectType) string {
	return filepath.Join(os.Getenv("PMFS_BASEDIR"), "products", strconv.Itoa(prj.ProductID), "projects", strconv.Itoa(prj.ID), "input")
}

func (s *server) handleRequirementAttachments(w http.ResponseWriter, r *http.Request, prj *PMFS.ProjectType, req *PMFS.Requirement, segs []string) {
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		inputDir := projectInputDir(prj)
		if err := os.MkdirAll(inputDir, 0o755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dst := filepath.Join(inputDir, header.Filename)
		out, err := os.Create(dst)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			out.Close()
			return
		}
		out.Close()
		att, err := prj.AddAttachmentFromInput(inputDir, header.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.AttachmentIndex = len(prj.D.Attachments) - 1
		if err := prj.Save(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.notifySubscribers(prj.ID)
		respondJSON(w, att)
	case http.MethodGet:
		if len(segs) < 1 {
			http.Error(w, "attachment id required", http.StatusBadRequest)
			return
		}
		aid, err := strconv.Atoi(segs[0])
		if err != nil {
			http.Error(w, "invalid attachment id", http.StatusBadRequest)
			return
		}
		var att *PMFS.Attachment
		for i := range prj.D.Attachments {
			if prj.D.Attachments[i].ID == aid {
				att = &prj.D.Attachments[i]
				break
			}
		}
		if att == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		full := filepath.Join(projectDir(prj.ProductID, prj.ID), att.RelPath)
		http.ServeFile(w, r, full)
	case http.MethodDelete:
		if len(segs) < 1 {
			http.Error(w, "attachment id required", http.StatusBadRequest)
			return
		}
		aid, err := strconv.Atoi(segs[0])
		if err != nil {
			http.Error(w, "invalid attachment id", http.StatusBadRequest)
			return
		}
		for i := range prj.D.Attachments {
			if prj.D.Attachments[i].ID == aid {
				full := filepath.Join(projectDir(prj.ProductID, prj.ID), prj.D.Attachments[i].RelPath)
				os.Remove(full)
				prj.D.Attachments = append(prj.D.Attachments[:i], prj.D.Attachments[i+1:]...)
				break
			}
		}
		if req.AttachmentIndex >= len(prj.D.Attachments) {
			req.AttachmentIndex = -1
		}
		if err := prj.Save(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.notifySubscribers(prj.ID)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// projectDir reimplements PMFS's internal path helpers for the example.
func projectDir(productID, projectID int) string {
	base := os.Getenv("PMFS_BASEDIR")
	return filepath.Join(base, "products", strconv.Itoa(productID), "projects", strconv.Itoa(projectID))
}

// ----------------------------------------------------------------------------------
