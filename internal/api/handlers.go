package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gomock/internal/entity"
	"gomock/internal/store"
)

// Handlers holds the dependencies needed to serve the API.
type Handlers struct {
	Store *store.Store
}

// NewHandlers builds a Handlers backed by the given store.
func NewHandlers(s *store.Store) *Handlers {
	return &Handlers{Store: s}
}

// typesEnvelope is the response body for GET /api/v1/types.
type typesEnvelope struct {
	Count   int              `json:"count"`
	Results []store.TypeInfo `json:"results"`
}

// indexResponse is the response body for GET /, a discovery document
// describing the API and linking to every currently loaded entity type.
type indexResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Types       map[string]string `json:"types"`
	TypesURL    string            `json:"typesUrl"`
}

// Index handles GET /, returning a small discovery document that
// describes the API and links to each currently loaded entity type, so
// clients can explore the API without prior knowledge of its data.
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	base := baseURL(r)

	types := h.Store.Types()
	links := make(map[string]string, len(types))
	for _, t := range types {
		links[t.Name] = base + "/api/v1/" + t.Name
	}

	writeJSON(w, http.StatusOK, indexResponse{
		Name:        "gomock",
		Description: "A generic, abstract mock-entity API. Each entry under \"types\" is a browsable REST resource.",
		Types:       links,
		TypesURL:    base + "/api/v1/types",
	})
}

// ListTypes handles GET /api/v1/types, listing every discovered entity
// type and how many records it currently holds.
func (h *Handlers) ListTypes(w http.ResponseWriter, r *http.Request) {
	types := h.Store.Types()
	writeJSON(w, http.StatusOK, typesEnvelope{Count: len(types), Results: types})
}

// ListEntities handles GET /api/v1/{type}, returning a filtered,
// paginated list of entities belonging to the requested type.
func (h *Handlers) ListEntities(w http.ResponseWriter, r *http.Request) {
	typeName := r.PathValue("type")

	items, ok := h.Store.List(typeName)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown entity type: "+typeName)
		return
	}

	filtered := applyFilters(r, items)
	env := paginate(r, filtered)
	writeJSON(w, http.StatusOK, env)
}

// GetEntity handles GET /api/v1/{type}/{idOrSlug}, returning a single
// entity looked up by numeric ID or by slug.
func (h *Handlers) GetEntity(w http.ResponseWriter, r *http.Request) {
	typeName := r.PathValue("type")
	idOrSlug := r.PathValue("idOrSlug")

	item, found, typeExists := h.Store.Get(typeName, idOrSlug)
	if !typeExists {
		writeError(w, http.StatusNotFound, "unknown entity type: "+typeName)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "entity not found: "+idOrSlug)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// ReplaceType handles PUT /api/v1/admin/{type}, replacing the entire
// in-memory dataset for typeName with the request body: a JSON array of
// flat objects adapted through the same rules used for on-disk data
// files. The change is not persisted to disk.
func (h *Handlers) ReplaceType(w http.ResponseWriter, r *http.Request) {
	typeName := r.PathValue("type")
	if strings.TrimSpace(typeName) == "" {
		writeError(w, http.StatusBadRequest, "entity type is required")
		return
	}

	var records []map[string]any
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: expected an array of objects: "+err.Error())
		return
	}

	entities := make([]entity.Entity, 0, len(records))
	for i, record := range records {
		ent, err := store.AdaptRecord(record)
		if err != nil {
			writeError(w, http.StatusBadRequest, "record "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		entities = append(entities, ent)
	}

	h.Store.Replace(typeName, entities)
	writeJSON(w, http.StatusOK, store.TypeInfo{Name: typeName, Count: len(entities)})
}
