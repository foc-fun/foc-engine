package routeutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// ReadJsonBody reads the body of an http.Request and unmarshals it into a struct.
//
//	Generic Param:
//	  bodyType: The type of the struct to unmarshal the body into.
//	Parameters:
//	  r: The http.Request to read the body from.
//	Returns:
//	  *bodyType: A pointer to the unmarshaled body.
func ReadJsonBody[bodyType any](r *http.Request) (*bodyType, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var body bodyType
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func PostJson(url string, body interface{}, headers map[string]string) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, io.NopCloser(bytes.NewBuffer(jsonBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	return client.Do(req)
}
