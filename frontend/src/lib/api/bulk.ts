import { graphql } from './client';
import type {
  BulkMoveCardsToColumnMutation,
  BulkMoveCardsToColumnMutationVariables,
  BulkUpdateCardSprintsMutation,
  BulkUpdateCardSprintsMutationVariables,
  BulkUpdateCardPropertiesMutation,
  BulkUpdateCardPropertiesMutationVariables,
  BulkTagCardsMutation,
  BulkTagCardsMutationVariables,
  BulkDeleteCardsMutation,
  BulkDeleteCardsMutationVariables,
  BulkMoveCardsToBacklogMutation,
  BulkMoveCardsToBacklogMutationVariables,
  CardPriority,
  TagOperation,
} from '../graphql/generated';

const BULK_RESULT_FIELDS = `
  successCount
  totalCount
  cards {
    id
  }
`;

const BULK_MOVE_CARDS_TO_COLUMN = `
  mutation BulkMoveCardsToColumn($input: BulkMoveCardsToColumnInput!) {
    bulkMoveCardsToColumn(input: $input) {
      ${BULK_RESULT_FIELDS}
    }
  }
`;

const BULK_UPDATE_CARD_SPRINTS = `
  mutation BulkUpdateCardSprints($input: BulkUpdateCardSprintsInput!) {
    bulkUpdateCardSprints(input: $input) {
      ${BULK_RESULT_FIELDS}
    }
  }
`;

const BULK_UPDATE_CARD_PROPERTIES = `
  mutation BulkUpdateCardProperties($input: BulkUpdateCardPropertiesInput!) {
    bulkUpdateCardProperties(input: $input) {
      ${BULK_RESULT_FIELDS}
    }
  }
`;

const BULK_TAG_CARDS = `
  mutation BulkTagCards($input: BulkTagCardsInput!) {
    bulkTagCards(input: $input) {
      ${BULK_RESULT_FIELDS}
    }
  }
`;

const BULK_DELETE_CARDS = `
  mutation BulkDeleteCards($input: BulkDeleteCardsInput!) {
    bulkDeleteCards(input: $input)
  }
`;

const BULK_MOVE_CARDS_TO_BACKLOG = `
  mutation BulkMoveCardsToBacklog($input: BulkMoveCardsToBacklogInput!) {
    bulkMoveCardsToBacklog(input: $input) {
      ${BULK_RESULT_FIELDS}
    }
  }
`;

export async function bulkMoveCardsToColumn(
  cardIds: string[],
  targetColumnId: string,
  boardId: string
) {
  const data = await graphql<BulkMoveCardsToColumnMutation>(BULK_MOVE_CARDS_TO_COLUMN, {
    input: { cardIds, targetColumnId, boardId },
  } as BulkMoveCardsToColumnMutationVariables);
  return data.bulkMoveCardsToColumn;
}

export async function bulkUpdateCardSprints(
  cardIds: string[],
  sprintId: string,
  boardId: string,
  add: boolean
) {
  const data = await graphql<BulkUpdateCardSprintsMutation>(BULK_UPDATE_CARD_SPRINTS, {
    input: { cardIds, sprintId, boardId, add },
  } as BulkUpdateCardSprintsMutationVariables);
  return data.bulkUpdateCardSprints;
}

export async function bulkUpdateCardProperties(
  cardIds: string[],
  boardId: string,
  properties: {
    priority?: CardPriority;
    assigneeId?: string;
    clearAssignee?: boolean;
    storyPoints?: number;
    clearStoryPoints?: boolean;
  }
) {
  const data = await graphql<BulkUpdateCardPropertiesMutation>(BULK_UPDATE_CARD_PROPERTIES, {
    input: { cardIds, boardId, ...properties },
  } as BulkUpdateCardPropertiesMutationVariables);
  return data.bulkUpdateCardProperties;
}

export async function bulkTagCards(
  cardIds: string[],
  boardId: string,
  tagIds: string[],
  operation: TagOperation
) {
  const data = await graphql<BulkTagCardsMutation>(BULK_TAG_CARDS, {
    input: { cardIds, boardId, tagIds, operation },
  } as BulkTagCardsMutationVariables);
  return data.bulkTagCards;
}

export async function bulkDeleteCards(cardIds: string[], boardId: string) {
  const data = await graphql<BulkDeleteCardsMutation>(BULK_DELETE_CARDS, {
    input: { cardIds, boardId },
  } as BulkDeleteCardsMutationVariables);
  return data.bulkDeleteCards;
}

export async function bulkMoveCardsToBacklog(cardIds: string[], boardId: string) {
  const data = await graphql<BulkMoveCardsToBacklogMutation>(BULK_MOVE_CARDS_TO_BACKLOG, {
    input: { cardIds, boardId },
  } as BulkMoveCardsToBacklogMutationVariables);
  return data.bulkMoveCardsToBacklog;
}
