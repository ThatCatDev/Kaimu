import { graphql } from './client';
import type {
  GetSprintQuery,
  GetSprintsQuery,
  GetActiveSprintQuery,
  GetFutureSprintsQuery,
  GetClosedSprintsQuery,
  GetSprintCardsQuery,
  GetBacklogCardsQuery,
  CreateSprintMutation,
  UpdateSprintMutation,
  DeleteSprintMutation,
  StartSprintMutation,
  CompleteSprintMutation,
  AddCardToSprintMutation,
  RemoveCardFromSprintMutation,
  SetCardSprintsMutation,
  MoveCardToBacklogMutation,
  CreateSprintInput,
  UpdateSprintInput,
} from '../graphql/generated';
import { SprintStatus } from '../graphql/generated';

// Type exports for components
export type SprintData = GetSprintsQuery['sprints'][0];
export type SprintWithBoard = NonNullable<GetSprintQuery['sprint']>;
export type SprintCard = GetSprintCardsQuery['sprintCards'][0];
export type BacklogCard = GetBacklogCardsQuery['backlogCards'][0];

// Paginated sprint result type
export interface ClosedSprintsResult {
  sprints: SprintData[];
  pageInfo: {
    hasNextPage: boolean;
    hasPreviousPage: boolean;
    startCursor: string | null;
    endCursor: string | null;
    totalCount: number;
  };
}

// Re-export SprintStatus enum
export { SprintStatus };

// Queries
const GET_SPRINT_QUERY = `
  query GetSprint($id: ID!) {
    sprint(id: $id) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
      board {
        id
        name
      }
    }
  }
`;

const GET_SPRINTS_QUERY = `
  query GetSprints($boardId: ID!) {
    sprints(boardId: $boardId) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const GET_ACTIVE_SPRINT_QUERY = `
  query GetActiveSprint($boardId: ID!) {
    activeSprint(boardId: $boardId) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const GET_FUTURE_SPRINTS_QUERY = `
  query GetFutureSprints($boardId: ID!) {
    futureSprints(boardId: $boardId) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const GET_CLOSED_SPRINTS_QUERY = `
  query GetClosedSprints($boardId: ID!, $first: Int, $after: String) {
    closedSprints(boardId: $boardId, first: $first, after: $after) {
      edges {
        node {
          id
          name
          goal
          startDate
          endDate
          status
          position
          createdAt
          updatedAt
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
        totalCount
      }
    }
  }
`;

const GET_SPRINT_CARDS_QUERY = `
  query GetSprintCards($sprintId: ID!) {
    sprintCards(sprintId: $sprintId) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      assignee {
        id
        username
        displayName
        avatarUrl
      }
      tags {
        id
        name
        color
      }
      column {
        id
        name
      }
    }
  }
`;

const GET_BACKLOG_CARDS_QUERY = `
  query GetBacklogCards($boardId: ID!) {
    backlogCards(boardId: $boardId) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      assignee {
        id
        username
        displayName
        avatarUrl
      }
      tags {
        id
        name
        color
      }
      column {
        id
        name
      }
    }
  }
`;

// Mutations
const CREATE_SPRINT_MUTATION = `
  mutation CreateSprint($input: CreateSprintInput!) {
    createSprint(input: $input) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const UPDATE_SPRINT_MUTATION = `
  mutation UpdateSprint($id: ID!, $input: UpdateSprintInput!) {
    updateSprint(id: $id, input: $input) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const DELETE_SPRINT_MUTATION = `
  mutation DeleteSprint($id: ID!) {
    deleteSprint(id: $id)
  }
`;

const START_SPRINT_MUTATION = `
  mutation StartSprint($id: ID!) {
    startSprint(id: $id) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const COMPLETE_SPRINT_MUTATION = `
  mutation CompleteSprint($id: ID!, $moveIncompleteToBacklog: Boolean) {
    completeSprint(id: $id, moveIncompleteToBacklog: $moveIncompleteToBacklog) {
      id
      name
      goal
      startDate
      endDate
      status
      position
      createdAt
      updatedAt
    }
  }
`;

const ADD_CARD_TO_SPRINT_MUTATION = `
  mutation AddCardToSprint($input: MoveCardToSprintInput!) {
    addCardToSprint(input: $input) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      sprints {
        id
        name
      }
    }
  }
`;

const REMOVE_CARD_FROM_SPRINT_MUTATION = `
  mutation RemoveCardFromSprint($input: MoveCardToSprintInput!) {
    removeCardFromSprint(input: $input) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      sprints {
        id
        name
      }
    }
  }
`;

const SET_CARD_SPRINTS_MUTATION = `
  mutation SetCardSprints($cardId: ID!, $sprintIds: [ID!]!) {
    setCardSprints(cardId: $cardId, sprintIds: $sprintIds) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      sprints {
        id
        name
      }
    }
  }
`;

const MOVE_CARD_TO_BACKLOG_MUTATION = `
  mutation MoveCardToBacklog($cardId: ID!) {
    moveCardToBacklog(cardId: $cardId) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      updatedAt
      sprints {
        id
        name
      }
    }
  }
`;

// Query functions
export async function getSprint(id: string): Promise<SprintWithBoard | null> {
  const data = await graphql<GetSprintQuery>(GET_SPRINT_QUERY, { id });
  return data.sprint ?? null;
}

export async function getSprints(boardId: string): Promise<SprintData[]> {
  const data = await graphql<GetSprintsQuery>(GET_SPRINTS_QUERY, { boardId });
  return data.sprints;
}

export async function getActiveSprint(boardId: string): Promise<SprintData | null> {
  const data = await graphql<GetActiveSprintQuery>(GET_ACTIVE_SPRINT_QUERY, { boardId });
  return data.activeSprint ?? null;
}

export async function getFutureSprints(boardId: string): Promise<SprintData[]> {
  const data = await graphql<GetFutureSprintsQuery>(GET_FUTURE_SPRINTS_QUERY, { boardId });
  return data.futureSprints;
}

export async function getClosedSprints(boardId: string, first: number = 20, after?: string): Promise<ClosedSprintsResult> {
  const data = await graphql<GetClosedSprintsQuery>(GET_CLOSED_SPRINTS_QUERY, { boardId, first, after });
  return {
    sprints: data.closedSprints.edges.map(edge => edge.node),
    pageInfo: {
      hasNextPage: data.closedSprints.pageInfo.hasNextPage,
      hasPreviousPage: data.closedSprints.pageInfo.hasPreviousPage,
      startCursor: data.closedSprints.pageInfo.startCursor ?? null,
      endCursor: data.closedSprints.pageInfo.endCursor ?? null,
      totalCount: data.closedSprints.pageInfo.totalCount,
    },
  };
}

export async function getSprintCards(sprintId: string): Promise<SprintCard[]> {
  const data = await graphql<GetSprintCardsQuery>(GET_SPRINT_CARDS_QUERY, { sprintId });
  return data.sprintCards;
}

export async function getBacklogCards(boardId: string): Promise<BacklogCard[]> {
  const data = await graphql<GetBacklogCardsQuery>(GET_BACKLOG_CARDS_QUERY, { boardId });
  return data.backlogCards;
}

// Mutation functions
export async function createSprint(input: CreateSprintInput): Promise<SprintData> {
  const data = await graphql<CreateSprintMutation>(CREATE_SPRINT_MUTATION, { input });
  return data.createSprint;
}

export async function updateSprint(id: string, input: UpdateSprintInput): Promise<SprintData> {
  const data = await graphql<UpdateSprintMutation>(UPDATE_SPRINT_MUTATION, { id, input });
  return data.updateSprint;
}

export async function deleteSprint(id: string): Promise<boolean> {
  const data = await graphql<DeleteSprintMutation>(DELETE_SPRINT_MUTATION, { id });
  return data.deleteSprint;
}

export async function startSprint(id: string): Promise<SprintData> {
  const data = await graphql<StartSprintMutation>(START_SPRINT_MUTATION, { id });
  return data.startSprint;
}

export async function completeSprint(id: string, moveIncompleteToBacklog: boolean = true): Promise<SprintData> {
  const data = await graphql<CompleteSprintMutation>(COMPLETE_SPRINT_MUTATION, {
    id,
    moveIncompleteToBacklog,
  });
  return data.completeSprint;
}

export async function addCardToSprint(cardId: string, sprintId: string): Promise<AddCardToSprintMutation['addCardToSprint']> {
  const data = await graphql<AddCardToSprintMutation>(ADD_CARD_TO_SPRINT_MUTATION, {
    input: { cardId, sprintId },
  });
  return data.addCardToSprint;
}

export async function removeCardFromSprint(cardId: string, sprintId: string): Promise<RemoveCardFromSprintMutation['removeCardFromSprint']> {
  const data = await graphql<RemoveCardFromSprintMutation>(REMOVE_CARD_FROM_SPRINT_MUTATION, {
    input: { cardId, sprintId },
  });
  return data.removeCardFromSprint;
}

export async function setCardSprints(cardId: string, sprintIds: string[]): Promise<SetCardSprintsMutation['setCardSprints']> {
  const data = await graphql<SetCardSprintsMutation>(SET_CARD_SPRINTS_MUTATION, {
    cardId,
    sprintIds,
  });
  return data.setCardSprints;
}

export async function moveCardToBacklog(cardId: string): Promise<MoveCardToBacklogMutation['moveCardToBacklog']> {
  const data = await graphql<MoveCardToBacklogMutation>(MOVE_CARD_TO_BACKLOG_MUTATION, {
    cardId,
  });
  return data.moveCardToBacklog;
}
