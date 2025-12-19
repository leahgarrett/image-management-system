# UI Flows

This document describes the user flows for the first version of the Image Management System.  
The goal is to keep the experience simple, clear and friendly for both admins and casual viewers.

Version one focuses on browsing, simple search, uploading, and batch tagging.

---

## Table of contents
1. [High-level navigation](#1-high-level-navigation)
2. [Home](#2-home)
3. [Browse flow](#3-browse-flow)
4. [Image detail flow](#4-image-detail-flow)
5. [Upload flow](#5-upload-flow)
6. [Batch tagging flow](#6-batch-tagging-flow)
7. [Admin settings](#7-admin-settings)  

V1 Notes [V1 Simplicity Notes](#v1-simplicity-notes)

---

## 1. High-level navigation

```mermaid
flowchart TD
    Home[Home] --> Browse
    Home --> Admin
    Home --> Upload[Upload Flow]
    Admin --> Roles[Roles<br/>]
    Admin --> People[People<br/>]

    Browse --> Detail[Image Detail]
    Browse --> SearchBar[Keyword / Tag / Comment Search]

    Upload --> BatchTag[Batch Tagging]
    BatchTag --> ReviewUpload[Review & Confirm]
    ReviewUpload --> DoneUpload[Upload Complete]
```

---

## 2. Home

### Actions
- Browse all images  
    - View image detail  
    - Search (tags, comments, text)
- Upload (images)  
- Admin settings (admin only)
- Publish (admin only)  

Light and simple.

---

## 3. Browse flow

```mermaid
flowchart TD
    Browse[Browse Grid]
    Filters[Filter Bar<br/>tags, date, people, type, source]
    Search[Search Bar<br/>]
    Detail[Image Detail]

    Browse --> Filters
    Browse --> Search
    Browse --> Detail
```

### Notes
- Original aspect ratio thumbnails  

- Search supports searching for search term in:  
  - tags  
  - comments  
  - all meta data text fields  
- Multi-select filters  
- Multiple filters can be applied at once  
- Clicking thumbnail opens Image Detail view 

---

## 4. Image detail flow

```mermaid
flowchart TD
    Detail[Image Detail]
    AddTag[Add Tag]
    AddComment[Add Comment]
    Back[Return to Browse]

    Detail --> AddTag
    Detail --> AddComment
    Detail --> Back
```

### Notes
- Users can tag + comment  
- EXIF hidden behind toggle  

---

## 5. Upload flow

```mermaid
flowchart TD
    Start[Start Upload] --> SelectFiles[Select or Drag in Files]
    SelectFiles --> Preview[Preview Thumbnails]
    Preview --> BatchTag[Batch Tagging]
    BatchTag --> Review[Review All]
    Review --> Upload[Upload]
    Upload --> Done[Complete]
```

### Notes
- Review to double check tags/comments before upload 

---

## 6. Batch tagging flow

```mermaid
flowchart TD
    BatchTag[Batch Tagging]
    Recent[Recently Used Tags]
    FreeText[Free Text Tag]
    Suggestions[Suggested From Existing]

    BatchTag --> Recent
    BatchTag --> FreeText
    BatchTag --> Suggestions
```

### Notes
- Use list display of images to select images to tag
- Show recently used tags for quick selection
- Show suggested tags from existing tags in system
- Allow free text entry for new tags  

---

## 7. Admin settings 

```mermaid
flowchart TD
    Admin[Admin Settings] --> People[People]
    Admin --> Roles[Roles<br/>]
```

### Notes
- invite people
- assign to roles
- manage roles and permissions

---

## V1 Simplicity Notes

- No favourites  
- No albums  
- No moderation workflow  
- No publish to static HTML
- When adding tags will encourage using an existing tag over creating a new one  
- EXIF hidden  
