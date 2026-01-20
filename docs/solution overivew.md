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

## Ingestion Service

To efficiently handle large file uploads, we propose a separate ingestion service built with **Go** for its superior concurrent processing and memory efficiency. The service handles multipart uploads, extracts EXIF metadata (with configurable privacy filtering for GPS/device data), generates thumbnails (300px) and web-optimized versions (1920px) supporting JPEG, PNG, HEIC, and RAW formats using `imaging` and `disintegration/imaging` libraries, enforces 15MB file size limits, and utilizes goroutines for parallel image processing. Go's performance advantage (3-5x faster than Node.js/Python for image operations) and efficient memory management make it ideal for handling concurrent uploads and CPU-intensive image transformations with minimal resource overhead.

For detailed technical decisions, library comparisons, and implementation examples, see [Ingestion Service Architecture](ingestion-service-architecture.md).

## S3 Storage

## Core

## Frontend