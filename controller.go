package rest

import (
	"encoding/json"
	"fmt"
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
	id, _ := strconv.Atoi(r.URL.Query().Get(":id"))
	entity, err := c.Repository.Read(int64(id))
	if err == ErrNotFound {
		msg := fmt.Sprintf("%s(id:%d) not found", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, 404, msg)
		return
	}
	if err != nil {
		c.errorf("reading %s(id:%d): %v", c.Repository.EntityName(), id, err)
		RespondWithError(w, 500, err.Error())
		return
	}
	RespondWithJSON(w, 200, &entity)
}

// GetAll handles the GET verb for the full collection
func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	options := c.parseOptions(r.URL.Query())
	entities, err := c.Repository.ReadAll(options)
	if err != nil {
		c.errorf("Error reading %s: %v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	count, _ := c.Repository.Count(options)
	w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
	RespondWithJSON(w, 200, &entities)
}

// Put handles the PUT verb
func (c *Controller) Put(w http.ResponseWriter, r *http.Request) {
	entity := c.Repository.NewInstance()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(entity); err != nil {
		c.errorf("parsing %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	err := c.Repository.Update(entity)
	if err == ErrNotFound {
		msg := fmt.Sprintf("%s not found", c.Repository.EntityName())
		c.warnf(msg)
		RespondWithError(w, 404, msg)
		return
	}
	if err != nil {
		c.errorf("updating %s: %v", c.Repository.EntityName(), err)
		RespondWithError(w, 500, err.Error())
		return
	}
	RespondWithJSON(w, 200, &entity)
}

// Post handles the POST verb
func (c *Controller) Post(w http.ResponseWriter, r *http.Request) {
	entity := c.Repository.NewInstance()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(entity); err != nil {
		c.errorf("parsing %s %#v", c.Repository.EntityName(), err)
		RespondWithError(w, http.StatusUnprocessableEntity, "Invalid request payload")
		return
	}
	id, err := c.Repository.Save(entity)
	if err != nil {
		c.errorf("saving %s %#v: %v", c.Repository.EntityName(), entity, err)
		RespondWithError(w, 500, err.Error())
		return
	}
	RespondWithJSON(w, 200, &map[string]int64{"id": id})
}

// Delete handles the DELETE verb
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get(":id"))
	err := c.Repository.Delete(int64(id))
	if err == ErrNotFound {
		msg := fmt.Sprintf("%s(id:%d) not found", c.Repository.EntityName(), id)
		c.warnf(msg)
		RespondWithError(w, 404, msg)
		return
	}
	if err != nil {
		c.errorf("deleting %s(id:%d): %v", c.Repository.EntityName(), id, err)
		RespondWithError(w, 500, err.Error())
		return
	}
	RespondWithJSON(w, 200, &map[string]string{})
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
		filters[k] = v[0]
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
