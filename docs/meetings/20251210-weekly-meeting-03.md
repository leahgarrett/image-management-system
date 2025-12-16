# Summer Project — Meeting Notes (Week 3)

**Date:** December 10, 2025  
**Attendees:** Leah, Laura, Iryna  

---

## Summary

Discussed meeting structure improvements, and project planning. Key focus areas included refining user roles, admin dashboard design, UI flows, and project conventions (specifically commit squashing strategy).

---

## Key Discussions

### Meeting Structure & Effectiveness
- Identified that attempting to read through entire issues quickly wasn't valuable; team needs deeper exploration time
- Agreed on new meeting format: 30 minutes status updates + 30 minutes forward planning, optional time after the first hour to drill into specific issues or questions
- Discussed difference between workshop-style meetings vs status updates/decision-making sessions
- Noted that planning phase is naturally broader with activities in multiple directions

### User Roles & Terminology
- **Decision:** Rename roles for clarity
  - "Maintainer" → "Admin" (better reflects management capabilities)
  - "Observer" → "Casual" (avoids technical pattern/observability confusion)
  - "Contributor" remains (post-MVP consideration)
- **MVP Focus:** Admin role is priority; may include casual users with restricted access
- Admin capabilities include managing organizations, membership, and image archives

### UI/UX Flows
- Reviewed admin dashboard design and navigation
- Discussed image upload flow: users should upload directly rather than admins transferring files
- Collections feature to organize and filter images, could be used for access control
- Proposed "recent changes" feature to show new/updated images since last visit
- Delete flag system: users flag, admins review and potentially permanently delete
- Filter bar with tags, date-based search, and metadata visibility options

### Project Conventions
- **Squashing commits:** Agreed to use squash on merge as standard practice
  - Small commits useful during active development for tracking progress
  - Squashing appropriate when finalizing work to keep main branch clean
  - GitHub preserves all commit messages in squashed commit description
  - Benefits: cleaner history, easier rollback points, better for code archaeology
- Leah to update project conventions document with this approach

### Tags & Organization
- Fabs shared updated tag structure (JSON format)
- Discussion on flat vs hierarchical tagging
- Team preference for simplicity based on past project complexity issues
- Plan to provide async feedback and consider pre-populating occasion values into tags and use the proposed date range structure as it allowsflexibility

### Architecture & Roadmap
- Need to create high-level architecture overview
- Focus on lightweight end-to-end setup for non-relational MongoDB database
- Roadmap document needed for feature planning

---

## Decisions Made

1. **Recurring meetings:** Set up recurring Zoom link to avoid access issues
2. **Meeting format:** 30 min updates + 30 min forward planning + optional extra time for deep dives
3. **Role names:** Admin, Contributor (post-MVP), Casual
4. **MVP scope:** Focus on Admin role, potentially include Casual with restrictions
5. **Commit strategy:** Use squash and merge as standard practice
6. **UI terminology:** Rename "Admin" page to "Admin Dashboard" for clarity

---

## Action Items
Note: these were automatically generated from meeting discussion. Mostly these are small admin tasks or updates to large tasks. We can review these in the next meeting and make any **Issues** as required.

### Leah
- [x] Set up recurring meeting to avoid link issues
- [x] Go over project conventions document offline with squash commits and feedback from Laura
- [x] Add both Iryna and Laura as reviewers to existing PRs so they can be merged
- [x] Update convention document to explain rule set for main branch protection
- [ ] Revise UI/UX document for consistent terminology and incorporate today's discussion
- [ ] Create roadmap document
- [x] Start thinking about project architecture ahead of next week
- [x] Ask Fabs if she's okay with publishing her thoughts and refinements somewhere
- [x] Give feedback to Fabs about preference for keeping tags flat rather than occasion structure
- [x] Check with Fabs about meeting time preference and set next meeting accordingly
- [x] Publish meeting notes from last time and today using AI transcript

### Laura
- [ ] Update roles page to use "admin" and "casual" instead of "maintainer" and "observer"
- [ ] Update roles page based on actions discussed
- [ ] Do another pass on Leah's UI flow for feedback
- [ ] Show Leah how to fix the mermaid diagram duplicate errors
- [ ] Ask Fabs on Slack if she's comfortable sharing what she had shared with Leah last week

### Iryna
- [ ] Double check the squash thing in project conventions document
- [ ] Review and provide feedback on UI/UX flows and role definitions
- [ ] (Note: Unavailable this week)

---

## Next Meeting Prep

- Review updated roles documentation with new terminology
- Walk through revised UI/UX flows with consistent naming
- Discuss roadmap priorities for MVP
- Review architecture approach for MongoDB setup
- Feedback on tag structure from Fabs
- Review mermaid diagram fixes

---

## Personal Updates

- **Iryna:** Had two job interviews this week, including first behavioral interview; awaiting results
- **Leah:** Applied for position requiring portfolio creation; took longer than expected but valuable exercise
- **Laura:** Managing health and energy after redundancy mid-October and surgery; focusing on self-care and gradual return to activities

---

## Notes

- Iterative planning approach using spiral model for continuous refinement
- Zoom AI transcript and summary features being tested for meeting notes
- Captions helpful for accessibility and noisy environments
- GitHub squash preserves branch history until branch deletion
