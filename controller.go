package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
type Controller struct {
	Repository Repository
	Logger     Logger
}

// Get handles the GET verb for individual items.
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")
	entity, err := c.Repository.Read(id)
	switch {
	case err == ErrNotFound:
		msg := fmt.Sprintf("%s(id:%s) not found", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, http.StatusNotFound, msg)
		return
	case err == ErrPermissionDenied:
		msg := fmt.Sprintf("Reading %s(id:%s): Permission denied", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, http.StatusForbidden, msg)
		return
	case err != nil:
		c.errorf("Reading %s(id:%s): %v", c.Repository.EntityName(), id, err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, &entity)
}

// GetAll handles the GET verb for the full collection
func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	options := c.parseOptions(r.URL.Query())
	entities, err := c.Repository.ReadAll(options)
	if err == ErrPermissionDenied {
		msg := fmt.Sprintf("Error reading %s: Permission denied", c.Repository.EntityName())
		c.warnf(msg)
		RespondWithError(w, http.StatusForbidden, msg)
		return
	}
	if err != nil {
		c.errorf("Error reading %s: %v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	count, _ := c.Repository.Count(options)
	w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
	RespondWithJSON(w, http.StatusOK, &entities)
}

// Put handles the PUT verb
func (c *Controller) Put(w http.ResponseWriter, r *http.Request) {
	rp, ok := c.Repository.(Persistable)
	if !ok {
		RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		c.errorf("reading body for %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	r.Body.Close()
	entity := c.Repository.NewInstance()
	decoder := json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(bodyBytes)))
	if err := decoder.Decode(entity); err != nil {
		c.errorf("parsing %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	fields, err := c.getFieldNames(bodyBytes)
	if err != nil {
		c.errorf("parsing %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	id := r.URL.Query().Get(":id")
	err = rp.Update(id, entity, fields...)
	switch {
	case err == ErrNotFound:
		msg := fmt.Sprintf("%s not found", c.Repository.EntityName())
		c.warnf(msg)
		RespondWithError(w, http.StatusNotFound, msg)
		return
	case err == ErrPermissionDenied:
		msg := fmt.Sprintf("Updating %s: Permission denied", c.Repository.EntityName())
		c.warnf(msg)
		RespondWithError(w, http.StatusForbidden, msg)
		return
	case err != nil:
		if e, ok := err.(*ValidationError); ok {
			c.warnf("Updating %s: %v", c.Repository.EntityName(), e.Error())
			RespondWithJSON(w, http.StatusBadRequest, e)
		} else {
			c.errorf("Updating %s: %v", c.Repository.EntityName(), err)
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.Get(w, r)
}

func (c *Controller) getFieldNames(bytes []byte) ([]string, error) {
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
func (c *Controller) Post(w http.ResponseWriter, r *http.Request) {
	rp, ok := c.Repository.(Persistable)
	if !ok {
		RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	entity := c.Repository.NewInstance()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(entity); err != nil {
		c.errorf("parsing %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	id, err := rp.Save(entity)
	switch {
	case err == ErrPermissionDenied:
		msg := fmt.Sprintf("Saving %s: Permission denied", c.Repository.EntityName())
		c.warnf(msg)
		RespondWithError(w, http.StatusForbidden, msg)
		return
	case err != nil:
		if e, ok := err.(*ValidationError); ok {
			c.warnf("Saving %s: %v", c.Repository.EntityName(), e.Error())
			RespondWithJSON(w, http.StatusBadRequest, e)
		} else {
			c.errorf("Saving %s: %v", c.Repository.EntityName(), err)
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	RespondWithJSON(w, http.StatusOK, &map[string]string{"id": id})
}

// Delete handles the DELETE verb
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	rp, ok := c.Repository.(Persistable)
	if !ok {
		RespondWithError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	id := r.URL.Query().Get(":id")
	err := rp.Delete(id)
	switch {
	case err == ErrNotFound:
		msg := fmt.Sprintf("%s(id:%s) not found", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, http.StatusNotFound, msg)
		return
	case err == ErrPermissionDenied:
		msg := fmt.Sprintf("Deleting %s(id:%s): Permission denied", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, http.StatusForbidden, msg)
		return
	case err != nil:
		c.errorf("Deleting %s(id:%s): %v", c.Repository.EntityName(), id, err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, &map[string]string{})
}

func (c *Controller) parseFilters(params url.Values) map[string]interface{} {
	var filterStr = params.Get("_filters")
	filters := make(map[string]interface{})
	if filterStr != "" {
		filterStr, _ = url.QueryUnescape(filterStr)
		if err := json.Unmarshal([]byte(filterStr), &filters); err != nil {
			c.warnf("Invalid filter specification: %s - %v", filterStr, err)
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
	return filters
}

func (c *Controller) parseOptions(params url.Values) QueryOptions {
	start, _ := strconv.Atoi(params.Get("_start"))
	end, _ := strconv.Atoi(params.Get("_end"))

	sortField := params.Get("_sort")
	sortDir := params.Get("_order")

	return QueryOptions{
		Sort:    sortField,
		Order:   strings.ToLower(sortDir),
		Offset:  start,
		Max:     int(math.Max(0, float64(end-start))),
		Filters: c.parseFilters(params),
	}
}

func (c *Controller) warnf(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Warnf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (c *Controller) errorf(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Errorf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
