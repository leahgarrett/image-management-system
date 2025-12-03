# UI Flows

This document describes the user flows for the first version of the Image Management System.  
The goal is to keep the experience simple, clear and friendly for both admins and casual viewers.

Version one focuses on browsing, simple search, uploading, batch tagging and a basic publish flow.

---

## 1. High-level navigation

```mermaid
flowchart TD
    Home[Home] --> Browse[Browse Grid]
    Home --> Admin --> Upload[Upload Flow<br/> ]
    Home --> Admin --> Settings[Admin Settings<br/>]
    Home --> Admin --> Publish[Publish Locally<br/>]
    Home --> Admin --> People[People & Roles<br/>]

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
- Use the collection dropdown (defaults to "All")  
- Upload (admin only)  
- Publish (admin only)  

Light and simple.

---

## 3. Browse flow

```mermaid
flowchart TD
    Browse[Browse Grid]
    Filters[Filter Bar<br/>tags, date, people, type, source]
    Search[Search Bar<br/>tags + comments + text]
    Detail[Image Detail]

    Browse --> Filters
    Browse --> Search
    Filters --> Browse
    Search --> Browse
    Browse --> Detail
```

### Notes
- Original aspect ratio thumbnails  
- Multi-select filters  
- Search supports:  
  - tags  
  - comments  
  - all text fields  

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
- Admins can tag + comment  
- Casuals browse only  
- EXIF hidden behind toggle  

---

## 5. Upload flow (admin)

```mermaid
flowchart TD
    Start[Start Upload] --> SelectFiles[Select or Drag in Files]
    SelectFiles --> Preview[Preview Thumbnails]
    Preview --> BatchTag[Batch Tagging]
    BatchTag --> Review[Review All]
    Review --> Upload[Upload]
    Upload --> Done[Complete]
```

---

## 6. Batch tagging flow

```mermaid
flowchart TD
    BatchTag[Batch Tagging]
    Recent[Recently Used Tags]
    FreeText[Free Text Tag]
    Suggestions[Suggested From Existing]
    Overrides[Per-image Overrides]

    BatchTag --> Recent
    BatchTag --> FreeText
    BatchTag --> Suggestions
    BatchTag --> Overrides
```

---

## 7. Publish locally

```mermaid
flowchart TD
    PublishStart[Publish Locally] --> ChooseScope[Choose Scope<br/>All or Collection]
    ChooseScope --> Generate[Generate HTML + Thumbnails]
    Generate --> Output[Output to Folder]
```

---

## 8. Admin settings (UI only)

```mermaid
flowchart TD
    Admin[Admin Settings] --> People[People & Roles]
    Admin --> AccessLists[Access Collection Based Lists<br/>]
    Admin --> Placeholder[Future Settings]
```

---

## V1 Simplicity Notes

- No favourites  
- No albums  
- No moderation workflow  
- Comments + tags only for admins  
- EXIF hidden  
- Access settings are UI-only  
- Publish flow is single-step  