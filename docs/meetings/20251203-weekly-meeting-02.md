# Summer Project — Meeting Notes (2025‑12‑03)

## Summary
- Reviewed project conventions (commits, labels), storage approach (S3 classes, Intelligent Tiering), high-level architecture direction (web app), roles/scopes, tagging/search strategy, and UI flows.
- Agreed to iterate in a spiral: refine roles, UI, and data model across passes.
 - Agreed to iterate in a spiral: refine roles, UI, and data model across passes.

   ![Iterative development diagram](../images/iterative-development.jpeg)

   [Iterative development diagram — open image](../images/iterative-development.jpeg)




## Decisions
- Web app as initial platform constraint; mobile-friendly, desktop-first is fine.
- Use predefined, flat tags; consider namespaces for clarity (e.g., `people/Bob`, `event/Wedding`).
- Focus initial role on “Maintainer” (super-user); refine Contributor/Observer later.
- Melbourne AWS region preferred (assumed local user base).

## Open Questions
- People tags vs separate “People” entity; face order array deferred.
- Date storage format(s) to support year/decade/range queries.
- Intelligent Tiering vs manual tiering: who owns transitions?
- Upload limits and chunking strategy; CDN specifics for peak events.

## Actions

- Project Conventions
  - Add commit message standards (branch prefixes, issue refs, lowercase style) to `CONTRIBUTING.md`. Owner: Speaker 1
  - Draft label taxonomy (e.g., `priority/*`, `people/*`, `area/*`) for team review before mass-adding. Owner: Speaker 1
  - Enable branch protections (required reviews/checks). Owner: Speaker 1

- Storage & Costs
  - Extend S3 cost estimation from Small to Medium and Large tiers; include access frequency assumptions. Owner: Speaker 3
  - Document S3 class selection: Standard for thumbnails/web, Glacier (Deep Archive) for backups; add rationale. Owner: Speaker 3
  - Evaluate Intelligent Tiering vs manual movement; recommend approach and ownership. Owner: Speaker 3
  - Confirm AWS region (Melbourne) and note implications. Owner: Speaker 3

- Resilience & Delivery
  - Spike on CloudFront CDN for image delivery; caching for high-traffic events (e.g., weddings). Owner: Speaker 6
  - Assess cross-region replication for durability (e.g., AU → SG/US). Owner: Speaker 6

- Architecture
  - Create a “High-Level Architecture” issue: 2‑tier vs 3‑tier, core flows, constraints; track via spiral iterations. Owner: Speaker 1
  - Add Mermaid diagrams for main flows (browse, image detail, upload); trim unnecessary arrows and define Admin/Maintainer role scope on diagrams. Owner: Speaker 1

- Data Model & Search
  - Define entities: group, collection/album, media, tags, comments, roles. Owner: Speaker 6
  - Propose tag namespaces and controlled vocab/aliasing to avoid `Bob/Robert` drift. Owner: Speaker 4
  - Choose date storage formats for flexible search (year/decade/range); document “best match” rules. Owner: Speaker 4

- UI & UX
  - Continue ASCII wireframes: filter placement, people/date/tag filters, admin settings. Owner: Speaker 1
  - Investigate upload constraints and chunked uploads; add preview thumbnails step. Owner: Speaker 1
  - Define search UX toggles (tags vs comments vs all text) and display transparency (e.g., “best match” note). Owner: Speaker 1

- Roles & Permissions
  - Align on role names and scopes (Maintainer/Admin/Contributor/Observer); specify Observer rights (edit/delete own comments). Owner: Speaker 6
  - Map capabilities to UI (who sees settings, destructive actions). Owner: Speaker 6

- Collaboration
  - Start focused sub-thread on tags/data structures (Fab to convene, collect examples from other projects). Owner: Speaker 4
  - Keep PRs visible early (WIP allowed); merge once Medium/Large estimates and agreed changes are complete. Owner: Speaker 1

## Next Meeting Prep
- Review updated S3 estimates and tiering recommendation.
- Review label taxonomy and commit standards draft.
- Walk through updated Mermaid flow diagrams and ASCII wireframes.
- Decide on people-as-tags vs separate entity.
- Confirm architecture outline and CDN spike findings.