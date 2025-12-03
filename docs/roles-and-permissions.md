# Roles and permissions

> **Note:** This pre-first release documentation outlines a possible future state and is subject to change. Features such as the “Contributor” and “Observer” roles may be deferred beyond the MVP. Similarly, entities such as "Custom metadata" may be renamed or removed as modelling evolves. Lastly, the Notes section is intended for project collaborators and can be removed prior to the first release.

A user is assigned one of the following roles for a given Group:
- Maintainer
- Contributor
- Observer

---

<!-- TOC created with the Markdown Preview Enhanced VS Code extension -->
<!-- @import "[TOC]" {cmd="toc" depthFrom=2 depthTo=2 orderedList=false} -->

<!-- code_chunk_output -->

- [Entities](#entities)
- [Scopes](#scopes)
  - [Media](#media)
  - [Collections](#collections)
  - [Groups](#groups)
  - [Backups](#backups)
- [Roles](#roles)
  - [Maintainer](#maintainer)
  - [Contributor](#contributor)
  - [Observer](#observer)
- [Notes](#notes)

<!-- /code_chunk_output -->

---


## Entities

Entities users can view (or manipulate), based on their roles:

- Media resources (e.g. images, videos)
  - image/video data
  - image metadata (e.g. EXIF data)
  - custom metadata (e.g. user-defined tags)
  - social: comments, votes, etc.
- Collections (e.g. albums)
- Groups (e.g. families, community groups, "orgs")
- Backups

## Scopes

Possible user actions assigned to each entity, or domain (e.g. "Media"):

### Media

| Scope          | Meaning                             |
| -------------- | ----------------------------------- |
| `media:read`   | Browse/search/view media + metadata |
| `media:create` | Upload new media                    |
| `media:update` | Edit metadata, tags, etc.           |
| `media:delete` | Delete media                        |
| `media:share`  | Share media (links, permissions)    |

### Collections

| Scope               | Meaning                             |
| ------------------- | ----------------------------------- |
| `collection:read`   | Browse/view albums                  |
| `collection:create` | Create albums                       |
| `collection:update` | Modify album membership or metadata |
| `collection:delete` | Delete albums                       |
| `collection:share`  | Share albums                        |
| `collection:invite` | Invite collaborators                |

### Groups

| Scope          | Meaning                              |
| -------------- | ------------------------------------ |
| `group:create` | Create group                         |
| `group:update` | Edit group info, manage member roles |
| `group:delete` | Remove group                         |
| `group:invite` | Invite members                       |

### Backups

| Scope            | Meaning             |
| ---------------- | ------------------- |
| `backup:run`     | Trigger a backup    |
| `backup:recover` | Restore from backup |

## Roles

### Maintainer

Maintainers are the super-users of their Groups and granted the following scopes:

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
backup:
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

### Observer

Observers are Group members with a more passive level of participation. This role is suitable for people who are hesitant about, or uncomfortable with, having deletion permissions.

``` yaml
media:
  - read
  - share
collection:
  - read
  - share
  - invite
```

## Notes

- A given user is limited only to the Groups, Collections, etc. they are granted access to
- For the first iteration of this project, perhaps all users are Maintainers (and Contributors and Observers are added later)
- Consider use of inheritance (e.g. does a Maintainer always have administrative privileges over all Collections in their Group?)
- Assumption: share links are not public (not by default, anyway)
- Potential future feature: lightweight custom roles
- Potential future feature: restrict downloads
- Consider adding social-specific scopes, to allow Observers to be able to edit/delete their own social interactions (e.g. comments, likes)