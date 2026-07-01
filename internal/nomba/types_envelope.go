package nomba

// Standard Nomba API envelope: { "code", "description", "data" }.
// Success responses use code "00".

type APIResponse[T any] struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        T      `json:"data"`
}

func (r APIResponse[T]) OK() bool {
	return r.Code == ResponseCodeSuccess
}

// APIError is the error body returned on 4xx/5xx responses.
type APIError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e APIError) Error() string {
	if e.Description != "" {
		return e.Description
	}
	return "nomba api error: " + e.Code
}

// Legacy mandate create response uses responseCode/responseMessage instead of code/description.
type LegacyMandateAPIResponse struct {
	ResponseCode    string              `json:"responseCode"`
	ResponseMessage string              `json:"responseMessage"`
	Data            CreateMandateResult `json:"data"`
}

func (r LegacyMandateAPIResponse) OK() bool {
	return r.ResponseCode == ResponseCodeSuccess
}

// StatusAPIResponse is used by some direct-debit endpoints.
type StatusAPIResponse[T any] struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Data        T      `json:"data"`
	Message     string `json:"message"`
	Status      bool   `json:"status"`
}

func (r StatusAPIResponse[T]) OK() bool {
	return r.Code == ResponseCodeSuccess
}
