# AI DevKit Rules

## Project Context
This project uses ai-devkit for structured AI-assisted development. Phase documentation is located in `docs/ai/`.
The runtime architecture is a Go monolith wired in `cmd/main.go`, with adapters in `internal/infra/` (PostgreSQL, Evolution API, Gemini, HTTP) and business logic in `internal/usecase/` + `internal/domain/`.

## Documentation Structure
- `docs/ai/requirements/` - Problem understanding and requirements
- `docs/ai/design/` - System architecture and design decisions (include mermaid diagrams)
- `docs/ai/planning/` - Task breakdown and project planning
- `docs/ai/implementation/` - Implementation guides and notes
- `docs/ai/testing/` - Testing strategy and test cases
- `docs/ai/deployment/` - Deployment and infrastructure docs
- `docs/ai/monitoring/` - Monitoring and observability setup

## Code Style & Standards
- Follow the project's established code style and conventions
- Write clear, self-documenting code with meaningful variable names
- Add comments for complex logic or non-obvious decisions
- Preserve layer boundaries: keep contracts in `internal/domain/ports/`, business rules in `internal/usecase/`, and provider-specific logic in `internal/infra/`.
- Keep user-facing texts and test assertions aligned with existing Portuguese phrasing used across handlers/use cases (e.g., `internal/infra/http/`).

## Development Workflow
- Review phase documentation in `docs/ai/` before implementing features
- Keep requirements, design, and implementation docs updated as the project evolves
- Reference the planning doc for task breakdown and priorities
- Copy the testing template (`docs/ai/testing/README.md`) before creating feature-specific testing docs
- Start from `.env.example` for local setup; required runtime envs are validated in `internal/config/config.go`.
- Prefer `make run` for day-to-day development (starts `postgres` + `redis` via Docker, runs app locally with `go run ./cmd/main.go`).
- Migrations are executed automatically on startup in `cmd/main.go` via `db.RunMigrations(...)`; use `make migrate*` only for manual control.
- Note: in `docker-compose.yml`, the `app` service is currently commented out; `make compose-up` is infra-oriented unless that service is enabled.

## AI Interaction Guidelines
- When implementing features, first check relevant phase documentation
- For new features, start with requirements clarification
- Update phase docs when significant changes or decisions are made

## Skills (Extend Your Capabilities)
Skills are packaged capabilities that teach you new competencies, patterns, and best practices. Check for installed skills in the project's skill directory and use them to enhance your work.

### Using Installed Skills
1. **Check for skills**: Look for `SKILL.md` files in the project's skill directory
2. **Read skill instructions**: Each skill contains detailed guidance on when and how to use it
3. **Apply skill knowledge**: Follow the patterns, commands, and best practices defined in the skill

### Key Installed Skills
- No `SKILL.md` files are currently present in this repository. If skills are added later, follow their instructions before implementation.

### When to Reference Skills
- Before implementing features that match a skill's domain
- When MCP tools are unavailable but skill provides CLI alternatives
- To follow established patterns and conventions defined in skills

## Knowledge Memory (Always Use When Helpful)
The AI assistant should proactively use knowledge memory throughout all interactions.

> **Tip**: If MCP is unavailable, use `npx ai-devkit memory ...` CLI commands directly.

### When to Search Memory
- Before starting any task, search for relevant project conventions, patterns, or decisions
- When you need clarification on how something was done before
- To check for existing solutions to similar problems
- To understand project-specific terminology or standards

**How to search**:
- Use `memory.searchKnowledge` MCP tool with relevant keywords, tags, and scope
- If MCP tools are unavailable, use `npx ai-devkit memory search` CLI command (see memory skill for details)
- Example: Search for "authentication patterns" when implementing auth features

### When to Store Memory
- After making important architectural or design decisions
- When discovering useful patterns or solutions worth reusing
- If the user explicitly asks to "remember this" or save guidance
- When you establish new conventions or standards for the project

**How to store**:
- Use `memory.storeKnowledge` MCP tool
- If MCP tools are unavailable, use `npx ai-devkit memory store` CLI command (see memory skill for details)
- Include clear title, detailed content, relevant tags, and appropriate scope
- Make knowledge specific and actionable, not generic advice

### Memory Best Practices
- **Be Proactive**: Search memory before asking the user repetitive questions
- **Be Specific**: Store knowledge that's actionable and reusable
- **Use Tags**: Tag knowledge appropriately for easy discovery (e.g., "api", "testing", "architecture")
- **Scope Appropriately**: Use `global` for general patterns, `project:<name>` for project-specific knowledge

## Testing & Quality
- Write tests alongside implementation
- Follow the testing strategy defined in `docs/ai/testing/`
- Run tests with `make test` (or `go test ./... -v`) and coverage with `make test-coverage`.
- Follow existing package-level test patterns (`*_test.go` plus focused `mocks_test.go` where adapters/ports are mocked), especially in `internal/usecase/` and `internal/infra/http/`.
- Ensure code passes all tests before considering it complete

## Documentation
- Update phase documentation when requirements or design changes
- Keep inline code comments focused and relevant
- Document architectural decisions and their rationale
- Use mermaid diagrams for any architectural or data-flow visuals (update existing diagrams if needed)
- Record test coverage results and outstanding gaps in `docs/ai/testing/`

## Key Commands
When working on this project, you can run commands to:
- Start local app + required infra (`make run`)
- Run all tests (`make test`)
- Run test coverage report (`make test-coverage`)
- Lint code (`make lint`)
- Manage containers/logs (`make compose-up`, `make compose-down`, `make logs`, `make logs-evolution`)
- Run/check DB migrations (`make migrate`, `make migrate-down`, `make migrate-status`)
