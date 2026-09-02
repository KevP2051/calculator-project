package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calculator/backend/internal/service"
)

func post(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	NewRouter(testOrigin).ServeHTTP(recorder, request)

	return recorder
}

func decodeResult(t *testing.T, recorder *httptest.ResponseRecorder) float64 {
	t.Helper()

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Result float64 `json:"result"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}

	return response.Result
}

func TestCalculateReturnsResults(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{"addition", `{"operation":"add","operands":[2,3]}`, 5},
		{"subtraction", `{"operation":"subtract","operands":[10,4]}`, 6},
		{"multiplication", `{"operation":"multiply","operands":[6,7]}`, 42},
		{"division", `{"operation":"divide","operands":[10,4]}`, 2.5},
		{"operand order", `{"operation":"subtract","operands":[4,10]}`, -6},
		{"decimals", `{"operation":"add","operands":[2.5,0.25]}`, 2.75},
		{"negatives", `{"operation":"multiply","operands":[-6,7]}`, -42},
		{"exact cancellation", `{"operation":"subtract","operands":[5,5]}`, 0},
		{"extra fields are ignored", `{"operation":"add","operands":[2,3],"precision":4,"nonsense":true}`, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeResult(t, post(t, tt.body)); got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseBodyKeepsRepresentationNoise(t *testing.T) {
	recorder := post(t, `{"operation":"add","operands":[0.1,0.2]}`)

	body := strings.TrimSpace(recorder.Body.String())
	if body != `{"result":0.30000000000000004}` {
		t.Errorf("body = %s, want {\"result\":0.30000000000000004}", body)
	}
}

func TestIdenticalRequestsProduceIdenticalResponses(t *testing.T) {
	const body = `{"operation":"divide","operands":[10,3]}`

	first := post(t, body).Body.String()
	second := post(t, body).Body.String()

	if first != second {
		t.Errorf("responses differ:\n%s\n%s", first, second)
	}
}

func TestSuccessResponsesAreJSON(t *testing.T) {
	recorder := post(t, `{"operation":"add","operands":[2,3]}`)

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestMalformedBodyIsRejected(t *testing.T) {
	recorder := post(t, `{"operation":`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if response.Error.Code != service.CodeMalformedJSON {
		t.Errorf("code = %q, want %q", response.Error.Code, service.CodeMalformedJSON)
	}
}

func TestCalculationFailureIsReported(t *testing.T) {
	recorder := post(t, `{"operation":"divide","operands":[1,0]}`)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}

	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if response.Error.Code != service.CodeDivisionByZero {
		t.Errorf("code = %q, want %q", response.Error.Code, service.CodeDivisionByZero)
	}
	if strings.Contains(recorder.Body.String(), "result") {
		t.Errorf("error response carries a result: %s", recorder.Body.String())
	}
}
