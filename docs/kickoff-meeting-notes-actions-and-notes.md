# Key Points and Takeaways

## 1. Problem and Scope

- There is a gap between:
  - personal photo storage (iCloud / Google Photos etc, often messy and locked in), and
  - a shared, curated archive for families / small groups where stories and context are preserved.
- Scope is not "fix personal photo management for everything on your phone".

It is:
- a curated set of important images (events, family history, special memories)
- that multiple people can browse and enrich over time.
- Target groups include: families, extended families, close friend groups, small communities, with varying tech skills.

---

## 2. Privacy, Access and Friction

- People like the idea of secure access, but logging in is friction.
- There is strong interest in a magic link style login (email-based, no password) to keep onboarding simple.
- Some images may need to stay private, some could be more public. The group does not want to overcomplicate this yet, but:
  - roles and permissions will likely be needed later (e.g. who can invite others, who can upload, who can administer).

---

## 3. Curation and Storage Costs

- Storage is not free (e.g. S3), and uncurated uploads could blow out costs.
- There is a tension between:
  - "throw everything in"
  - versus encouraging people to curate on the way in (like old photo albums).
- Ideas raised:
  - "Similar image" detection to suggest: these all look the same, pick one.
  - Possibly a paid tier as a soft deterrent, but this adds complexity.
  - Thumbnails were suggested so browsing uses small files and full-resolution originals are only needed for printing / download.

---

## 4. Architecture Direction (Early)

- Likely local-first plus optional cloud:
  - Tier 1: Local setup (DB + files) where someone can run everything at home.
  - Tier 2: Cloud tier (e.g. S3) for easier sharing and as a backup.
- A "publish locally" idea:
  - Admin can generate a static bundle (HTML + JSON + thumbnails)
  - that can be copied to drives, shared, and used offline without special software.
- Data model:
  - Strong lean towards NoSQL / MongoDB for flexible, uneven metadata.
  - Images from phones can have rich EXIF; scanned images will have less.
  - Tags and metadata should be stored centrally and linked to each image, not baked into file names.

---

## 5. Metadata, Tags and UX

- Metadata and tagging are core to the value:
  - Extract metadata (date, location etc) from files when available.
  - Allow additional tags, descriptions and comments that become part of the archive.
- UX ideas:
  - Batch tagging when scanning / importing sets ("these are from X event with Y people").
  - Recently used tags to make adding tags quick and to reduce typos.
  - Clean data is important to avoid near-duplicate tags and categories.
- Comments and stories are seen as part of the richness of the archive, not just a nice-to-have.

---

## 6. Roles and Collaboration

- There is often a "go-to" person (the family archivist / admin) who:
  - manages uploads, structure, backups
  - invites others
  - does more of the organisational work.
- Other users might:
  - browse, comment, add stories
  - sometimes upload, sometimes just consume.
- Roles discussed (to be refined):
  - Admin / power user (sets things up, invites others, manages structure)
  - Uploader (can add images, tags, descriptions)
  - Viewer / contributor (can browse, comment, maybe suggest tags)
- There is interest in playful collaboration patterns:
  - "family review party" style sessions (e.g. picking between two photos, or talking through old images together), but this should be optional, not forced.

---

## 7. Backup, Resilience and Vendor Independence

- Desire to avoid over-dependence on a single vendor or proprietary setup.
- S3 is a strong candidate, but:
  - need to think about login / access for groups
  - need to understand costs and failure modes.
- Backup and recovery:
  - May be solved by recommended practices (e.g. regular local export / download) rather than heavy built-in features at first.
  - Publishing static bundles to drives was seen as a good, low-tech hedge.

---

# Key Actions and Owners

> You can turn this into an issue list in GitHub.

## Technical Spikes

### AWS S3 storage & costs
- Estimate costs for storage and typical family-style access patterns.
- Consider how multiple admins / groups can safely access without sharing a single personal account.
- **Owners:** Iryna (+ Fabs, with Leah to review).

### Database choice and reversibility
- Treat MongoDB / NoSQL as the default,
- Research pros and cons versus relational for this use case.
- Consider how painful it would be to change later.
- **Owner:** Leah (open to others joining).

### Tagging / metadata model
- Explore simple patterns for tags, categories, metadata fields, and how to keep the schema flexible but not overwhelming.
- Look at how to avoid duplicate / messy tags.
- **Owner:** Fabs.

## UX and Roles

### UI / flow wireframes
- Sketch flows for:
  - Admin / power user (setup, upload, publish, manage groups)
  - Casual contributor (browse, comment, add tags)
  - Simple upload & batch tagging with recently-used tags.
- **Owner:** Leah.

### Role definitions & user experiences
- Draft possible roles (admin/uploader/viewer, archive owner vs casual family member etc).
- Consider how archive/backup responsibilities differ from "what people see day to day".
- **Owner:** ___

## Process and Collaboration

### S3 / backup investigation
- Iryna to also look at what backup means if using S3 (is S3 enough, what failure points exist, plus access patterns).

### Slack channel for the project
- New summer-project channel created; use this for updates and discussion.
- Post spike findings as threads to keep noise down.

### Repo and docs workflow
- Leah to type up notes and update docs via PR so others can see diffs.

### Next meeting
- Meet again at the same time next week.
- Everyone to add findings asynchronously to the docs or Slack channel before then.