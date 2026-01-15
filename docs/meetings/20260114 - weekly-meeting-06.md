# Meeting Notes - Week 6

## Quick Recap

Leah and Iryna discussed technical details around photo storage and authentication for a project, including:
- S3 storage options
- Metadata privacy concerns
- Authentication methods like magic links

Iryna shared her research on different photo formats and storage approaches, while Leah provided input on security considerations and S3 URL obfuscation. 

## Next Steps

### Iryna's Action Items
- Review and update the S3 storage documentation to ensure all assumptions and details (especially regarding photo storage/retrieval and archive vs. intelligent tier) are current and consistent across all sections
- Review the project documentation in VS Code (with split view if needed) to check for and resolve any duplicated or outdated information, particularly around photo storage and retrieval assumptions
- Investigate and document authentication options (e.g., Magic Link, password via email, third-party OAuth) for future project versions, starting with base authentication
- Approve the UI wireframes pull request (if not already done during the meeting)

### Leah's Action Items
- Add a summary and/or links to Iryna's detailed S3 storage spike in the main project documentation, ensuring key findings and rationale are accessible
- Sense check Go language libraries and capabilities for the project to ensure Go is a suitable choice, and document justification for language selection
- Double check S3 URL obfuscation and security practices to ensure photo privacy and prevent unauthorized access, and document findings

---

## Summary

### Image Format Storage Solutions

Iryna and Leah discussed technical details about image handling and storage, with Iryna presenting her research on photo formats and storage considerations. They explored options for handling different image formats (**JPEG**, **HEIC**, **WebP**) across different devices, with Iryna proposing to:
- Maintain original files
- Serve optimized versions to users

Leah praised Iryna's detailed spike documentation and they discussed Leah's own thoughts on Go vs Node comparisons, with Amazon Q suggesting **Python Fast API** as the preferred option.

### NoSQL Database Selection Discussion

Iryna and Leah discussed database options:
- **First choice:** NoSQL databases like DynamoDB
- **Second option:** Postgres

Leah mentioned her experience with Go and its advantages, while Iryna shared anecdotal perspective that relational databases are less relevant today, citing company use of MongoDB. They agreed that a **NoSQL database** would be more flexible for their specific case involving photos and metadata.

### UI Review and Project Planning

Leah and Iryna discussed UI wireframe reviews and approvals, with Leah explaining the process of reviewing changes and approving them. They also talked about Leah's project using Go, with Leah considering Python due to broader usage. 

**Key discussion points:**
- Adding a table of contents
- Linking to Iryna's detailed document on S3 storage
- Potentially creating a separate document for detailed information
- Iryna's suggestion for storing original photos in an archive, allowing users to access them occasionally, with the option for more frequent access at a higher cost

### Data Storage and Authentication Planning

Iryna and Leah discussed the need to separate S3 types for data storage and printing purposes. They reviewed a pull request containing:
- UI changes and documentation
- A wireframe diagram representing file transfers

**Authentication discussion:**
The conversation shifted to authentication options, with Iryna mentioning the possibility of third-party authentication through services like Facebook, though they agreed to implement **basic authentication first**. 

**Privacy concerns:**
They also discussed potential privacy concerns regarding:
- Photo metadata
- Geolocation data

Iryna spoke of using of **back-end for front-end (BFF) authentication** as a secure solution. 
