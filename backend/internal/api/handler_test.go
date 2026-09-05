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

func envelopeShape(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}

	if _, hasResult := body["result"]; hasResult {
		t.Errorf("error response carries a result: %s", raw)
	}
	if len(body) != 1 {
		t.Errorf("error response has %d top-level keys, want 1: %s", len(body), raw)
	}

	detail, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("error response has no error object: %s", raw)
	}
	if len(detail) != 2 {
		t.Errorf("error object has %d keys, want 2: %s", len(detail), raw)
	}
	if _, ok := detail["code"].(string); !ok {
		t.Errorf("error object has no string code: %s", raw)
	}
	if message, ok := detail["message"].(string); !ok || message == "" {
		t.Errorf("error object has no non-empty message: %s", raw)
	}

	return detail
}

func TestErrorResponsesUseOneEnvelopeAndTheRightStatus(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   service.Code
	}{
		{"unterminated body", `{"operation":`, http.StatusBadRequest, service.CodeMalformedJSON},
		{"body is not an object", `[1,2]`, http.StatusBadRequest, service.CodeMalformedJSON},
		{"empty body", ``, http.StatusBadRequest, service.CodeMalformedJSON},
		{"operation absent", `{"operands":[1,2]}`, http.StatusBadRequest, service.CodeMissingField},
		{"operands absent", `{"operation":"add"}`, http.StatusBadRequest, service.CodeMissingField},
		{"unknown operation", `{"operation":"tangent","operands":[1,2]}`, http.StatusBadRequest, service.CodeUnsupportedOperation},
		{"too few operands", `{"operation":"add","operands":[1]}`, http.StatusBadRequest, service.CodeInvalidOperandCount},
		{"too many operands", `{"operation":"add","operands":[1,2,3]}`, http.StatusBadRequest, service.CodeInvalidOperandCount},
		{"operand is text", `{"operation":"add","operands":[1,"abc"]}`, http.StatusBadRequest, service.CodeInvalidOperand},
		{"operand is the word NaN", `{"operation":"add","operands":[1,"NaN"]}`, http.StatusBadRequest, service.CodeInvalidOperand},
		{"operand is the word Infinity", `{"operation":"add","operands":[1,"Infinity"]}`, http.StatusBadRequest, service.CodeInvalidOperand},
		{"operand is a boolean", `{"operation":"add","operands":[1,true]}`, http.StatusBadRequest, service.CodeInvalidOperand},
		{"operand is null", `{"operation":"add","operands":[1,null]}`, http.StatusBadRequest, service.CodeInvalidOperand},
		{"operand too large", `{"operation":"add","operands":[1e400,1]}`, http.StatusBadRequest, service.CodeOperandOutOfRange},
		{"divisor is zero", `{"operation":"divide","operands":[1,0]}`, http.StatusUnprocessableEntity, service.CodeDivisionByZero},
		{"result overflows", `{"operation":"multiply","operands":[1e308,10]}`, http.StatusUnprocessableEntity, service.CodeResultOverflow},
		{"result underflows", `{"operation":"multiply","operands":[1e-200,1e-200]}`, http.StatusUnprocessableEntity, service.CodeResultUnderflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := post(t, tt.body)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			detail := envelopeShape(t, recorder.Body.Bytes())
			if detail["code"] != string(tt.wantCode) {
				t.Errorf("code = %v, want %q", detail["code"], tt.wantCode)
			}
		})
	}
}

func TestEveryCodeSerializesIntoTheSameEnvelope(t *testing.T) {
	for _, code := range service.AllCodes {
		t.Run(string(code), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeError(recorder, &service.Error{Code: code, Message: "failure detail"})

			status, mapped := statusByCode[code]
			if !mapped {
				t.Fatalf("code %q has no status mapping", code)
			}
			if recorder.Code != status {
				t.Errorf("status = %d, want %d", recorder.Code, status)
			}

			detail := envelopeShape(t, recorder.Body.Bytes())
			if detail["code"] != string(code) {
				t.Errorf("code = %v, want %q", detail["code"], code)
			}
		})
	}
}

func TestErrorResponsesNeverContainNonFiniteTokens(t *testing.T) {
	bodies := []string{
		`{"operation":"multiply","operands":[1e308,10]}`,
		`{"operation":"multiply","operands":[1e-200,1e-200]}`,
		`{"operation":"divide","operands":[1,0]}`,
		`{"operation":"add","operands":[1e400,1]}`,
	}

	for _, body := range bodies {
		t.Run(body, func(t *testing.T) {
			response := post(t, body).Body.String()
			for _, token := range []string{"Infinity", "-Infinity", "NaN", "Inf"} {
				if strings.Contains(response, token) {
					t.Errorf("response contains %q: %s", token, response)
				}
			}
		})
	}
}

func TestSubnormalSubtractionSucceedsOverHTTP(t *testing.T) {
	tests := []struct {
		name string
		body string
		want float64
	}{
		{"equal subnormals cancel", `{"operation":"subtract","operands":[5e-324,5e-324]}`, 0},
		{"adjacent subnormals", `{"operation":"subtract","operands":[1.5e-323,1e-323]}`, 5e-324},
		{"exact cancellation", `{"operation":"subtract","operands":[5,5]}`, 0},
		{"zero operand in multiplication", `{"operation":"multiply","operands":[0,5]}`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeResult(t, post(t, tt.body)); got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuccessfulResponsesNeverContainNonFiniteTokens(t *testing.T) {
	bodies := []string{
		`{"operation":"multiply","operands":[1e308,1]}`,
		`{"operation":"divide","operands":[1e308,10]}`,
		`{"operation":"add","operands":[1.7976931348623157e308,0]}`,
		`{"operation":"subtract","operands":[5e-324,5e-324]}`,
	}

	for _, body := range bodies {
		t.Run(body, func(t *testing.T) {
			recorder := post(t, body)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d, body %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			for _, token := range []string{"Infinity", "-Infinity", "NaN", "Inf", "null"} {
				if strings.Contains(recorder.Body.String(), token) {
					t.Errorf("response contains %q: %s", token, recorder.Body.String())
				}
			}
		})
	}
}
