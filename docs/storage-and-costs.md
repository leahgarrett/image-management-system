# AWS S3 storage and access costs

## TL;DR

We store three versions of each photo optimized for different use cases:

- **Thumbnails** (30KB) → S3 Standard - fast browsing, always hot
- **Web-optimized** (300KB) → S3 Standard - full-screen viewing, instant access
- **Originals** (15MB) → S3 Intelligent-Tiering - printing/downloading, auto-tiering based on access patterns
- **Backup** (15MB) → S3 Glacier Deep Archive (Sydney) - disaster recovery only, 12-hour retrieval

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
  - Thumbnails: 30 KB average (10-50 KB) --> 300 px long side, JPEG (legacy-compliant) / WebP (best)
  - Web-optimised: 300 KB average (100-400 KB) --> 1920 px long side, JPEG (legacy) / WebP (better) / AVIF (best)
  - Originals: 15 MB average (modern phone photos 3-15 MB, professional cameras produce larger sizes) --> original size in px and format (HEIC / JPEG / RAW)
- Upload pattern: Bulk upload initially, then 50-200 new photos/month
- Viewing pattern: Recent photos viewed frequently, older photos rarely
- Retrieval:
  - Normal downloads: originals retrieved instantly from S3 Intelligent-Tiering (printing, sharing, viewing)
  - Glacier backup: only accessed for disaster recovery (data loss, corruption, regional failure)
  - Expected Glacier retrievals: 0
- No S3 Select: Image files don't benefit from query-in-place

## Reasoning:

- Thumbnails are not eligible for auto-tiering due to smaller size (< 128KB). When added in Intelligent-Tiering, they are always stored at the "Frequent Access", which is identical to S3 Standard, but adds unnecessary complexity.
- Web-optimised images, when stored at the Intelligent - Tiering, incur monitoring fees in addition to storage fees per volume unit.
- WebP and AVIF - not supported by IE, iOS before 14 (2020 release)
- JPEG format is the safest option to access all users; easy to implement; low cost
- On-the-fly transformation with [Dynamic Image Transformation for Amazon CloudFront](https://aws.amazon.com/solutions/implementations/dynamic-image-transformation-for-amazon-cloudfront/) - great support for all users; higher architecture complexity; variable costs. Keep in mind for a later app version.

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

Estimations are based on the following assumptions:

- One photo upload = 3 PUT/POST requests (thumbnail + web-optimised -> S3 Standard, original -> S3 IT)
- Estimations are per "family unit" (scaling: GET requests - exponential, other types of requests - linear, )

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

- Frequent Access: 5% (Only the newest uploads)
- Infrequent Access (30 days): 10% (Last month's memories)
- Archive Instant Access (90 days): 85% (The rest of the history)

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
