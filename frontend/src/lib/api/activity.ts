import { graphql } from './client';
import type {
  GetOrganizationActivityQuery,
  GetOrganizationActivityQueryVariables,
  GetProjectActivityQuery,
  GetProjectActivityQueryVariables,
  GetBoardActivityQuery,
  GetBoardActivityQueryVariables,
  GetEntityHistoryQuery,
  GetEntityHistoryQueryVariables,
  AuditEntityType,
  AuditFilters,
} from '../graphql/generated';

export type AuditEvent = NonNullable<GetOrganizationActivityQuery['organizationActivity']['edges'][0]['node']>;
export type AuditEventConnection = GetOrganizationActivityQuery['organizationActivity'];

const GET_ORGANIZATION_ACTIVITY_QUERY = `
  query GetOrganizationActivity($organizationId: ID!, $first: Int, $after: String, $filters: AuditFilters) {
    organizationActivity(organizationId: $organizationId, first: $first, after: $after, filters: $filters) {
      edges {
        node {
          id
          occurredAt
          action
          entityType
          entityId
          stateBefore
          stateAfter
          metadata
          actor {
            id
            username
            displayName
            avatarUrl
          }
          organization {
            id
            name
          }
          project {
            id
            name
          }
          board {
            id
            name
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

const GET_PROJECT_ACTIVITY_QUERY = `
  query GetProjectActivity($projectId: ID!, $first: Int, $after: String) {
    projectActivity(projectId: $projectId, first: $first, after: $after) {
      edges {
        node {
          id
          occurredAt
          action
          entityType
          entityId
          stateBefore
          stateAfter
          metadata
          actor {
            id
            username
            displayName
            avatarUrl
          }
          organization {
            id
            name
          }
          project {
            id
            name
          }
          board {
            id
            name
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

const GET_BOARD_ACTIVITY_QUERY = `
  query GetBoardActivity($boardId: ID!, $first: Int, $after: String) {
    boardActivity(boardId: $boardId, first: $first, after: $after) {
      edges {
        node {
          id
          occurredAt
          action
          entityType
          entityId
          stateBefore
          stateAfter
          metadata
          actor {
            id
            username
            displayName
            avatarUrl
          }
          organization {
            id
            name
          }
          project {
            id
            name
          }
          board {
            id
            name
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

const GET_ENTITY_HISTORY_QUERY = `
  query GetEntityHistory($entityType: AuditEntityType!, $entityId: ID!, $first: Int, $after: String) {
    entityHistory(entityType: $entityType, entityId: $entityId, first: $first, after: $after) {
      edges {
        node {
          id
          occurredAt
          action
          entityType
          entityId
          stateBefore
          stateAfter
          metadata
          actor {
            id
            username
            displayName
            avatarUrl
          }
          organization {
            id
            name
          }
          project {
            id
            name
          }
          board {
            id
            name
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

export async function getOrganizationActivity(
  organizationId: string,
  first?: number,
  after?: string,
  filters?: AuditFilters
): Promise<AuditEventConnection> {
  const data = await graphql<GetOrganizationActivityQuery>(
    GET_ORGANIZATION_ACTIVITY_QUERY,
    { organizationId, first, after, filters } as GetOrganizationActivityQueryVariables
  );
  return data.organizationActivity;
}

export async function getProjectActivity(
  projectId: string,
  first?: number,
  after?: string
): Promise<AuditEventConnection> {
  const data = await graphql<GetProjectActivityQuery>(
    GET_PROJECT_ACTIVITY_QUERY,
    { projectId, first, after } as GetProjectActivityQueryVariables
  );
  return data.projectActivity;
}

export async function getBoardActivity(
  boardId: string,
  first?: number,
  after?: string
): Promise<AuditEventConnection> {
  const data = await graphql<GetBoardActivityQuery>(
    GET_BOARD_ACTIVITY_QUERY,
    { boardId, first, after } as GetBoardActivityQueryVariables
  );
  return data.boardActivity;
}

export async function getEntityHistory(
  entityType: AuditEntityType,
  entityId: string,
  first?: number,
  after?: string
): Promise<AuditEventConnection> {
  const data = await graphql<GetEntityHistoryQuery>(
    GET_ENTITY_HISTORY_QUERY,
    { entityType, entityId, first, after } as GetEntityHistoryQueryVariables
  );
  return data.entityHistory;
}
