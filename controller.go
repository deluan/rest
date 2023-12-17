package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

/*
Controller implements a set of RESTful handlers, compatible with the JSON Server API "dialect". Please prefer to use
the functions provided in the handler.go file instead of these.
*/
type Controller[T any] struct {
	Repository Repository[T]
}

// Get handles the GET verb for individual items.
func (c *Controller[T]) Get(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	entity, err := c.Repository.Read(r.Context(), id)
	switch {
	case err == nil:
		_ = RespondWithJSON(w, http.StatusOK, entity)
	case errors.Is(err, ErrNotFound):
		_ = RespondWithError(w, http.StatusNotFound, fmt.Sprintf("%s(id:%s) not found", c.entityName(), id))
	case errors.Is(err, ErrPermissionDenied):
		_ = RespondWithError(w, http.StatusForbidden, fmt.Sprintf("Reading %s(id:%s): Permission denied", c.entityName(), id))
	default:
		_ = RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

// GetAll handles the GET verb for the full collection
func (c *Controller[T]) GetAll(w http.ResponseWriter, r *http.Request) {
	options, err := c.parseOptions(r.URL.Query())
	if err != nil {
		_ = RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	entities, err := c.Repository.ReadAll(r.Context(), options)
	switch {
	case err == nil:
		count, _ := c.Repository.Count(r.Context(), options)
		w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
		if len(entities) == 0 {
			_ = RespondWithJSON(w, http.StatusOK, []string{})
		} else {
			_ = RespondWithJSON(w, http.StatusOK, &entities)
		}
	case errors.Is(err, ErrPermissionDenied):
		_ = RespondWithError(w, http.StatusForbidden, fmt.Sprintf("Error reading %s: Permission denied", c.entityName()))
	default:
		_ = RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

// Put handles the PUT verb
func (c *Controller[T]) Put(w http.ResponseWriter, r *http.Request) {
	repo, ok := c.Repository.(Persistable[T])
	if !ok {
		_ = RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		_ = RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	var entity T
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	if err := decoder.Decode(&entity); err != nil {
		_ = RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	fields, err := c.getFieldNames(bodyBytes)
	if err != nil {
		_ = RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	id := r.URL.Query().Get(":id")
	err = repo.Update(r.Context(), id, entity, fields...)
	var vErr *ValidationError
	switch {
	case err == nil:
		c.Get(w, r)
	case errors.Is(err, ErrNotFound):
		_ = RespondWithError(w, http.StatusNotFound, fmt.Sprintf("%s not found", c.entityName()))
	case errors.Is(err, ErrPermissionDenied):
		_ = RespondWithError(w, http.StatusForbidden, fmt.Sprintf("Updating %s: Permission denied", c.entityName()))
	case errors.As(err, &vErr):
		_ = RespondWithJSON(w, http.StatusBadRequest, vErr)
	default:
		_ = RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (c *Controller[T]) getFieldNames(bytes []byte) ([]string, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	var fields []string
	for k := range m {
		fields = append(fields, k)
	}
	return fields, nil
}

// Post handles the POST verb
func (c *Controller[T]) Post(w http.ResponseWriter, r *http.Request) {
	repo, ok := c.Repository.(Persistable[T])
	if !ok {
		_ = RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	var entity T
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entity); err != nil {
		_ = RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	id, err := repo.Save(r.Context(), &entity)
	var vErr *ValidationError
	switch {
	case err == nil:
		_ = RespondWithJSON(w, http.StatusOK, &map[string]string{"id": id})
	case errors.Is(err, ErrPermissionDenied):
		_ = RespondWithError(w, http.StatusForbidden, fmt.Sprintf("Saving %s: Permission denied", c.entityName()))
	case errors.As(err, &vErr):
		_ = RespondWithJSON(w, http.StatusBadRequest, vErr)
	default:
		_ = RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

// Delete handles the DELETE verb
func (c *Controller[T]) Delete(w http.ResponseWriter, r *http.Request) {
	repo, ok := c.Repository.(Persistable[T])
	if !ok {
		_ = RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	ids := r.URL.Query()[":id"]
	err := repo.Delete(r.Context(), ids...)
	switch {
	case err == nil:
		_ = RespondWithJSON(w, http.StatusOK, &map[string]string{})
	case errors.Is(err, ErrNotFound):
		_ = RespondWithError(w, http.StatusNotFound, fmt.Sprintf("%s(id:%s) not found", c.entityName(), ids))
	case errors.Is(err, ErrPermissionDenied):
		_ = RespondWithError(w, http.StatusForbidden, fmt.Sprintf("Deleting %s(id:%s): Permission denied", c.entityName(), ids))
	default:
		_ = RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (c *Controller[T]) entityName() string {
	return strings.TrimPrefix(fmt.Sprintf("%T", (*T)(nil)), "*")
}

func (c *Controller[T]) parseFilters(params url.Values) (map[string]any, error) {
	var filterStr = params.Get("_filters")
	filters := make(map[string]any)
	if filterStr != "" {
		filterStr, _ = url.QueryUnescape(filterStr)
		if err := json.Unmarshal([]byte(filterStr), &filters); err != nil {
			return nil, err
		}
	}
	for k, v := range params {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if len(v) == 1 {
			filters[k] = v[0]
		} else {
			filters[k] = v
		}
	}
	return filters, nil
}

func (c *Controller[T]) parseOptions(params url.Values) (QueryOptions, error) {
	start, _ := strconv.Atoi(params.Get("_start"))
	end, _ := strconv.Atoi(params.Get("_end"))

	sortField := params.Get("_sort")
	sortDir := params.Get("_order")
	filters, err := c.parseFilters(params)
	if err != nil {
		return QueryOptions{}, err
	}
	return QueryOptions{
		Sort:    sortField,
		Order:   strings.ToLower(sortDir),
		Offset:  start,
		Max:     int(math.Max(0, float64(end-start))),
		Filters: filters,
	}, nil
}
