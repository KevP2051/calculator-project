package api

import (
	"encoding/json"
	"net/http"

	"calculator/backend/internal/service"
)

type calculateRequest struct {
	Operation string `json:"operation"`
	Operands  []any  `json:"operands"`
}

type calculateResponse struct {
	Result float64 `json:"result"`
}

type errorDetail struct {
	Code    service.Code `json:"code"`
	Message string       `json:"message"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	var request calculateRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, &service.Error{
			Code:    service.CodeMalformedJSON,
			Message: "request body is not a valid JSON object",
		})
		return
	}

	result, failure := service.Calculate(service.Request{
		Operation: request.Operation,
		Operands:  request.Operands,
	})
	if failure != nil {
		writeError(w, failure)
		return
	}

	writeJSON(w, http.StatusOK, calculateResponse{Result: result})
}

func writeError(w http.ResponseWriter, failure *service.Error) {
	status, ok := statusByCode[failure.Code]
	if !ok {
		status = http.StatusInternalServerError
	}

	writeJSON(w, status, errorResponse{
		Error: errorDetail{Code: failure.Code, Message: failure.Message},
	})
}
