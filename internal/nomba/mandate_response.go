package nomba

import (
	"encoding/json"
	"net/http"
	"strings"
)

type nombaErrorBody struct {
	Code            string   `json:"code"`
	Description     string   `json:"description"`
	Message         string   `json:"message"`
	ResponseCode    string   `json:"responseCode"`
	ResponseMessage string   `json:"responseMessage"`
	Errors          []string `json:"errors"`
}

// HTTPErrorFromNombaBody builds an HTTPError from a Nomba JSON body (any envelope).
func HTTPErrorFromNombaBody(statusCode int, raw []byte) *HTTPError {
	e := &HTTPError{
		StatusCode: statusCode,
		RawBody:    strings.TrimSpace(string(raw)),
	}
	var body nombaErrorBody
	if err := json.Unmarshal(raw, &body); err != nil {
		if e.RawBody != "" {
			e.Description = e.RawBody
		}
		return e
	}
	e.Code = firstNonEmpty(body.Code, body.ResponseCode)
	e.Description = firstNonEmpty(body.Description, body.ResponseMessage, body.Message)
	e.Errors = body.Errors
	return e
}

func (e *HTTPError) userMessage() string {
	if len(e.Errors) > 0 {
		return strings.Join(e.Errors, "; ")
	}
	if e.Description != "" {
		return e.Description
	}
	if e.Code != "" {
		return e.Code
	}
	if e.RawBody != "" {
		return e.RawBody
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseMandateCreateResponse(raw json.RawMessage) (CreateMandateResult, error) {
	var zero CreateMandateResult
	if len(raw) == 0 {
		return zero, HTTPErrorFromNombaBody(http.StatusOK, raw)
	}

	var std APIResponse[CreateMandateResult]
	if err := json.Unmarshal(raw, &std); err == nil && std.Code != "" {
		if std.OK() {
			return std.Data, nil
		}
	}

	var legacy LegacyMandateAPIResponse
	if err := json.Unmarshal(raw, &legacy); err == nil && legacy.ResponseCode != "" {
		if legacy.OK() {
			return legacy.Data, nil
		}
	}

	var status StatusAPIResponse[CreateMandateResult]
	if err := json.Unmarshal(raw, &status); err == nil && status.Code != "" {
		if status.OK() {
			return status.Data, nil
		}
	}

	return zero, HTTPErrorFromNombaBody(http.StatusOK, raw)
}

func parseMandateStatusResponse(raw json.RawMessage) (MandateStatusResult, error) {
	var zero MandateStatusResult
	if len(raw) == 0 {
		return zero, HTTPErrorFromNombaBody(http.StatusOK, raw)
	}

	var std APIResponse[MandateStatusResult]
	if err := json.Unmarshal(raw, &std); err == nil && std.Code != "" {
		if std.OK() {
			return std.Data, nil
		}
	}

	var status StatusAPIResponse[MandateStatusResult]
	if err := json.Unmarshal(raw, &status); err == nil && status.Code != "" {
		if status.OK() {
			return status.Data, nil
		}
	}

	return zero, HTTPErrorFromNombaBody(http.StatusOK, raw)
}
