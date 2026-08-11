# AI usage and references

## Background

Go is new to me — my day-to-day work is backend Java
infrastructure and Node.js. I kept seeing Go listed in backend job postings,
I kept seeing Go listed in backend job postings,
so I took this assignment as a reason to learn it properly rather than reach
for a language I already knew. That meant reading the language's own
documentation alongside writing the code, and asking a lot of "why is it done
this way in Go" questions rather than translating Java patterns line by line.

## Time spent

Roughly ten hours, against the suggested two to four. Most of the difference
went into learning Go rather than writing the calculator: reading the language
documentation, understanding why pointers are the idiomatic way to express an
optional value, why interfaces are declared by the consumer, and how the
testing and coverage tooling is meant to be used. Someone already fluent in Go
would have finished the same code well inside the suggested window. I decided
that learning the language properly was worth more than matching the estimate,
and I would rather state that than imply otherwise.

## Documentation used

- [go.dev/doc](https://go.dev/doc/) — the language documentation as a whole
- [Organizing a Go module](https://go.dev/doc/modules/layout) — the project
  structure follows its *server project* guidance: the binary under `cmd/`,
  all application code under `internal/`, non-Go directories alongside them.
  I deliberately did **not** follow the widely-cited
  `golang-standards/project-layout` repository, which is not official and
  would add `pkg/` and `api/` directories with no purpose at this size.
- [Coverage profiling support for integration tests](https://go.dev/doc/build-cover)
  — background on how Go's coverage tooling works, which informed how the
  reports in `docs/` were produced and how much weight to put on the numbers.
- [Effective Go](https://go.dev/doc/effective_go) — naming, error handling,
  and the "accept interfaces, return structs" convention.

## AI usage

The project was built with Claude (Anthropic) in an extended pair-programming
session, used both to write code and to explain the language as I went.

**Learning Go.** Pointers as a way to express optional values, implicit
interface satisfaction, sentinel errors and `errors.Is`, table-driven tests,
and the standard library's HTTP server — explained against my background in
embedded C and Java/Spring, which made the differences easier to see.

**Design discussion.** Whether the assignment warranted multiple services
(it does not), 400 versus 422 for undefined arithmetic, where float rounding
belongs, and why chasing 100% coverage on a defensive branch is not worth
the test scaffolding it requires.

**Code and tests,** written iteratively: the domain layer with its tests
first, then the HTTP transport, then the frontend. Nothing was accepted
without running it.

**Documentation.** This README and the design-decision section were drafted
with the assistant as well, then edited down — the first version was several
times longer than what shipped. Trimming it was its own exercise in deciding
which decisions actually needed explaining and which were obvious from the
code.

### Sample prompts

- "I have never used Go. I will be working on a microservice architecture with
  Go — explain how Go and microservices work and how I would build one; we also
  need unit tests, integration tests and so on."
- "Is this structure actually appropriate for a microservice architecture?"
- "Let's follow a layout consistent with https://go.dev/doc/modules/layout"
- "Coverage stopped at 94% — there is something we missed, let's add it."
- "Why did `omitempty` not work here?"
- "When describing a project developed with Go and TS, what should the ideal readme file look like?

## What I verified myself

Every command in the README was run locally. The coverage reports in `docs/`
come from real test runs, both Docker images were built and the containers
started, and each API example was checked with curl before being documented.