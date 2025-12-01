# Project Conventions

This document explains how we work together on this project. It covers branching, issues, PRs, documentation, and general workflow. The goal is to keep things simple, consistent, and friendly for contributors.

## Branch naming

All branches should follow this pattern:

`prefix/issue-number-short-description`

Common prefixes:

- `feature/` — new functionality
- `fix/` — bug fixes
- `spike/` — investigations and spikes
- `design/` — UI or product design work
- `docs/` — documentation changes

Examples:

```
feature/12-upload-flow
spike/7-s3-costs
design/15-ui-wireframes
fix/22-duplicate-tags
docs/18-roles-permissions
```

Why we do this:

- Keeps the repo tidy
- Automatically links branches to issues
- Makes PRs easy to understand
- Helps multiple people work in parallel

## Issues

We use GitHub issues to track spikes, tasks, design work, documentation, bugs, and questions.

Every issue should include:

- A clear title
- A short description of the goal
- Deliverables or acceptance criteria
- An owner (if known)
- Labels for type and area

### Issues Labels

Use labels to keep the project organized and to clarify what type of work each issue represents. Apply labels when creating the issue; they can be adjusted later.

**Type labels** (nature of work):

- `type: feature`
- `type: fix`
- `type: spike`
- `type: design`
- `type: docs`
- `type: tooling`

**Area labels** (project area):

- `area: infra`
- `area: architecture`
- `area: data-model`
- `area: ux`
- `area: product`
- `area: repo` (project setup, conventions, etc.)

**Priority labels**:

- `priority: high`
- `priority: medium`
- `priority: low`

**Optional labels**:

- `good-first-issue` — invites newcomers
- `discussion` — open questions or ideas
- `help-wanted` — when a task needs extra hands

General guidance:

- Every issue should have at least one type label.
- Add an area label if it makes sense.
- Use priority labels when planning work.
- Keep labels simple; refine them as patterns emerge.

## Pull requests

**Expectations**

- Prefer small, focused PRs
- Reference the issue number (e.g. `Fixes #12`)
- Include a short explanation of what changed
- Add screenshots or diagrams when helpful
- Friendly, constructive review comments welcome

**Review process**

- PRs can be merged once they have at least one review
- Docs-only PRs can be merged more quickly
- Be kind and constructive when reviewing

## Documentation

All shared documentation lives in the `docs/` folder.

Current files include:

- `project-conventions.md` (this file)
- `kickoff-deck.md`
- `product-comparison.md`
- `database-choice.md`
- `metadata-model.md`
- `roles-and-permissions.md`
- `ui-flows.md` and `/ui-wireframes/`

Adding new docs:

- Use clear names (e.g. `publish-workflow.md`, `storage-and-costs.md`)
- Keep documents short and focused
- Use PRs for visibility so others can review

You may also want a `docs/README.md` that links and describes key documents.

## Communication

- Use the `#summer-project` Slack channel for updates
- Share progress as you go
- Ask questions freely
- Celebrate wins and small steps

## Meeting rhythm

- Light weekly catch-ups during the summer
- Async updates during the week
- Post meeting agendas ahead of time


