---
title: GraphQL API
description: Pulse GraphQL API reference
---

Pulse exposes a GraphQL API for all operations. The API is available at `/graphql`.

## Endpoint

- **Development**: `http://localhost:3000/graphql`
- **Production**: `https://api.yourdomain.com/graphql`

## GraphQL Playground

In development mode, a GraphQL playground is available at:
```
http://localhost:3000/ui/playground
```

## Authentication

Most queries and mutations require authentication. Include the session cookie in your requests (handled automatically by the frontend).

For API access, use the `pulse_token` cookie or pass a JWT in the Authorization header:

```
Authorization: Bearer <jwt-token>
```

## Schema Overview

### Queries

```graphql
type Query {
  # Authentication
  me: User
  oidcProviders: [OIDCProvider!]!

  # Organizations
  organizations: [Organization!]!
  organization(id: ID!): Organization

  # Projects
  projects(organizationId: ID!): [Project!]!
  project(id: ID!): Project

  # Boards
  boards(projectId: ID!): [Board!]!
  board(id: ID!): Board

  # Cards
  card(id: ID!): Card

  # Tags
  tags(projectId: ID!): [Tag!]!
}
```

### Mutations

```graphql
type Mutation {
  # Authentication
  login(input: LoginInput!): AuthPayload!
  register(input: RegisterInput!): AuthPayload!
  logout: Boolean!

  # Organizations
  createOrganization(input: CreateOrganizationInput!): Organization!
  updateOrganization(id: ID!, input: UpdateOrganizationInput!): Organization!
  deleteOrganization(id: ID!): Boolean!

  # Projects
  createProject(input: CreateProjectInput!): Project!
  updateProject(id: ID!, input: UpdateProjectInput!): Project!
  deleteProject(id: ID!): Boolean!

  # Boards
  createBoard(input: CreateBoardInput!): Board!
  updateBoard(id: ID!, input: UpdateBoardInput!): Board!
  deleteBoard(id: ID!): Boolean!

  # Cards
  createCard(input: CreateCardInput!): Card!
  updateCard(id: ID!, input: UpdateCardInput!): Card!
  moveCard(id: ID!, input: MoveCardInput!): Card!
  deleteCard(id: ID!): Boolean!

  # Tags
  createTag(input: CreateTagInput!): Tag!
  updateTag(id: ID!, input: UpdateTagInput!): Tag!
  deleteTag(id: ID!): Boolean!
}
```

## Types

### User

```graphql
type User {
  id: ID!
  username: String!
  email: String
  displayName: String
  avatarUrl: String
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Organization

```graphql
type Organization {
  id: ID!
  name: String!
  slug: String!
  description: String
  projects: [Project!]!
  members: [OrganizationMember!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Project

```graphql
type Project {
  id: ID!
  name: String!
  description: String
  organization: Organization!
  boards: [Board!]!
  tags: [Tag!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Board

```graphql
type Board {
  id: ID!
  name: String!
  project: Project!
  columns: [BoardColumn!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Card

```graphql
type Card {
  id: ID!
  title: String!
  description: String
  position: Int!
  column: BoardColumn!
  tags: [Tag!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Tag

```graphql
type Tag {
  id: ID!
  name: String!
  color: String!
  project: Project!
}
```

## Example Queries

### Get Current User

```graphql
query Me {
  me {
    id
    username
    email
    displayName
  }
}
```

### List Organizations

```graphql
query Organizations {
  organizations {
    id
    name
    slug
    projects {
      id
      name
    }
  }
}
```

### Get Board with Cards

```graphql
query Board($id: ID!) {
  board(id: $id) {
    id
    name
    columns {
      id
      name
      position
      cards {
        id
        title
        position
        tags {
          id
          name
          color
        }
      }
    }
  }
}
```

## Example Mutations

### Create Organization

```graphql
mutation CreateOrganization($input: CreateOrganizationInput!) {
  createOrganization(input: $input) {
    id
    name
    slug
  }
}
```

Variables:
```json
{
  "input": {
    "name": "My Organization"
  }
}
```

### Create Card

```graphql
mutation CreateCard($input: CreateCardInput!) {
  createCard(input: $input) {
    id
    title
    description
    column {
      id
      name
    }
  }
}
```

Variables:
```json
{
  "input": {
    "title": "New Task",
    "description": "Task description",
    "columnId": "column-uuid"
  }
}
```

### Move Card

```graphql
mutation MoveCard($id: ID!, $input: MoveCardInput!) {
  moveCard(id: $id, input: $input) {
    id
    position
    column {
      id
      name
    }
  }
}
```

Variables:
```json
{
  "id": "card-uuid",
  "input": {
    "columnId": "new-column-uuid",
    "position": 0
  }
}
```

## Error Handling

GraphQL errors follow this format:

```json
{
  "errors": [
    {
      "message": "Error description",
      "path": ["mutation", "fieldName"],
      "extensions": {
        "code": "ERROR_CODE"
      }
    }
  ]
}
```

Common error codes:
- `UNAUTHENTICATED` - Not logged in
- `FORBIDDEN` - Not authorized for this action
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Invalid input

## Rate Limiting

Currently, there are no rate limits. For production deployments, consider adding rate limiting at the reverse proxy level.
