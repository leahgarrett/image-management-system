# Meeting Notes

## Quick Recap

The team reviewed technical documentation and outlined high-level features for a photo management system, including authentication, backend microservices, and frontend development approaches. The team concluded by discussing technical implementation details for the ingestion service and database solutions, while also addressing file storage and performance considerations.

## Next Steps

### Iryna
- Complete and finalize the S3 cost estimation task today
- Research and spike on Node/MongoDB choice for core service, including database access/ORM and initial database design
- Research authentication options as a separate spike task
- Review the cost estimation PR after completion today

### Leah
- Write up and create specific tasks/issues for ingestion service spike (including Go library investigation, metadata extraction, image conversion, file size/limitation checks, and directory structure), core service spike (Node, MongoDB, ORM, initial database design), and file naming/directory structure
- Investigate and document front-end technology choices (React, CSS libraries, etc.)
- Post the created issues/tasks in the Slack channel to solicit feedback and potential participation from other team members
- DM people in the Slack channel who haven't attended meetings to determine their interest and potential task assignments
- Assign reviewers to existing pull requests and ensure reviews happen during meetings
- Schedule and prepare for next week's meeting (same time) to review progress on spikes/tasks

### Laura
- Review meeting notes and volunteer for one of the technical spike tasks

---

## Summary

### Team Updates

The group discussed their recent break and upcoming availability.
Leah mentioned that Fab would be unable to attend and planned to discuss project participation and future goals with the remaining team members.

### Project Participation and Task Planning

Leah, Irina, and Laura discussed their availability and participation in a project. Irina mentioned she would be unavailable for the last week of January due to travel to New Zealand and was unsure about her schedule after that. Laura expressed uncertainty about her time management but committed to participating in one or two sessions while away. They agreed to break tasks into smaller, manageable pieces, especially for microservices on the backend. Leah suggested Irina could help by reviewing pull requests and providing feedback, which Irina agreed to. They also discussed reaching out to others who had joined the Slack channel but not the meetings to gauge their interest in participating.

### Project Planning and Development Strategy

Leah discussed a document titled "Solution Overview" that outlines high-level features for a project, including grid view of images, aspect ratio, and thumbnails. She proposed using GitHub issues as individual tasks and suggested starting with a JSON file for frontend development while the backend is still being worked on. The team discussed the potential structure of epics, features, and tasks, deciding to keep it simple with individual tasks. They also touched on the possibility of implementing the API before the frontend and considered building a microservice for ingesting images.

### Authentication, Travel, and Project Planning

The team discussed authentication and authorization implementation, with Leah and Iryna agreeing that adding authorization early can prevent bugs and improve testing. The conversation ended with Leah offering to work on front-end tasks and suggesting an ingestion microservice project in Go, which she would be happy to lead and open for review.

### Cost Estimation and PR Review

Leah discussed the need to assign a reviewer for a PR that she couldn't merge, and decided to assign Laura to review it. Iryna mentioned she would complete the cost estimation work that day. They discussed the role of product owners and technical team members in cost estimation, with Laura explaining that in her experience, product owners typically handled this task with input from BAs or tech leads if available. Iryna then explained her approach to S3 cost estimation for different types of picture storage, including original photos, web-adjusted pictures, and thumbnails, and how she planned to use intelligent tiering to manage costs.

### Image Ingestion System Architecture Planning

The team discussed the architecture and tasks for an image ingestion system. They decided to break down the work into smaller tasks, starting with uploading and converting images, and later adding functionality to create different versions and store them in separate S3 buckets. They considered using Go for the ingestion service and discussed making it available as a REST API. The team also talked about the need for a core service to interact with the database and potentially a separate service for image conversion. They agreed to investigate technology choices and confirm library availability as part of the initial tasks.

### Ingestion Service Architecture Discussion

The team discussed the architecture of their ingestion service, deciding to implement it as a single microservice with two parallel tasks: one for creating different versions of images and another for extracting metadata. They considered using Go instead of Node for this service due to its better support for parallel processing, though none of the team members had extensive experience with Go. The team also discussed the need to set limitations on file upload size and the number of objects to upload, and debated whether to implement a feature to prevent uploading of similar images, which they left open for further discussion.

### Photo Management System Technical Planning

The team discussed technical implementation details for a photo management system. They agreed to use Node.js as the core API, MongoDB as the database, and React for the frontend. Leah proposed two technical spikes: one to investigate the ingestion service in Go, including support for various image formats and file size limitations, and another to explore Node.js options such as ORM and database design. They also discussed file naming conventions and directory structure for storing images in S3, deciding to investigate further how these might impact performance.

### Technical Implementation and Task Planning

Leah and Iryna discussed technical implementation details, with Iryna agreeing to research Node and MongoDB as potential database solutions while keeping authentication as a separate task. Leah will investigate front-end CSS libraries and create public issues for team feedback. 