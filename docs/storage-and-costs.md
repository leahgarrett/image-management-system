# AWS S3 storage and access costs

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

## Backup suggestion

Use S3 Glacier Flexible Retrieval or S3 Glacier Deep Archive located in Sydney:

- Australian data sovereignty
- Same legal jurisdiction
- Good disaster recovery (different availability zones)

## Estimates for small archive (5–20 GB):

- Upfront cost: 0.02 USD
- Monthly cost: 0.64 USD
- Total 12 months cost: 7.70 USD (Includes upfront cost)

For each object that is stored in the S3 Glacier Flexible Retrieval and S3 Glacier Deep Archive storage classes, AWS charges for 40 KB of additional metadata for each archived object, with 8 KB charged at S3 Standard rates and 32 KB charged at S3 Glacier Flexible Retrieval or S3 Deep Archive rates.

## Links:

- [S3 pricing](https://aws.amazon.com/s3/pricing/?refid=f3dc2e1f-b810-4402-8702-6d57e59856bd)
- [Estimates for small archive (5–20 GB)](https://calculator.aws/#/estimate?id=e0d4c29b986daeae96f3c3e8e6c3f23eb32e8b74)

## Storage Distribution by Scenario

### Small Archive (15GB total)

- Standard: 2GB (thumbnails, recent photos)
- Intelligent Tiering: 10GB (main photos)
- Glacier Flexible: 3GB (photos >2 years old)

### Medium Archive (60GB total)

- Standard: 5GB (thumbnails, recent photos)
- Intelligent Tiering: 40GB (main photos)
- Glacier Flexible: 15GB (photos >2 years old)

### Large Archive (200GB total)

- Standard: 10GB (thumbnails, recent photos)
- Intelligent Tiering: 120GB (main photos)
- Glacier Flexible: 70GB (photos >2 years old)

## Object sizes for estimation

- Originals: 15MB average (modern phone photos 3-15MB, professional cameras produce larger sizes) - used this size in Infrequent Access tier and for backup
- Thumbnails: 50KB average (150x150px JPEG)
- Web-optimized: 500KB average (1920px wide, compressed) - used this size in Standard / Frequent Access tier
