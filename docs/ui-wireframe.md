# UI Wireframes

This document contains low-fi wireframes for version one of the Image Management System.

The aim is to keep the experience simple, friendly and easy to use for both admins and casual viewers.

See also: [UI Flows](./ui-flows.md).

---

## Table of contents

1. [Home](#home)  
2. [Browse grid](#browse-grid)  
3. [Image detail](#image-detail)  
4. [Upload](#upload)  
5. [Batch tagging](#batch-tagging)  
6. [Publish locally](#publish-locally)  
7. [Admin settings](#admin-settings)  
8. [V1 simplicity notes](#v1-simplicity-notes)

---

## Home

    +-------------------------------------------------------+
    |  Collection: [ All ▼ ]      Search: [__________]      |
    +-------------------------------------------------------+
    
    Welcome back
    
    [ Browse Images ]      [ Upload Images ] (admin only)
    [ Publish Locally ]    [ Admin Settings ] (admin only)
    
    Recent activity:
    [ thumbnail ][ thumbnail ][ thumbnail ][ thumbnail ]

**Notes**

- Collection dropdown defaults to **All**.
- Search is simple keyword search (tags, comments, text).
- Admin-only actions are not shown to casual users.

---

## Browse grid

    +-------------------------------------------------------+
    | Collection: [ All ▼ ]   Search: [ tags/comments ]     |
    +-------------------------------------------------------+
    | Filters: [Tags ▼] [People ▼] [Year ▼] [Type ▼] [Src ▼]|
    +-------------------------------------------------------+
    
    [   img   ] [   img   ] [   img   ] [   img   ]
    [   img   ] [   img   ] [   img   ] [   img   ]
    [   img   ] [   img   ] [   img   ] [   img   ]

**Notes**

- Thumbnails keep original aspect ratio.
- Filters support tags, people, year, type and source.
- Clicking a thumbnail opens the Image detail view.

---

## Image detail

    +-------------------------------------------------------+
    | < Back                      Collection: [ All ▼ ]     |
    +-------------------------------------------------------+
    
                        [   LARGE IMAGE   ]
    
    Tags:
    [pill] [pill] [pill]      [+ Add Tag]  (admin only)
    
    Comments:
    -----------------------------------------------
    | Comment by user...                         |
    -----------------------------------------------
    | Another comment...                         |
    -----------------------------------------------
    
    [ + Add Comment ]  (admin only)
    
    [ More details ▼ ]  (EXIF / advanced metadata)

**Notes**

- Admins can add tags and comments.
- Casual users see tags and comments but no add controls.
- EXIF and advanced metadata are hidden behind “More details”.

---

## Upload

    +-------------------------------------------------------+
    | Upload Images                                          |
    +-------------------------------------------------------+
    
    Step 1 of 3
    
    [ Drag files here ]
                or
    [ Choose files ]
    
    Selected:
    
    [ img ] [ img ] [ img ] [ img ] ...
    
    [ Next: Batch tagging ]
    [ Cancel ]

**Notes**

- Supports drag-and-drop and file picker.
- Large batches (200+) allowed.
- Metadata extraction happens automatically after selection.

---

## Batch tagging

    +-------------------------------------------------------+
    | Batch Tagging                                         |
    +-------------------------------------------------------+
    
    Step 2 of 3
    
    Apply to all images in this batch:
    
    Recently used:
    [pill] [pill] [pill]
    
    Suggested (from existing tags):
    [pill] [pill]
    
    Add new tag:
    [______________]  [ + Add ]
    
    ---------------------------------------------------------
    Per-image overrides:
    ---------------------------------------------------------
    [ img ]  Tags: [pill] [x] [pill] [x]  [ + Add ]
    [ img ]  Tags: [pill]      [ + Add ]
    [ img ]  Tags:             [ + Add ]
    ---------------------------------------------------------
    
    [ Back ]
    [ Next: Review ]

**Notes**

- Batch tags apply to all images by default.
- Per-image overrides allow fine tuning.
- Recently used tags help avoid typos and duplicates.

---

## Publish locally

    +-------------------------------------------------------+
    | Publish Locally                                       |
    +-------------------------------------------------------+
    
    Choose scope:
    
    (•) All images
    ( ) This collection: [ All ▼ ]
    
    [ Publish ]
    
    Status:
    [✔] Generating thumbnails…
    [✔] Building HTML bundle…
    [✔] Writing files to folder…
    
    Output folder:
    [ /path/to/exported-bundle ]

**Notes**

- Generates static HTML + thumbnails + JSON in a folder.
- Simple, single-step flow with clear progress.

---

## Admin settings

    +-------------------------------------------------------+
    | Admin Settings                                        |
    +-------------------------------------------------------+
    
    People & Roles
    ---------------------------------------------------------
    | Name        | Email               | Role    | Actions |
    ---------------------------------------------------------
    | Alice       | ...                 | Admin   | [Edit]  |
    | Bob         | ...                 | Viewer  | [Edit]  |
    ---------------------------------------------------------
    [ + Add person ]   (UI only in V1)
    
    Collection access lists
    ---------------------------------------------------------
    Collection: [ All ▼ ]
    [ Person A ] [ Person B ] [ + Add person ]
    ---------------------------------------------------------
    
    Future settings
    ---------------------------------------------------------
    (Placeholder for later options)
    ---------------------------------------------------------

**Notes**

- V1 is UI only (no backend wiring yet).
- Designed to be simple for non-technical users.
- Focus is on showing intent of roles and access, not full behaviour.

---

## V1 simplicity notes

Version one is deliberately minimal:

- No favourites.  
- No albums (collections only, with tag/date-driven browsing).  
- No moderation workflow.  
- Tags and comments are admin-only actions.  
- EXIF and advanced metadata are hidden by default.  
- Access management is UI only.  
- Publish flow is a single, simple step.

Future versions can build on this without breaking the mental model for early users.