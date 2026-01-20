# UI Wireframes

This document contains low-fi wireframes for version one of the Image Management System.

The aim is to keep the experience simple, friendly and easy to use for both admins and casual viewers.

See also: [UI Flows](./ui-flows.md).

---

## Table of contents

- [Table of contents](#table-of-contents)
- [Home](#home)
- [Browse grid](#browse-grid)
- [Image detail](#image-detail)
- [Upload](#upload)
- [Batch tagging](#batch-tagging)
- [Review](#review)
- [Admin settings](#admin-settings)
- [V1 simplicity notes](#v1-simplicity-notes)

---

## Home

    +-------------------------------------------------------+
    |  Collection: [ All ▼ ]      Search: [__________]      |
    +-------------------------------------------------------+
    
    Welcome back
    
    [ Browse Images ]      
    [ Upload Images ]    [ Admin Settings ] (admin only)
    
    Recent activity:
    [ thumbnail ][ thumbnail ][ thumbnail ][ thumbnail ]

**Notes**

- Collection dropdown defaults to **All**.
- Search is simple keyword search (tags, comments, text).
- Admin-only actions are not shown to other users.

---

## Browse grid

    +-------------------------------------------------------+
    | Collection: [ All ▼ ]   Search: [ tags/comments ]     |
    +-------------------------------------------------------+
    | Filters: [Tags ▼] [People ▼] [Date ▼] [Type ▼] [Src ▼]|
    +-------------------------------------------------------+
    
    [   img   ] [   img   ] [   img   ] [   img   ]
    [   img   ] [   img   ] [   img   ] [   img   ]
    [   img   ] [   img   ] [   img   ] [   img   ]

**Notes**

- Thumbnails keep original aspect ratio.
- Filters support 
    - tags
    - people
    - date taken (drop to choose specific date or range)
    - date added (drop to choose specific date or range)
    - type (photo, video, scan, other)
    - source (uploaded by attribution, import, other)
- Clicking a thumbnail opens the Image detail view.

---

## Image detail

    +-------------------------------------------------------+
    | < Back                      [ Image Name ]       |
    +-------------------------------------------------------+
    
                        [   LARGE IMAGE   ]
    
    Tags:
    [tag] [tag] [tag]      [+ Add Tag] 
    
    Comments:
    -----------------------------------------------
    | Comment by user...                         |
    -----------------------------------------------
    | Another comment...                         |
    -----------------------------------------------
    
    [ + Add Comment ]  
    
    [ More details ▼ ]  (EXIF / advanced metadata)

**Notes**

- Contributors can add tags and comments.
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
- Large batches allowed
- Metadata extraction happens automatically after selection

---

## Batch tagging

    +-------------------------------------------------------+
    | Batch Tagging                                         |
    +-------------------------------------------------------+
    
    Step 2 of 3
    
    Apply to all images in this batch:
    
    Recently used tags:
    [tag] [tag] [tag]
    
    Suggested (from existing tags):
    [tag] [tag]
    
    Add new tag:
    [______________]  [ + Add ]
    
    ---------------------------------------------------------
    Select one or more to add tag to individual images:
    ---------------------------------------------------------
    [x] [ img ]  Tags: [tag]  [tag] 
    [ ] [ img ]  Tags: [tag]     
    [x] [ img ]  Tags: [tag]     
    [x] [ img ]  Tags:             
    ---------------------------------------------------------
    
    [ Back ]
    [ Next: Review ]

**Notes**

- Use list style select to make it easy to see images and their current tags
- Show recently used tags and similar tags help avoid typos and duplicates

## Review

    +-------------------------------------------------------+
    | Review & Confirm                                      |
    +-------------------------------------------------------+
    
    Step 3 of 3
    
    You are about to upload 50 images with the following tags:
    
    [tag] [tag] [tag] ...
    
    Confirm that everything looks good before proceeding.
    
    [ Back ]
    [ Confirm & Upload ]
---

## Admin settings

    +-------------------------------------------------------+
    | Admin Settings                                        |
    +-------------------------------------------------------+
    
    People 
    ---------------------------------------------------------
    | Name        | Email               | Role        | Actions |
    ---------------------------------------------------------
    | Alice       | ...                 | Admin        | [Edit]  |
    | Bob         | ...                 | Contributor  | [Edit]  |
    ---------------------------------------------------------
    [ + Add person ]   
    
    Collection access lists
    ---------------------------------------------------------
    Collection: [ All ▼ ]
    ---------------------------------------------------------
    ---------------------------------------------------------
    | Name            | Collection   | Actions |
    ---------------------------------------------------------
    | Alice           | Collection A | [Edit]  |
    | Alice           | Collection B | [Edit]  |
    | Alice           | Collection C | [Edit]  |
    | Bob             | Collection A | [Edit]  |
    | Bob             | Collection B | [Edit]  |
    ---------------------------------------------------------
    [ + Add person / collection pair ]  


    Roles
    ---------------------------------------------------------
    | Role            | Actions |
    ---------------------------------------------------------
    | Admin           | `media:create` | [X]  |
    | Admin           | `media:update` | [X]  |
    | Admin           | `media:delete` | [X]  |
    | Contributor     | `media:create` | [X]  |
    | Contributor     | `media:update` | [X]  |
    | Contributor     | `media:delete` | [-]  |
    ---------------------------------------------------------

**Notes**

- Only a subset of actions shown in Roles section for illustrative purposes (expect full list to be dislayed)
- This could be read only for V1 (can populate via a seed script)
<p></p> <!-- empty paragraph tag to "exit" the bullet list and fix rendering of the 'Tags' code block below -->

<!-- Admin settings cont. -->
    Tags
    ---------------------------------------------------------
    | Tag       | Usage Count |     | Actions |
    ---------------------------------------------------------
    | Tag A     10                  [Edit]
    | Tag B     0                   [Edit]
    | Tag C     23                  [Edit]
    | Tag D     555                 [Edit]
    | Tag E     11                  [Edit]
    | Tag F     22                  [Edit]
    ---------------------------------------------------------
    [ + Add tag ]  

    
**Notes**

- Admin can add tags when setting up system
- Admin can delete unused tags
<p></p> <!-- rendering fix -->

<!-- Admin settings cont. -->

    Images to be deleted
    ---------------------------------------------------------
    | Image       | Usage Count |     | Actions |
    ---------------------------------------------------------
    | Image A     1                   [Edit]
    | Image B     0                   [Edit]
    | Image C     2                   [Edit]
    | Image D     5                   [Edit]
    | Image E     1                   [Edit]
    | Image F     2                   [Edit]
    ---------------------------------------------------------

**Notes**
- Admin can see images marked for deletion
- Admin can permanently delete images
- Usage count shows how many collections the image is in

---

## V1 simplicity notes

Version one is deliberately minimal:

- No favourites.  
- No albums (collections only, with tag/date-driven browsing).  
- No moderation workflow.  
- Tags and comments are admin-only actions.  
- EXIF and advanced metadata are hidden by default.  
- Access management is UI only.  
- No "publish locally" flow.

Future versions can build on this without breaking the mental model for early users.