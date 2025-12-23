# Roles and permissions

> **Note:** This pre-first release documentation outlines a possible future state and is subject to change. Features such as the “Casual” role and "Backups" are deferred beyond V1. Similarly, entities such as "Custom metadata" may be renamed or removed as modelling evolves. Lastly, the [Notes](#notes) section is intended for project collaborators and can be removed prior to the first release.

A user is assigned one of the following roles for a given Group:
- Admin
- Contributor
- Casual *(deferred beyond V1)*

---

<!-- TOC created with the Markdown Preview Enhanced VS Code extension -->
<!-- @import "[TOC]" {cmd="toc" depthFrom=2 depthTo=2 orderedList=false} -->

<!-- code_chunk_output -->

- [Entities](#entities)
- [Scopes](#scopes)
- [Roles](#roles)
- [Notes](#notes)

<!-- /code_chunk_output -->

---


## Entities

Entities users can view (or manipulate), based on their roles:

- Media resources (e.g. images, videos)
  - image/video data
  - image metadata (e.g. EXIF data)
  - custom metadata
    - tags
    - comments
    - deletion flag with comment *(deferred beyond V1)*
- Collections (e.g. albums)
- Groups (e.g. families, community groups, "orgs")
- Backups *(deferred beyond V1)*

## Scopes

Possible user actions assigned to each entity, or domain (e.g. "Media"):

### Media

| Scope          | Meaning                                    |
| -------------- | ------------------------------------------ |
| `media:read`   | Browse/search/view media + metadata        |
| `media:create` | Upload new media                           |
| `media:update` | Edit custom metadata (tagging, commenting) |
| `media:delete` | Delete media                               |
| `media:share`  | Share media (links, permissions)           |

### Collections

| Scope               | Meaning                             |
| ------------------- | ----------------------------------- |
| `collection:read`   | Browse/view albums                  |
| `collection:create` | Create albums                       |
| `collection:update` | Modify album membership or metadata |
| `collection:delete` | Delete albums                       |
| `collection:share`  | Share albums                        |
| `collection:invite` | Invite collaborators and casuals    |


### Groups

| Scope          | Meaning                              |
| -------------- | ------------------------------------ |
| `group:create` | Create group                         |
| `group:update` | Edit group info, manage member roles |
| `group:delete` | Remove group                         |
| `group:invite` | Invite members                       |

### Backups

 *(deferred beyond V1)*

| Scope            | Meaning             |
| ---------------- | ------------------- |
| `backup:run`     | Trigger a backup    |
| `backup:recover` | Restore from backup |

## Roles

### Admin

Admins are the super-users of their Groups and granted the following scopes:

``` yaml
media:
  - read
  - create
  - update
  - delete
  - share
collection:
  - read
  - create
  - update
  - delete
  - share
  - invite
group:
  - create
  - update
  - delete
  - invite
backup: # deferred beyond V1
  - run
  - recover
```

### Contributor

Contributors are active participants, adding to and managing Collections, but not concerned with the adminstration of their Group:

``` yaml
media:
  - read
  - create
  - update
  - delete
  - share
collection:
  - read
  - create
  - update
  - delete
  - share
  - invite
```

### Casual

*(deferred beyond V1)*

Casuals are the passive participants of their Groups. This role provides read-only access to Collections and is suitable for people less comfortable navigating digital interfaces:

``` yaml
media:
  - read
  - share
collection:
  - read
  - share
```

## Notes

- A given user’s visibility is limited to the Groups, Collections, and related resources granted to them (i.e. an Admin does not have visibility of all Groups)
- Contributor has been added to scope for V1 so that Admins don't have to upload all media themselves (i.e. the Upload Flow is separated from the Admin Flow: [High-level navigation](./ui-flow.md#1-high-level-navigation))
- Consideration: use of inheritance (e.g. does an Admin always have administrative privileges over all Collections in their Group?)
- Assumption: share links are not public (not by default, anyway)
- Potential future feature: lightweight custom roles
- Potential future feature: restrict downloads
- Potential future feature: allow Casuals to be able to edit/delete their own social interactions (e.g. comments, likes)
- Potential future feature: allow Casuals to flag an image or video for deletion (along with a comment)