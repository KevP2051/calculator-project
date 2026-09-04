# Calculator Frontend

This is the web app for the calculator project, built with React, TypeScript, and Vite.

You type two numbers, pick an operation, and it calls the API and shows the result. The Go microservice it talks to lives in [`../backend`](../backend/README.md).

For a complete full-stack guide, including the API reference and Docker setup, see the [project README](../README.md).

## Start from a fresh clone

Run these commands from a terminal:

```bash
git clone https://github.com/KevP2051/calculator-project.git
cd calculator-project/frontend
```

If the repository is already cloned, open a terminal in its root and run `cd frontend`.

## Prerequisites

**Node 22.12 or later.** Vite runs on Node 20.19, but Vitest needs 22.12, so on Node 20.19 the app works and the tests do not start.

```bash
node --version
```

## Run the app locally

### 1. Start the microservice

The frontend requires the API. In a separate terminal, from the repository root, run:

```bash
cd backend
go run ./cmd/server
```

Keep that terminal open. The API should be available at `http://localhost:8080`.

### 2. Create the frontend environment file

On macOS or Linux:

```bash
cp .env.example .env
```

On Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

### 3. Install dependencies and start Vite

```bash
npm ci
npm run dev
```

Open `http://localhost:5173` in a browser.

The microservice has to be running too, otherwise every calculation shows a "Cannot reach the server" message. See [`backend/README.md`](../backend/README.md).

### Configuration

One variable, in `.env`:

| Variable | Default | Purpose |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080/api/v1` | Where the microservice is |

Vite reads it when it builds, not when the page runs, so a production build has the URL baked into it. Changing that URL means building again.

## Run with Docker

Docker starts this frontend and the Go microservice together. From the repository root (the directory containing `docker-compose.yml`), run:

```bash
docker compose up --build
```

Open `http://localhost:5173`. To stop the containers, run:

```bash
docker compose down
```

## Scripts

| Script | What it does |
|---|---|
| `npm run dev` | Dev server with hot reload |
| `npm run build` | Type-checks and builds to `dist/` |
| `npm run preview` | Serves the built app |
| `npm run lint` | ESLint |
| `npm test` | Runs the tests once |
| `npm run test:watch` | Re-runs tests while you edit |
| `npm run test:coverage` | Tests plus a coverage report |

## Run the tests

```bash
npm ci
npm test
```

Coverage report:

```bash
npm run test:coverage
```

Prints a table and writes `coverage/index.html`. That report is committed, along with [`coverage/lcov.info`](coverage/lcov.info) for tooling, and rerunning the command overwrites both in place.

| File | Statements | Branches |
|---|---|---|
| **All files** | **96.55%** | **96.55%** |
| `actions/calculate.action.ts` | 100% | 100% |
| `api/calculator.api.ts` | 0% | 100% |
| `constants/error-messages.ts` | 100% | 100% |
| `constants/operations.ts` | 100% | 100% |
| `hooks/useCalculate.tsx` | 100% | 100% |
| `pages/CalculatorPage.tsx` | 100% | 94.44% |
| `schemas/calculation.schema.ts` | 100% | 100% |
| `utils/format-result.ts` | 100% | 100% |

There are three test files:

| File | What it covers |
|---|---|
| `schemas/calculation.schema.test.ts` | which inputs the form accepts and which it rejects |
| `pages/CalculatorPage.test.tsx` | rendering, typing, switching operation, the request it sends, validation errors, and the five microservice-failure cases |
| `utils/format-result.test.ts` | how results are formatted for display |

`api/calculator.api.ts` is at 0% because it is three lines that create the axios instance, and it is the module the tests replace with a mock.

## Project structure

```text
src/
├── main.tsx                 mounts the app
├── CalculatorApp.tsx        providers
├── calculator/
│   ├── api/                 axios instance
│   ├── actions/             the request function
│   ├── hooks/               useCalculate
│   ├── pages/               CalculatorPage
│   ├── schemas/             zod schema for the form
│   ├── constants/           operations list, error messages
│   ├── types/               request and response types
│   └── utils/               result formatting
├── components/ui/           shadcn components
├── lib/                     shadcn helper
└── test/                    test setup
```

The code is organized **by feature**. Everything the calculator needs lives in `src/calculator/`, so the whole feature can be read, moved or removed in one place. What is shared across features, like the shadcn components in `components/ui/`, stays outside it.

Inside the feature the folders are named after the role each file plays: the page renders the form, the hook handles the request, the action builds it, and the api sends it. Imports only go one way, from `pages` down to `api`.

## Design decisions

### Architecture

The code is grouped **by feature**, not by file type. The alternative is a top-level `hooks/`, `components/` and `utils/` with files from every feature mixed together. That works at this size, but once there are a few features, finding everything that belongs to one means opening several folders. Keeping the calculator in `src/calculator/` means the whole feature is one folder. What is genuinely shared, like the shadcn components in `components/ui/`, stays outside it.

Inside the feature the folders are named after the role each file plays: `pages` renders, `hooks` handles the request, `actions` builds it, `api` sends it. Imports go one way, from `pages` down to `api`. That is also what made the tests simple to write, since replacing `api` with a mock leaves everything above it running normally.

There is one page and one request. The page holds the form, `useCalculate` handles the request, `calculateAction` builds the payload, and the axios instance sends it.

When the microservice returns an error it sends a code, like `DIVISION_BY_ZERO`, along with a message. The frontend only reads the code and looks up its own text for it. Its message is for logs, so wording stays a frontend decision.

The list of operations lives in one place, `constants/operations.ts`. It holds the value, symbol, label and how many numbers each operation takes. The dropdown, the number of inputs shown, and the schema all read from it.

### Dependencies

- **Tailwind CSS + shadcn/ui:** Tailwind CSS provides utility-based styling and responsive layouts, while shadcn/ui provides accessible and reusable UI primitives that are easy to customize without introducing a heavy component library.

- **React Hook Form:** React Hook Form manages form state and submission efficiently, with minimal re-renders and straightforward integration with controlled and uncontrolled form components.

- **Zod:** Zod provides declarative and centralized schema validation for form inputs. It also integrates with React Hook Form and TypeScript, allowing the validation schema to serve as a single source of truth for the expected form data.

- **TanStack Query:** TanStack Query manages the asynchronous calculation request and its loading, success, and error states. While local React state could be sufficient for this small application, TanStack Query provides a consistent abstraction for server interactions and keeps API state management separate from the UI.

- **Axios:** Axios was chosen as the HTTP client to centralize API configuration and provide a consistent interface for requests and error handling.

- **Sonner:** Sonner shows errors as toasts, so they appear and go away instead of sitting in the page.

### Number formatting

The microservice returns results exactly as computed, so `0.1 + 0.2` comes back as `0.30000000000000004`. Formatting them is the frontend's job.

`formatResult` rounds to 15 significant digits, which hides that noise without changing numbers that are already exact. Very large or very small results switch to scientific notation, because writing `1e305` in full would be 306 digits.

### Testing

The tests mock the axios instance and the toast library, nothing else. The form, the schema, the hook and the error handling all run for real, so the tests follow the same path the app does.

They check what the user would see, not CSS classes, so the styling can change without breaking them.

### Docker image

The image is built in two stages. Node installs the dependencies and runs `npm run build`, then only `dist/` is copied into an Nginx image and the whole Node stage is discarded.

Nginx is there because `npm run dev` and `npm run preview` are development tools, not production servers. What `npm run build` produces is plain HTML, CSS and JavaScript, so nothing has to run the app, something just has to hand those files to the browser. Nginx does that, and the image that ships carries no Node, no `node_modules` and no source.

Nginx only serves files. It does not proxy `/api` through to the microservice, so the browser calls `http://localhost:8080` directly, exactly as it does when running locally with `npm run dev`. That keeps one request path instead of two, and it is why `VITE_API_URL` is baked in at build time and the microservice sets a CORS origin.

## Assumptions

- The calculator accepts up to two numbers, since none of the supported operations requires more inputs, so the form never displays a third input.
- All user-facing text is hardcoded in English; internationalization is not implemented.
- There is no calculation history. Each new calculation replaces the previous result, and calculations are not persisted.
- If a calculation fails, the result box returns to its `--` placeholder and the error is shown as a toast, because TanStack Query clears the mutation data when a new one starts. Keeping the previous result visible next to the error would be a product decision and is outside the scope of this implementation.
- Request cancellation is not implemented, so submitting multiple calculations in quick succession may send multiple requests. This is acceptable for the expected usage and scope of the application.
