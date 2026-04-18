# AWS S3 storage and access costs

## TL;DR

We store three versions of each photo optimized for different use cases:

- **Thumbnails** (30KB) -> S3 Standard - fast browsing, always hot
- **Web-optimized** (300KB) -> S3 Standard - full-screen viewing, instant access
- **Originals** (15MB) -> S3 Intelligent-Tiering - printing/downloading, auto-tiering based on access patterns
- **Backup** (15MB) -> S3 Glacier Deep Archive (Sydney) - disaster recovery only, 12-hour retrieval

Users always get instant access to all photos. Glacier backups are never accessed during normal operations.

**Estimated monthly costs:** Small (10GB) ~$0.83 | Medium (50GB) ~$3.83 | Large (250GB) ~$27.11

**Suggested file formats:** JPEG for thumbnails and web-optimized versions (universal compatibility, simple V1 implementation). Originals stored in their native format (HEIC/JPEG/RAW).

---

_Note:_ The prices are correct at the time of writing this doc during the initial spike. The pricing region is Asia Pacific (Melbourne). The prices are in US dollars.

The pricing page for up-to-date information: https://aws.amazon.com/s3/pricing/. Make sure to use the correct region(s).

## Prices (for the first 50 TB where applicable; per month unless stated otherwise):

- S3 Standard: $0.025 per GB
- S3 Standard-Infrequent Access: $0.0138 per GB
- S3 Intelligent - Tiering:
  - Monitoring and Automation, All Storage - $0.0025 per 1,000 objects
  - Frequent Access Tier - $0.025 per GB
  - Infrequent Access Tier, All Storage - $0.0138 per GB
  - Archive Instant Access Tier, All Storage - $0.005 per GB
  - Archive Access Tier, All Storage - $0.0045 per GB
  - Deep Archive Access Tier, All Storage - $0.002 per GB
- S3 Glacier Instant Retrieval (instant retrieval in milliseconds), All Storage - $0.005 per GB
- S3 Glacier Flexible Retrieval (1 minute to 12 hours), All Storage - $0.0045 per GB
- S3 Glacier Deep Archive (accessed once or twice in a year and can be restored within 12 hours), All Storage - $0.002 per GB

## Summary table for storage:

| Asset Type             | Primary Storage Class  | Backup Strategy         |
| ---------------------- | ---------------------- | ----------------------- |
| Thumbnails (30 KB)     | S3 Standard            | Don't back up           |
| Web-optimized (300 KB) | S3 Standard            | Don't back up           |
| Originals (15 MB)      | S3 Intelligent-Tiering | S3 Glacier Deep Archive |

## Backup suggestion

Use [S3 Glacier](https://aws.amazon.com/s3/storage-classes/glacier/) Deep Archive located in Sydney for cross-region replication.

- Lowest cost storage for long-lived archive data that is accessed less than once per year and is retrieved asynchronously.
- Retrieval: 12 hours
- Australian data sovereignty
- Same legal jurisdiction
- Good disaster recovery

## Key assumptions

- Target audience includes users with legacy devices (e.g. iPhones with iOS 13 released before 2020)
- Image sizes:
  - Thumbnails: 30 KB average (10-50 KB) --> 300 px long side, JPEG (legacy-compliant, recommended for V1) / WebP (best)
  - Web-optimised: 300 KB average (100-400 KB) --> 1920 px long side, JPEG (legacy, recommended for V1) / WebP (better) / AVIF (best)
  - Originals: 15 MB average (modern phone photos 3-15 MB, professional cameras produce larger sizes) --> original size in px and format (HEIC / JPEG / RAW)
- Upload pattern: Bulk upload initially, then 50-200 new photos/month
- Viewing pattern: Recent photos viewed frequently, older photos rarely accessed
- Access patterns:
  - Normal operations: All photos (including originals) instantly accessible from S3 Standard and Intelligent-Tiering
  - Disaster recovery: Glacier Deep Archive backups only accessed in emergency scenarios (data loss, corruption, regional failure)
  - Expected Glacier retrievals: 0 per month
  - No S3 Select: Image files don't benefit from query-in-place

## Design decisions

**Why S3 Standard for thumbnails and web-optimized:**

- Thumbnails < 128KB are not eligible for Intelligent-Tiering auto-optimization
- Frequently accessed content benefits from hot storage
- No monitoring fees (Intelligent-Tiering charges $0.0025 per 1,000 objects)

**Why Intelligent-Tiering for originals:**

- Automatic cost optimization based on access patterns
- Recent uploads stay in Frequent Access tier
- Older photos automatically move to cheaper tiers (Infrequent -> Archive Instant -> Archive Access)
- No retrieval fees when accessing archived tiers

**Why JPEG for generated versions:**

- Universal browser/device support (including iOS 13, IE)
- Simple implementation for V1
- WebP/AVIF can be added later

**Future considerations for modern image formats:**

**Option A: Pre-generate multiple formats (WebP/AVIF + JPEG)**

- Store 2 versions of thumbnails and web-optimized images (JPEG for legacy, WebP/AVIF for modern browsers)
- Backend detects client capabilities (Accept header or User-Agent) and serves appropriate format
- Pros: Faster delivery (no processing), predictable performance, simpler CDN caching
- Cons: ~2x storage cost for thumbnails/web-optimized (~22GB -> 44GB for 250GB archive), more complex upload pipeline
- Best for: High-traffic scenarios where performance is critical

**Option B: On-the-fly transformation via CloudFront**

- Use [Dynamic Image Transformation for Amazon CloudFront](https://aws.amazon.com/solutions/implementations/dynamic-image-transformation-for-amazon-cloudfront/)
- Store only JPEG, convert to WebP/AVIF on-demand based on client support
- Pros: Lower storage costs, single source of truth, automatic format optimization
- Cons: Processing latency on first request, more complex architecture, variable costs
- Best for: Cost-sensitive scenarios with moderate traffic

**Recommendation:** Start with JPEG-only (V1), evaluate traffic patterns, then choose Option A for high-traffic or Option B for cost optimization

## Estimates for small archive (10 GB):

- Upfront cost: 0.00 USD
- Monthly cost: 0.83 USD
- Total 12 months cost: 9.96 USD (Includes upfront cost)

## Estimates for med archive (50 GB):

- Upfront cost: 0.00 USD
- Monthly cost: 3.83 USD
- Total 12 months cost: 45.96 USD (Includes upfront cost)

## Estimates for large archive (250 GB):

- Upfront cost: 0.00 USD
- Monthly cost: 27.11 USD
- Total 12 months cost: 325.32 USD (Includes upfront cost)

**Estimation methodology:**

- One photo upload = 4 PUT requests (thumbnail + web-optimised -> S3 Standard, original -> S3 IT, backup -> Glacier Deep Archive)
- Costs calculated per "family unit" or group
- Scaling assumptions: GET requests grow exponentially with archive size, PUT requests scale linearly

| S3 Standard Metric                    | Small (10 GB) | Medium (50 GB) | Large (250 GB) |
| ------------------------------------- | ------------- | -------------- | -------------- |
| Storage (total GB)                    | 2.2 GB        | 11 GB          | 55 GB          |
| - _Thumbnails_ (30 KB)                | 0.2 GB        | 1 GB           | 5 GB           |
| - _Web-optimised_ (300 KB)            | 2 GB          | 10 GB          | 50 GB          |
| PUT/COPY/POST/LIST requests (monthly) | 450           | 1,300          | 5,900          |
| GET/SELECT requests (monthly)         | 5,000         | 25,000         | 100,000        |
| S3 Select                             | 0             | 0              | 0              |
| Data Transfer OUT (Internet)          | 5 GB          | 25 GB          | 200 GB         |

| S3 Intelligent-Tiering Metric         | Small (10 GB) | Medium (50 GB) | Large (250 GB) |
| ------------------------------------- | ------------- | -------------- | -------------- |
| Storage (total GB)                    | 10 GB         | 50 GB          | 250 GB         |
| PUT/COPY/POST/LIST requests (monthly) | 50            | 100            | 300            |
| GET/SELECT requests (monthly)         | 100           | 500            | 2,000          |
| Lifecycle Transitions (monthly)       | 0             | 0              | 0              |
| S3 Select                             | 0             | 0              | 0              |

**Intelligent-Tiering distribution for originals:**

- Frequent Access: 5% (newest uploads, actively viewed)
- Infrequent Access: 10% (not accessed for 30+ days)
- Archive Instant Access: 85% (not accessed for 90+ days, still instant retrieval)
- Archive Access: 0% (optional tier for 180+ days)
- Deep Archive Access: 0% (not used - separate Glacier backup exists)

| S3 Glacier Deep Archive Metric      | Small (10 GB) | Medium (50 GB) | Large (250 GB) |
| ----------------------------------- | ------------- | -------------- | -------------- |
| Storage (total GB)                  | 10 GB         | 50 GB          | 250 GB         |
| PUT/COPY/POST/LIST                  | 50            | 100            | 300            |
| All other requests, retrievals, etc | 0             | 0              | 0              |
| Data Transfer                       | 1 GB          | 2 GB           | 5 GB           |

For each object that is stored in the S3 Glacier Flexible Retrieval and S3 Glacier Deep Archive storage classes, AWS charges for 40 KB of additional metadata for each archived object, with 8 KB charged at S3 Standard rates and 32 KB charged at S3 Glacier Flexible Retrieval or S3 Deep Archive rates.

## Links:

- [Estimates for small archive (10 GB)](https://calculator.aws/#/estimate?id=a34d1c6d7f918fe55963a2100c503b0d25ea5c67)
- [Estimates for med archive (50 GB)](https://calculator.aws/#/estimate?id=cb554c6e1e51e02988512c83a5d327b0cdc221d3)
- [Estimates for large archive (250 GB)](https://calculator.aws/#/estimate?id=2fb863ded74b113c7d69bee45d06628f1fbb28ab)
- [Free tier offers 20K GET requests and 2K PUT requests](https://aws.amazon.com/free/storage/s3/)

## S3 Structure

Use prefixes (folders) based on a family/group ID
