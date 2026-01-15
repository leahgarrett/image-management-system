# Solution Overview

## Table of Contents

1. [Version 1 Features](#version-1-features)
   - [Core Features](#core-features)
   - [Features for V2 and beyond](#features-for-v2-and-beyond)
2. [Technical Overview](#technical-overview)
3. [Ingestion Service](#ingestion-service)
4. [S3 Storage](#s3-storage)
5. [Core](#core)
6. [Frontend](#frontend)

---

## Version 1 Features

### Core Features

**Browse & View**
- Grid view of images with original aspect ratio thumbnails
- Search across tags, comments, and metadata text fields
- Multi-select filters (tags, date, people, type, source)
- Image detail view with EXIF data (hidden behind toggle)

**Upload & Organization**
- Drag-and-drop or select files to upload
- Preview thumbnails before uploading
- Add tags and comments to individual images
- Batch tagging with recently used tags, suggested tags, and free text entry
- Review and confirm before finalizing upload

**Admin**
- Invite people and assign roles
- Manage roles and permissions
- Publish functionality (admin only)

### Features for V2 and beyond
- Favourites
- Albums
- Moderation workflow
- Publish to static HTML
- Tag creation is discouraged in favor of reusing existing tags


## Technical Overview
- Frontend: React, TypeScript, Tailwind CSS
- Backend: Node.js, Express
- Database: MongoDB 
- Image Storage: AWS S3
- Authentication: JWT-based authentication system
- Deployment: Docker, Kubernetes
- CI/CD: GitHub Actions for automated testing and deployment
- Testing: Jest and React Testing Library for frontend, Mocha and Chai for backend
- Monitoring: 
- Logging:

Diagram
Propose separating the upload service from the main application to handle large file uploads more efficiently.
![Architecture Sketch](images/architecture%20sketch%20-%20revised%20whiteboard.png)

## Ingestion service
To efficiently handle large file uploads, we propose a separate ingestion service. This service will manage the upload process.

## S3 Storage

## Core

## Frontend