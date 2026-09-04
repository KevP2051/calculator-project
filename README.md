# Full-Stack Calculator

A calculator application built for the Sezzle technical assessment. It pairs a React + TypeScript interface with a small Go REST API, supporting addition, subtraction, multiplication, division, exponentiation, square root, and percentage calculations.

The project prioritizes a clear API contract, input validation, focused tests, and a setup that can run either locally or with Docker.

## Contents

- [Architecture](#architecture)
- [Requirements](#requirements)
- [Run locally](#run-locally)
- [Run with Docker](#run-with-docker)
- [Tests and coverage](#tests-and-coverage)
- [API examples](#api-examples)
- [Design decisions](#design-decisions)
- [Assumptions and limitations](#assumptions-and-limitations)
- [Project documentation](#project-documentation)

## Architecture

```text
Browser
  │
  ▼
React + TypeScript frontend ──HTTP/JSON──► Go REST API
http://localhost:5173                       http://localhost:8080
```

The frontend sends each calculation to `POST /api/v1/calculate`. The API validates the request, performs the calculation, and returns a JSON result or a predictable JSON error.

```text
calculator-project/
├── frontend/                 React, TypeScript, Vite, Tailwind, and tests
├── backend/                  Go API, arithmetic domain, and tests
├── specs/                    API specification, research, plan, and task list
├── docker-compose.yml        Runs both services together
└── PROMPTS.md                Prompts used during development
```

## Requirements

Choose one of the following ways to run the project:

- **Local development:** Git, Go 1.22 or later, Node.js 22.12 or later, and npm.
- **Docker:** Git and Docker Desktop (or Docker Engine with the Compose plugin).

Check the local runtimes, if using local development:

```bash
git --version
go version
node --version
npm --version
```

## Run locally

### 1. Clone the repository

```bash
git clone https://github.com/KevP2051/calculator-project.git
cd calculator-project
```

### 2. Start the backend

Open a terminal in the repository root and run:

```bash
cd backend
go run ./cmd/server
```

The API listens on `http://localhost:8080`. Keep this terminal open.

Optionally, confirm that it is healthy in another terminal:

```bash
curl http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

### 3. Configure and start the frontend

Open a second terminal in the repository root.

On macOS or Linux:

```bash
cd frontend
cp .env.example .env
npm ci
npm run dev
```

On Windows PowerShell:

```powershell
cd frontend
Copy-Item .env.example .env
npm ci
npm run dev
```

Open [http://localhost:5173](http://localhost:5173). The frontend reads `VITE_API_URL` from `frontend/.env`; its default points to the local backend.

## Run with Docker

Docker runs the frontend as a production static site and the backend as a Go service. No Node.js or Go installation is needed on the host.

### 1. Clone the repository

```bash
git clone https://github.com/KevP2051/calculator-project.git
cd calculator-project
```

### 2. Build and start both services

```bash
docker compose up --build
```

Then open [http://localhost:5173](http://localhost:5173). The API health check is available at [http://localhost:8080/healthz](http://localhost:8080/healthz).

To run the services in the background instead:

```bash
docker compose up --build -d
```

To stop and remove the containers:

```bash
docker compose down
```

## Tests and coverage

Run tests from their respective directories.

### Backend

```bash
cd backend
go test ./... -cover
```

To generate an HTML coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Frontend

```bash
cd frontend
npm ci
npm test
```

To generate frontend coverage:

```bash
npm run test:coverage
```

The report is written to `frontend/coverage/index.html`.

## API examples

The calculation endpoint is `POST http://localhost:8080/api/v1/calculate`.

### Successful calculation

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"multiply","operands":[6,7]}'
```

Response:

```json
{"result":42}
```

### Validation error

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"divide","operands":[1,0]}'
```

Response (HTTP 422):

```json
{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero"}}
```

PowerShell users can replace `curl` with `curl.exe`, or use the API through the frontend.

See [backend/README.md](backend/README.md) for the complete operation and error-code reference.

## Design decisions

### Frontend rationale

- **Feature-based organization:** files that implement the calculator live in `src/calculator/`; shared primitives, such as the shadcn/ui components, remain outside it. This is more cohesive than grouping every hook, component, and utility by file type across the whole application, and it keeps the feature easy to read, move, or extend.
- **Clear, one-way request flow:** `CalculatorPage` renders the form, `useCalculate` manages the request lifecycle, `calculateAction` builds the payload, and the Axios instance sends it. Dependencies flow from `pages` down to `api`, which gives each layer one responsibility and makes individual boundaries easier to test.
- **One source of truth for operations:** `constants/operations.ts` defines an operation's value, symbol, label, and operand count. The selector, visible inputs, and validation schema consume that same definition, so adding an operation does not require duplicating its rules in several places.
- **Client validation and server state are separate:** React Hook Form manages input and submission state, while Zod declares the form rules. TanStack Query manages loading, success, and failure of the server request. This separation avoids turning a small page into a single component that owns unrelated concerns.
- **Stable error experience:** the API returns a stable error code, such as `DIVISION_BY_ZERO`; the frontend maps that code to its own user-facing message. This keeps API diagnostics separate from UI copy and prevents the client from relying on backend wording.
- **Deliberate UI dependencies:** Tailwind CSS provides responsive utility styling, shadcn/ui supplies accessible primitives that remain easy to customize, and Sonner presents transient error feedback without permanently occupying the page.
- **Presentation-level number formatting:** the backend preserves the value it calculated, while `formatResult` rounds only for display to 15 significant digits and uses scientific notation for extreme values. This hides familiar floating-point noise without changing the API contract.
- **Behavior-focused tests:** frontend tests mock only the Axios instance. The form, schema, hook, rendering, and error handling run together, so tests follow the user-visible path rather than asserting implementation details such as CSS classes.

### Backend rationale

- **Spec-driven development:** the backend's requirements, clarification work, research, plan, and task breakdown were created with GitHub Spec Kit before implementation. This process is documented in the [backend design decisions](backend/README.md#why-spec-driven-development) and the [`specs/001-calculator-rest-api/`](specs/001-calculator-rest-api/) directory.
- **Topology sized to the problem:** one stateless Go microservice serves the entire backend contract, and the React client calls it directly over HTTP. A gateway, BFF, broker, or additional services would introduce operational and coordination costs without solving a problem present in this assessment.
- **Layered, inward dependencies:** `api → service → calc` separates HTTP/JSON concerns, validation and orchestration, and arithmetic rules. The `calc` package imports neither HTTP nor JSON code, so mathematical behavior can be tested without a server and the API boundary stays small.
- **A compact, API-first contract:** every operation uses `POST /api/v1/calculate`, with an `operands` array rather than separate named fields. This keeps validation and error handling consistent, makes operation arity explicit, and turns a new operation into a registry change instead of another endpoint.
- **Predictable errors:** calculation failures use a single JSON envelope with a stable machine-readable `code` and a human-readable `message`. Validation occurs in a fixed sequence—request structure, operation, operand count, operand values, calculation—so callers receive deterministic errors. `404` and `405` remain ordinary router responses rather than calculation errors.
- **Explicit numerical behavior:** Go `float64` supplies a well-defined IEEE 754 binary64 range and precision model, without an arbitrary digit limit. The API does not apply presentation rounding; it explicitly guards division by zero, overflow, underflow, invalid real-number domains, and undefined results. `json.Decoder.UseNumber()` preserves the distinction between invalid JSON numbers and numbers that cannot be represented by `float64`.
- **Intentional operation rules:** percentage is defined as `(a / 100) * b`; operand order matters for subtraction, division, exponentiation, and percentage; negative bases with fractional exponents are rejected because the service is limited to real-number results.
- **Small dependency surface and focused tests:** the service relies only on Go's standard library, including `net/http` and `encoding/json`, and adds CORS as small middleware. Table-driven tests concentrate on arithmetic and validation, where the calculation and input-handling risks live; exact comparisons are used except where an irrational value requires a documented tolerance.

### Local delivery with Docker

- **One-command full-stack run:** Docker Compose starts the API and frontend together, matching the same browser-to-API relationship used locally.
- **Multi-stage images:** the backend image compiles a small Go binary, while the frontend image builds the Vite application and serves static assets through Nginx. This keeps runtime images free of the respective build toolchains.
- **Explicit browser configuration:** `VITE_API_URL` is supplied at frontend build time and CORS is configured for `http://localhost:5173`, making the local browser-to-API connection visible rather than implicit.

## Assumptions

- Calculations use Go `float64` (IEEE 754 binary64). Results are not rounded by the API; display formatting belongs to the frontend.
- Each supported operation has a fixed arity of one or two operands. The interface shows at most two inputs, and square root is the only one-operand operation. This is intentionally scoped as an operation-based calculator for the assessment, not an expression calculator that evaluates arbitrary formulas or chains calculations across multiple operands.
- The system uses one backend microservice. Given the application's size and the absence of cross-service concerns, the React client calls that service directly over HTTP. An API gateway, backend-for-frontend layer, message broker, and additional services would add complexity without a corresponding benefit for this scope.
- The calculator operates only on real numbers. For example, square roots of negative values and negative bases raised to fractional exponents are rejected.
- Each request performs one calculation. Expression parsing, calculation history, authentication, persistence, and rate limiting are outside the assessment scope.
- The Docker setup is for local full-stack execution; it is not a production deployment configuration.

## Project documentation

- [Frontend guide](frontend/README.md#calculator-frontend): setup, scripts, tests, and frontend rationale.
- [Backend guide](backend/README.md#calculator-backend): setup, operations, API contract, and backend rationale.
- [API specification](specs/001-calculator-rest-api/contracts/openapi.yaml): endpoint contract.
- [Development specification](specs/001-calculator-rest-api/spec.md): functional behavior and acceptance criteria.
- [Prompts used during development](PROMPTS.md).
