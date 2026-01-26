# Backend Architecture

_Note:_ This document outlines the technical decisions for the Node.js backend, MongoDB data access, database schema design, authentication approach, and API design patterns.

## Node.js Backend Framework

### Recommended Approach: Express

**Rationale:**
- Mature, stable ecosystem with extensive middleware availability
- Large community support and well-documented patterns
- Flexibility for custom implementations without opinionated structure
- Low learning curve for team members
- Excellent for RESTful API development

### Alternatives Considered

| Framework | Pros | Cons |
|-----------|------|------|
| **Express** | Mature ecosystem, flexible, simple, extensive middleware | Minimal structure, requires manual setup |
| **Fastify** | High performance (2x faster), built-in schema validation, TypeScript support | Smaller ecosystem, less middleware available |
| **NestJS** | TypeScript-first, Angular-like architecture, built-in DI, great for large teams | Steep learning curve, opinionated structure, overhead for smaller projects |

**Decision:** Express provides the best balance of simplicity, maturity, and flexibility for this project's scope.

**Documentation:** https://expressjs.com/

---

## MongoDB Access/ORM

### Recommended Approach: Mongoose

**Rationale:**
- Schema validation with built-in types and custom validators
- Middleware hooks (pre/post save, validate) for business logic
- Population for relationship management
- Query builder with intuitive syntax
- Excellent for document-oriented image metadata

### Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| **Mongoose** | Schema validation, middleware hooks, active records pattern, excellent docs | Performance overhead, abstraction layer |
| **Prisma** | Type-safe queries, auto-generated types, great DX, migration management | Better for SQL databases, MongoDB support is beta, adds complexity |
| **Native Driver** | Maximum performance, no abstraction overhead, full control | Manual validation, more boilerplate, no schema enforcement |

**Decision:** Mongoose's schema validation and middleware capabilities align well with our need for data consistency and business logic hooks (e.g., thumbnail generation triggers).

s3Key will be in the ENVironment variables.

**Example Schema:**
```javascript
const imageSchema = new mongoose.Schema({
  imageId: { type: String, required: true, unique: true, index: true },
  originalFilename: String,
  thumbnailKey: String,
  webOptimizedKey: String,
  originalKey: String,
  thumbnailFileSize: Number,
  webOptimizedFileSize: Number,
  originalFileSize: Number,
  originalDimensions: { width: Number, height: Number },
  uploadedBy: { type: mongoose.Schema.Types.ObjectId, ref: 'User' },
  uploadedAt: { type: Date, default: Date.now },
  collections: [{ type: String, index: true }],
  tags: [{ type: String, index: true }],
  people: [{ name: String, index: true }],
  dateRange: {
    type: { type: String, enum: ['exact', 'range', 'approximate'] },
    exactDate: Date,
    startDate: Date,
    endDate: Date,
    approximateDate: { year: Number, month: Number }
  },
  occasion: {
    category: { type: String, enum: ['birthday', 'wedding', 'graduation', 'holiday', 'vacation', 'work_event', 'party', 'family_gathering', 'sports_event', 'concert', 'conference', 'ceremony', 'casual', 'other'] },
    eventName: String
  },
  exif: { type: mongoose.Schema.Types.Mixed },
  published: { type: Boolean, default: false },
  moderationStatus: { type: String, enum: ['pending', 'approved', 'rejected'], default: 'pending' }
});
```

**Documentation:** https://mongoosejs.com/

---

## Database Schema Design

### Collections

#### Images
- Stores metadata, S3 references, EXIF data, tagging information
- Indexed on: `imageId`, `uploadedBy`, `tags`, `people`, `uploadedAt`, `published`
- Uses embedded documents for dateRange and occasion (no separate collections needed)

#### Users
```javascript
{
  _id: ObjectId,
  email: String (unique, indexed),
  name: String,
  role: String (enum: ['admin', 'contributor']),
  collections: [{ type: String }],
  permissions: [String],
  createdAt: Date,
  lastLoginAt: Date,
  invitedBy: ObjectId (ref: User),
  status: String (enum: ['invited', 'active', 'suspended'])
}
```

#### Tags
```javascript
{
  _id: ObjectId,
  name: String (unique, indexed),
  usageCount: Number,
  createdAt: Date,
  createdBy: ObjectId (ref: User)
}
```

#### Comments
```javascript
{
  _id: ObjectId,
  imageId: String (indexed, ref: Image),
  userId: ObjectId (ref: User),
  text: String,
  parentCommentId: ObjectId (ref: Comment), // for threading
  createdAt: Date,
  updatedAt: Date,
  moderationStatus: String (enum: ['approved', 'flagged', 'hidden'])
}
```

**Design Principles:**
- Denormalize frequently accessed data (embed people, dateRange in images)
- Normalize entities that need independent management (users, comments)
- Use indexes strategically for search and filter operations
- Keep tag usage counts for suggestion algorithms

---

## Authentication & Authorization

### Recommended Approach: JWT with Magic Link

**Rationale:**
- **Passwordless authentication** reduces security risks (no password storage, no password reset flows)
- **Better UX** for family/non-technical users
- **Email verification** built into the flow
- **JWT tokens** provide stateless authentication, easy to scale
- **Role-based access control** with permissions stored in token

### Implementation Pattern

1. **Magic Link Flow:**
   - User enters email
   - Backend generates short-lived token (15 min expiry), stores in Redis/MongoDB
   - Email sent with magic link: `https://app.example.com/auth/verify?token=xyz`
   - User clicks link, backend validates token, issues JWT
   - JWT stored in httpOnly cookie for security

2. **JWT Payload:**
```javascript
{
  userId: "507f1f77bcf86cd799439011",
  email: "user@example.com",
  role: "contributor",
  permissions: ["images.view", "images.upload", "images.tag"],
  iat: 1642546800,
  exp: 1642633200  // 24 hour expiry
}
```

3. **Authorization Middleware:**
```javascript
const requirePermission = (permission) => {
  return (req, res, next) => {
    if (!req.user.permissions.includes(permission)) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    next();
  };
};

// Usage
router.post('/images', requirePermission('images.upload'), uploadController);
```

**Libraries:**
- `jsonwebtoken` for JWT generation/verification
- `nodemailer` for email delivery
- `express-rate-limit` to prevent magic link abuse

**Alternatives Considered:**
- Traditional email/password: More complex, security burden
- OAuth (Google/GitHub): External dependency, not all family members have accounts
- Session-based: Requires server-side storage, harder to scale

**Documentation:**
- JWT: https://jwt.io/
- nodemailer: https://nodemailer.com/

---

## API Design

### REST Structure

**Base URL:** `/api/v1`

**Endpoint Conventions:**
```
GET    /api/v1/images              # List images (with filters)
GET    /api/v1/images/:id          # Get single image
POST   /api/v1/images              # Upload image
PATCH  /api/v1/images/:id          # Update metadata
DELETE /api/v1/images/:id          # Delete image

GET    /api/v1/images/:id/comments # Get comments for image
POST   /api/v1/images/:id/comments # Add comment

GET    /api/v1/tags                # Get all tags
GET    /api/v1/tags/suggestions    # Get suggested tags

POST   /api/v1/auth/login          # Request magic link
GET    /api/v1/auth/verify         # Verify magic link token
POST   /api/v1/auth/logout         # Invalidate token

GET    /api/v1/users               # List users (admin only)
POST   /api/v1/users/invite        # Invite new user (admin only)
PATCH  /api/v1/users/:id/role      # Update user role (admin only)
```

### Error Handling

**Standardized Error Response (RFC 7807):**
```javascript
{
  "type": "https://api.example.com/errors/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "Image file size exceeds maximum allowed (15MB)",
  "instance": "/api/v1/images",
  "errors": [
    {
      "field": "file",
      "message": "File size must be less than 15MB"
    }
  ]
}
```

**Centralized Error Handler:**
```javascript
// middleware/errorHandler.js
module.exports = (err, req, res, next) => {
  const status = err.status || 500;
  const response = {
    type: err.type || 'about:blank',
    title: err.title || 'Internal Server Error',
    status: status,
    detail: err.message,
    instance: req.path
  };
  
  if (err.errors) response.errors = err.errors;
  if (process.env.NODE_ENV === 'development') response.stack = err.stack;
  
  res.status(status).json(response);
};
```

**HTTP Status Codes:**
- `200` OK - Successful GET, PATCH
- `201` Created - Successful POST
- `204` No Content - Successful DELETE
- `400` Bad Request - Validation errors
- `401` Unauthorized - Missing/invalid token
- `403` Forbidden - Insufficient permissions
- `404` Not Found - Resource doesn't exist
- `409` Conflict - Duplicate resource
- `413` Payload Too Large - File size exceeded
- `429` Too Many Requests - Rate limit exceeded
- `500` Internal Server Error - Unexpected errors

### Query Parameters for Filtering

```
GET /api/v1/images?tags=vacation,beach&people=John&dateFrom=2024-01-01&dateTo=2024-12-31&limit=20&offset=0
```

**Response Format:**
```javascript
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

---

## Proof of Concept Example

**Express + Mongoose Basic Setup:**
```javascript
// server.js
const express = require('express');
const mongoose = require('mongoose');
const helmet = require('helmet');
const cors = require('cors');
const rateLimit = require('express-rate-limit');

const app = express();

// Security middleware
app.use(helmet());
app.use(cors({ origin: process.env.FRONTEND_URL, credentials: true }));
app.use(express.json({ limit: '10mb' }));

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100
});
app.use('/api/', limiter);

// Database connection
mongoose.connect(process.env.MONGODB_URI, {
  useNewUrlParser: true,
  useUnifiedTopology: true
});

// Routes
app.use('/api/v1/images', require('./routes/images'));
app.use('/api/v1/auth', require('./routes/auth'));
app.use('/api/v1/users', require('./routes/users'));

// Error handling
app.use(require('./middleware/errorHandler'));

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`Server running on port ${PORT}`));
```

---

## Summary

This architecture prioritizes **simplicity, security, and scalability** while leveraging proven technologies. Express and Mongoose provide a solid foundation with minimal complexity, JWT with magic links offers excellent security and UX, and the REST API design follows industry standards for consistency and maintainability.
