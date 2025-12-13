---
title: Organizations
description: Managing organizations, members, and roles in Kaimu
---

Organizations are the top-level container in Kaimu. They represent a company, team, or group that shares projects and resources.

## Creating an Organization

When you first sign up, you'll be prompted to create an organization:

1. Click **Create Organization**
2. Enter a **name** (e.g., "Acme Corp")
3. The **slug** is auto-generated from the name (e.g., "acme-corp")
4. Optionally add a **description**

You automatically become the **Owner** of organizations you create.

## Organization Settings

Access settings by clicking the gear icon on the organization page:

- **General**: Edit name, slug, and description
- **Members**: Manage team members and roles
- **Roles**: View and customize roles (Admin only)
- **Invitations**: View pending invitations

## Managing Members

### Inviting Members

1. Go to **Organization Settings** → **Members**
2. Click **Invite Member**
3. Enter the email address
4. Select a role (Admin, Member, or Viewer)
5. Click **Send Invitation**

The user receives an email invitation and can join once they create an account.

### Member Roles

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| **Owner** | Full access, cannot be removed | All permissions |
| **Admin** | Manage organization and projects | All except delete org, manage roles |
| **Member** | Contribute to projects | View, create, edit cards |
| **Viewer** | Read-only access | View only |

### Changing Roles

1. Go to **Organization Settings** → **Members**
2. Find the member in the list
3. Click the role dropdown
4. Select the new role

:::note
You cannot change the Owner's role. To transfer ownership, the current owner must do so explicitly.
:::

### Removing Members

1. Go to **Organization Settings** → **Members**
2. Find the member in the list
3. Click the **Remove** button
4. Confirm the removal

## Permissions Reference

### Organization Permissions

| Permission | Owner | Admin | Member | Viewer |
|------------|:-----:|:-----:|:------:|:------:|
| View organization | ✓ | ✓ | ✓ | ✓ |
| Manage settings | ✓ | ✓ | - | - |
| Delete organization | ✓ | - | - | - |
| Invite members | ✓ | ✓ | - | - |
| Remove members | ✓ | ✓ | - | - |
| Manage roles | ✓ | - | - | - |

### Project Permissions

| Permission | Owner | Admin | Member | Viewer |
|------------|:-----:|:-----:|:------:|:------:|
| View projects | ✓ | ✓ | ✓ | ✓ |
| Create projects | ✓ | ✓ | ✓ | - |
| Manage projects | ✓ | ✓ | - | - |
| Delete projects | ✓ | ✓ | - | - |

### Board Permissions

| Permission | Owner | Admin | Member | Viewer |
|------------|:-----:|:-----:|:------:|:------:|
| View boards | ✓ | ✓ | ✓ | ✓ |
| Create boards | ✓ | ✓ | ✓ | - |
| Manage boards | ✓ | ✓ | - | - |
| Delete boards | ✓ | ✓ | - | - |

### Card Permissions

| Permission | Owner | Admin | Member | Viewer |
|------------|:-----:|:-----:|:------:|:------:|
| View cards | ✓ | ✓ | ✓ | ✓ |
| Create cards | ✓ | ✓ | ✓ | - |
| Edit cards | ✓ | ✓ | ✓ | - |
| Move cards | ✓ | ✓ | ✓ | - |
| Delete cards | ✓ | ✓ | - | - |
| Assign cards | ✓ | ✓ | ✓ | - |

## Multiple Organizations

Users can belong to multiple organizations:

- Use the organization switcher in the sidebar to switch between them
- Each organization has its own projects, boards, and members
- Roles are specific to each organization (you might be Admin in one and Viewer in another)

## Next Steps

- [Projects & Boards](/usage/projects-boards/) - Create your first project
- [Cards & Sprints](/usage/cards-sprints/) - Start tracking work
