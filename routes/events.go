package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/b-j-roberts/foc-engine/internal/db/mongo"
	routeutils "github.com/b-j-roberts/foc-engine/routes/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitEventsRoutes() {
	http.HandleFunc("/events/get-block-events", GetBlockEvents)
  http.HandleFunc("/events/get-latest-event", GetLatestEvent)
  http.HandleFunc("/events/get-latest-with", GetLatestWith)
  http.HandleFunc("/events/get-events-ordered", GetEventsOrdered)
  http.HandleFunc("/events/get-unique-ordered", GetUniqueOrdered)
}

func GetBlockEvents(w http.ResponseWriter, r *http.Request) {
  blockNumberStr := r.URL.Query().Get("blockNumber")
  if blockNumberStr == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing blockNumber parameter")
    return
  }
  blockNumber, err := strconv.Atoi(blockNumberStr)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid blockNumber parameter")
    return
  }

  res, err := mongo.GetFocEngineEventsCollection().Find(r.Context(), map[string]interface{}{
    "block_number": blockNumber,
  })
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query events")
    return
  }
  defer res.Close(r.Context())

  var events []map[string]interface{}
  for res.Next(r.Context()) {
    var event map[string]interface{}
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
      return
    }
    events = append(events, event)
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over events")
    return
  }
  if len(events) == 0 {
    // TODO: Return empty array instead of 404
    routeutils.WriteErrorJson(w, http.StatusNotFound, "No events found for the specified block number")
    return
  }

  // Convert events to JSON
  eventsJson, err := json.Marshal(events)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal events to JSON")
    return
  }
  // Write the JSON response
  routeutils.WriteDataJson(w, string(eventsJson))
}

func GetLatestEvent(w http.ResponseWriter, r *http.Request) {
  contractAddress := r.URL.Query().Get("contractAddress")
  if contractAddress == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing contractAddress parameter")
    return
  }

  eventType := r.URL.Query().Get("eventType")
  if eventType == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing eventType parameter")
    return
  }

  findOptions := options.Find().SetSort(map[string]interface{}{
    "_id": -1,
  }).SetLimit(1)
  res, err := mongo.GetFocEngineEventsCollection().Find(r.Context(), map[string]interface{}{
    "contract_address": contractAddress,
    "event_type":      eventType,
  }, findOptions)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query events")
    return
  }
  defer res.Close(r.Context())

  var event map[string]interface{}
  if res.Next(r.Context()) {
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
      return
    }
  } else {
    // TODO: Return empty array instead of 404
    routeutils.WriteErrorJson(w, http.StatusNotFound, "No events found for the specified contract address and event type")
    return
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over events")
    return
  }
  // Convert event to JSON
  eventJson, err := json.Marshal(event)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal event to JSON")
    return
  }
  // Write the JSON response
  routeutils.WriteDataJson(w, string(eventJson))
}

func GetLatestWith(w http.ResponseWriter, r *http.Request) {
  contractAddress := r.URL.Query().Get("contractAddress")
  if contractAddress == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing contractAddress parameter")
    return
  }

  eventType := r.URL.Query().Get("eventType")
  if eventType == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing eventType parameter")
    return
  }

  // TODO: Nest filters in the body
  // Filters from body
  userFilters := r.Body
  if userFilters == nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing filters parameter")
    return
  }
  defer userFilters.Close()
  // Parse filters ( example: {"key":"value","key2":"value2"} )
  var filters map[string]interface{}
  if err := json.NewDecoder(userFilters).Decode(&filters); err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid filters parameter")
    return
  }
  // Add contract address and event type to the filters
  filters["contract_address"] = contractAddress
  filters["event_type"] = eventType

  findOptions := options.Find().SetSort(map[string]interface{}{
    "_id": -1,
  }).SetLimit(1)
  res, err := mongo.GetFocEngineEventsCollection().Find(r.Context(), filters, findOptions)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query events")
    return
  }
  defer res.Close(r.Context())

  var event map[string]interface{}
  if res.Next(r.Context()) {
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
      return
    }
  } else {
    // TODO: Return empty array instead of 404
    routeutils.WriteErrorJson(w, http.StatusNotFound, "No events found for the specified contract address and event type")
    return
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over events")
    return
  }
  // Convert event to JSON
  eventJson, err := json.Marshal(event)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal event to JSON")
    return
  }
  // Write the JSON response
  routeutils.WriteDataJson(w, string(eventJson))
}

func GetEventsOrdered(w http.ResponseWriter, r *http.Request) {
  contractAddress := r.URL.Query().Get("contractAddress")
  if contractAddress == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing contractAddress parameter")
    return
  }

  eventType := r.URL.Query().Get("eventType")
  if eventType == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing eventType parameter")
    return
  }

  // Pagination
  // TODO: Max limit
  defaultPage := 1
  defaultLimit := 10
  pageStr := r.URL.Query().Get("page")
  if pageStr == "" {
    pageStr = strconv.Itoa(defaultPage)
    return
  }
  page, err := strconv.Atoi(pageStr)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid page parameter")
    return
  }
  limitStr := r.URL.Query().Get("limit")
  if limitStr == "" {
    limitStr = strconv.Itoa(defaultLimit)
    return
  }
  limit, err := strconv.Atoi(limitStr)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid limit parameter")
    return
  }
  skip := (page - 1) * limit
  if skip < 0 {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid page parameter")
    return
  }
  findOptions := options.Find().SetSort(map[string]interface{}{
    "_id": -1,
  }).SetLimit(int64(limit)).SetSkip(int64(skip))

  // Filters
  userFilters := r.Body
  if userFilters == nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing filters parameter")
    return
  }
  defer userFilters.Close()

  // TODO: Allow empty filters
  // Parse filters ( example: {"key":"value","key2":"value2"} )
  var filters map[string]interface{}
  if err := json.NewDecoder(userFilters).Decode(&filters); err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid filters parameter")
    return
  }

  // Add contract address and event type to the filters
  filters["contract_address"] = contractAddress
  filters["event_type"] = eventType

  res, err := mongo.GetFocEngineEventsCollection().Find(r.Context(), filters, findOptions)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query events")
    return
  }
  defer res.Close(r.Context())

  var events []map[string]interface{}
  for res.Next(r.Context()) {
    var event map[string]interface{}
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
      return
    }
    events = append(events, event)
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over events")
    return
  }
  
  // Convert events to JSON
  eventsJson, err := json.Marshal(events)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal events to JSON")
    return
  }
  // Write the JSON response
  routeutils.WriteDataJson(w, string(eventsJson))
}

func GetUniqueOrdered(w http.ResponseWriter, r *http.Request) {
  contractAddress := r.URL.Query().Get("contractAddress")
  if contractAddress == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing contractAddress parameter")
    return
  }

  eventType := r.URL.Query().Get("eventType")
  if eventType == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing eventType parameter")
    return
  }

  // Pagination
  defaultPage := 1
  defaultLimit := 10
  pageStr := r.URL.Query().Get("page")
  if pageStr == "" {
    pageStr = strconv.Itoa(defaultPage)
  }
  page, err := strconv.Atoi(pageStr)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid page parameter")
    return
  }
  limitStr := r.URL.Query().Get("limit")
  if limitStr == "" {
    limitStr = strconv.Itoa(defaultLimit)
  }
  limit, err := strconv.Atoi(limitStr)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid limit parameter")
    return
  }
  skip := (page - 1) * limit
  if skip < 0 {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Invalid page parameter")
    return
  }

  uniqueKey := r.URL.Query().Get("uniqueKey")
  if uniqueKey == "" {
    routeutils.WriteErrorJson(w, http.StatusBadRequest, "Missing uniqueKey parameter")
    return
  }

  pipeline := []bson.M{
    {
      "$match": bson.M{
        "contract_address": contractAddress,
        "event_type":      eventType,
      },
    },
    {
      "$group": bson.M{
        "_id":   "$" + uniqueKey,
        "event": bson.M{
          "$last": "$$ROOT",
        },
      },
    },
    {
      "$sort": bson.M{
        "_id": -1,
      },
    },
    {
      "$skip": skip,
    },
    {
      "$limit": limit,
    },
    {
      "$replaceRoot": bson.M{
        "newRoot": "$event",
      },
    },
  }
  res, err := mongo.GetFocEngineEventsCollection().Aggregate(r.Context(), pipeline)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to query events")
    return
  }
  defer res.Close(r.Context())

  var events []map[string]interface{}
  for res.Next(r.Context()) {
    var event map[string]interface{}
    if err := res.Decode(&event); err != nil {
      routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to decode event")
      return
    }
    events = append(events, event)
  }
  if err := res.Err(); err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to iterate over events")
    return
  }
  if len(events) == 0 {
    routeutils.WriteErrorJson(w, http.StatusNotFound, "No events found for the specified contract address and event type")
    return
  }
  // Convert events to JSON
  eventsJson, err := json.Marshal(events)
  if err != nil {
    routeutils.WriteErrorJson(w, http.StatusInternalServerError, "Failed to marshal events to JSON")
    return
  }
  // Write the JSON response
  routeutils.WriteDataJson(w, string(eventsJson))
}
