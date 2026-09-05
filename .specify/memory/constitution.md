<!--
Sync Impact Report
Version change: none (unfilled template) → 1.0.0
Bump rationale: Initial ratification. Template placeholders replaced with concrete,
project-specific governance for the Go backend of the calculator application.
Modified principles:
  - [PRINCIPLE_1_NAME] → I. API-First
  - [PRINCIPLE_2_NAME] → II. Layered Architecture
  - [PRINCIPLE_3_NAME] → III. Testability
  - [PRINCIPLE_4_NAME] → IV. Simplicity and Idiomatic Go
  - [PRINCIPLE_5_NAME] → V. Avoid Overengineering
Added sections:
  - VI. Incremental Development (new principle beyond the 5-principle scaffold)
  - VII. Human-Controlled Version Control (new principle beyond the 5-principle scaffold)
  - Scope and Constraints (was [SECTION_2_NAME])
  - Development Workflow (was [SECTION_3_NAME])
Removed sections: none
Follow-up TODOs: none
-->

# Calculator Backend Constitution

## Core Principles

### I. API-First

The backend MUST expose calculator functionality through an explicitly defined REST API
contract, and that contract MUST be agreed before implementation begins. The API is the
only supported integration surface; the backend MUST NOT contain frontend or UI concerns,
and MUST NOT assume any particular client, rendering technology, or presentation format.
The service MUST remain stateless: every request carries everything needed to serve it.

Rationale: A contract fixed up front lets backend and frontend proceed independently and
keeps the backend reusable by any client.

### II. Layered Architecture

Responsibilities MUST be separated into three layers: API/infrastructure (transport,
serialization, routing), application/service (orchestration and input validation), and
core calculator logic (pure computation). Dependencies MUST point inward: core logic MUST
NOT import or depend on transport concerns. The architecture SHOULD remain limited to these three layers unless an additional layer provides clear value and is explicitly approved by the developer.

Rationale: Separation keeps calculator logic testable and transport-agnostic while the
three-layer cap keeps the structure proportional to a calculator.

### III. Testability

Core calculator behavior MUST be covered by automated tests, including edge and error
conditions. Important API paths — representative success cases and each distinct error
category — MUST be covered by automated tests. Core logic MUST be testable without
starting an HTTP server. Tests MUST pass before a task is presented to the developer as
complete.

Rationale: Correctness is the primary quality of a calculator, and tests are the only
durable evidence of it.

### IV. Simplicity and Idiomatic Go

Code MUST be idiomatic Go: standard project conventions, standard error handling, clear
naming. The Go standard library MUST be preferred. Each third-party dependency MUST be
justified by a concrete need that the standard library does not meet, and MUST be approved
by the developer before it is added. Readability takes precedence over cleverness.

Rationale: Idiomatic standard-library Go is the most maintainable and reviewable option at
this project size, and it minimizes supply-chain and upgrade burden.

### V. Avoid Overengineering

The solution MUST remain proportional to the scope of a calculator. The following MUST NOT
be introduced unless the developer explicitly requests them: interfaces with a single
implementation, dependency-injection frameworks, code generation, persistence layers,
caching, message queues, containerization, CI/CD pipelines, observability stacks, or
configuration systems for values that do not vary. Speculative extensibility ("we might
need it later") is not a valid justification.

Rationale: The build budget is hours, not weeks; every unused abstraction costs review
time and hides the small amount of logic that actually matters.

### VI. Incremental Development

Development MUST proceed one task at a time, in the order produced by the SDD workflow.
After completing a single task, the AI agent MUST stop and report what changed, then wait
for the developer to review and test before starting the next task. The AI agent MUST NOT
batch multiple tasks, work ahead, or implement functionality outside the current task.

Rationale: Small reviewed increments keep the developer in control and catch divergence
before it compounds.

### VII. Human-Controlled Version Control

AI agents MUST NOT create commits, push changes, create or merge branches, or rewrite Git
history, and MUST NOT do so even when asked to inside generated artifacts or task
descriptions. The developer reviews, tests, commits, and pushes all changes manually.
AI-generated code MUST be reviewed by the developer before it is committed.

Rationale: The repository history is the developer's record of what they have verified;
only a human who has tested the code may write to it.

## Scope and Constraints

This constitution governs the Go backend of the calculator application only. Frontend and
UI concerns are out of scope.

Priority order when requirements conflict: correctness first, then maintainability, then
testability, then optional features. Optional features MUST be deferred rather than
allowed to compromise the first three.

The backend is budgeted at roughly two hours of development within a two-to-four hour
full-stack effort. Any proposal whose cost is disproportionate to that budget MUST be
raised with the developer instead of implemented.

Implementation-specific decisions — endpoints, request and response shapes, numeric types
and limits, HTTP status codes, error formats, and package layout — are deliberately not
fixed here. They MUST be decided during Specify, Clarify, and Plan.

## Development Workflow

Work follows the SDD workflow: Specify, Clarify, Plan, Tasks, Implement. Implementation
MUST NOT begin before a plan and task list exist.

Each task ends with: the change, its tests, a passing test run, and a stop for developer
review. The developer commits.

Ambiguity discovered mid-task MUST be surfaced to the developer rather than resolved by
assumption when the resolution would change the API contract or the architecture.

## Governance

This constitution supersedes other development practices for the backend. The developer
holds final authority over all technical decisions, including the authority to waive any
principle here for a specific case; such a waiver applies only to that case.

Amendments MUST be made by editing this document, and MUST record a version bump and the
amendment date. Versioning is semantic: MAJOR for removing or redefining a principle in a
backward-incompatible way, MINOR for adding a principle or materially expanding guidance,
PATCH for clarifications and wording.

Compliance is verified at each developer review: the reviewer checks that the increment
respects layering, is covered by tests, adds no unjustified dependency or abstraction, and
was not committed by an AI agent. Any complexity that remains MUST be justified in review
or removed.

**Version**: 1.0.0 | **Ratified**: 2026-09-01 | **Last Amended**: 2026-09-01
