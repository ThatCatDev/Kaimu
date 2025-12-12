import { graphql } from './client';
import type {
  BoardQuery,
  BoardQueryVariables,
  BoardsQuery,
  BoardsQueryVariables,
  TagsQuery,
  TagsQueryVariables,
  CreateBoardMutation,
  CreateBoardMutationVariables,
  UpdateBoardMutation,
  UpdateBoardMutationVariables,
  DeleteBoardMutation,
  DeleteBoardMutationVariables,
  CreateColumnMutation,
  CreateColumnMutationVariables,
  UpdateColumnMutation,
  UpdateColumnMutationVariables,
  ReorderColumnsMutation,
  ReorderColumnsMutationVariables,
  ToggleColumnVisibilityMutation,
  ToggleColumnVisibilityMutationVariables,
  DeleteColumnMutation,
  DeleteColumnMutationVariables,
  CreateCardMutation,
  CreateCardMutationVariables,
  UpdateCardMutation,
  UpdateCardMutationVariables,
  MoveCardMutation,
  MoveCardMutationVariables,
  DeleteCardMutation,
  DeleteCardMutationVariables,
  CreateTagMutation,
  CreateTagMutationVariables,
  UpdateTagMutation,
  UpdateTagMutationVariables,
  DeleteTagMutation,
  DeleteTagMutationVariables,
  CardPriority,
} from '../graphql/generated';

export type BoardWithColumns = NonNullable<BoardQuery['board']>;
export type BoardColumn = BoardWithColumns['columns'][0];
export type BoardCard = BoardColumn['cards'][0];
export type Tag = TagsQuery['tags'][0];

const BOARD_QUERY = `
  query Board($id: ID!) {
    board(id: $id) {
      id
      name
      description
      isDefault
      createdAt
      updatedAt
      project {
        id
        name
        key
        organization {
          id
          name
          slug
        }
      }
      columns {
        id
        name
        position
        isBacklog
        isHidden
        color
        wipLimit
        cards {
          id
          title
          description
          position
          priority
          dueDate
          createdAt
          updatedAt
          tags {
            id
            name
            color
          }
          assignee {
            id
            username
            displayName
          }
          sprints {
            id
            name
            status
          }
        }
      }
    }
  }
`;

const BOARDS_QUERY = `
  query Boards($projectId: ID!) {
    boards(projectId: $projectId) {
      id
      name
      description
      isDefault
      createdAt
    }
  }
`;

const TAGS_QUERY = `
  query Tags($projectId: ID!) {
    tags(projectId: $projectId) {
      id
      name
      color
      description
      createdAt
    }
  }
`;

const CREATE_BOARD_MUTATION = `
  mutation CreateBoard($input: CreateBoardInput!) {
    createBoard(input: $input) {
      id
      name
      description
      isDefault
      createdAt
    }
  }
`;

const UPDATE_BOARD_MUTATION = `
  mutation UpdateBoard($input: UpdateBoardInput!) {
    updateBoard(input: $input) {
      id
      name
      description
      updatedAt
    }
  }
`;

const DELETE_BOARD_MUTATION = `
  mutation DeleteBoard($id: ID!) {
    deleteBoard(id: $id)
  }
`;

const CREATE_COLUMN_MUTATION = `
  mutation CreateColumn($input: CreateColumnInput!) {
    createColumn(input: $input) {
      id
      name
      position
      isBacklog
      isHidden
      color
      wipLimit
      createdAt
    }
  }
`;

const UPDATE_COLUMN_MUTATION = `
  mutation UpdateColumn($input: UpdateColumnInput!) {
    updateColumn(input: $input) {
      id
      name
      color
      wipLimit
      updatedAt
    }
  }
`;

const REORDER_COLUMNS_MUTATION = `
  mutation ReorderColumns($input: ReorderColumnsInput!) {
    reorderColumns(input: $input) {
      id
      position
    }
  }
`;

const TOGGLE_COLUMN_VISIBILITY_MUTATION = `
  mutation ToggleColumnVisibility($id: ID!) {
    toggleColumnVisibility(id: $id) {
      id
      isHidden
    }
  }
`;

const DELETE_COLUMN_MUTATION = `
  mutation DeleteColumn($id: ID!) {
    deleteColumn(id: $id)
  }
`;

const CREATE_CARD_MUTATION = `
  mutation CreateCard($input: CreateCardInput!) {
    createCard(input: $input) {
      id
      title
      description
      position
      priority
      dueDate
      createdAt
      tags {
        id
        name
        color
      }
      assignee {
        id
        username
        displayName
      }
    }
  }
`;

const UPDATE_CARD_MUTATION = `
  mutation UpdateCard($input: UpdateCardInput!) {
    updateCard(input: $input) {
      id
      title
      description
      priority
      dueDate
      updatedAt
      tags {
        id
        name
        color
      }
      assignee {
        id
        username
        displayName
      }
    }
  }
`;

const MOVE_CARD_MUTATION = `
  mutation MoveCard($input: MoveCardInput!) {
    moveCard(input: $input) {
      id
      position
      column {
        id
      }
    }
  }
`;

const DELETE_CARD_MUTATION = `
  mutation DeleteCard($id: ID!) {
    deleteCard(id: $id)
  }
`;

const CREATE_TAG_MUTATION = `
  mutation CreateTag($input: CreateTagInput!) {
    createTag(input: $input) {
      id
      name
      color
      description
      createdAt
    }
  }
`;

const UPDATE_TAG_MUTATION = `
  mutation UpdateTag($input: UpdateTagInput!) {
    updateTag(input: $input) {
      id
      name
      color
      description
    }
  }
`;

const DELETE_TAG_MUTATION = `
  mutation DeleteTag($id: ID!) {
    deleteTag(id: $id)
  }
`;

// Board operations
export async function getBoard(id: string): Promise<BoardWithColumns | null> {
  const data = await graphql<BoardQuery>(BOARD_QUERY, { id } as BoardQueryVariables);
  return data.board ?? null;
}

export async function getBoards(projectId: string): Promise<BoardsQuery['boards']> {
  const data = await graphql<BoardsQuery>(BOARDS_QUERY, { projectId } as BoardsQueryVariables);
  return data.boards;
}

export async function createBoard(
  projectId: string,
  name: string,
  description?: string
): Promise<CreateBoardMutation['createBoard']> {
  const data = await graphql<CreateBoardMutation>(CREATE_BOARD_MUTATION, {
    input: { projectId, name, description },
  } as CreateBoardMutationVariables);
  return data.createBoard;
}

export async function updateBoard(
  id: string,
  name?: string,
  description?: string
): Promise<UpdateBoardMutation['updateBoard']> {
  const data = await graphql<UpdateBoardMutation>(UPDATE_BOARD_MUTATION, {
    input: { id, name, description },
  } as UpdateBoardMutationVariables);
  return data.updateBoard;
}

export async function deleteBoard(id: string): Promise<boolean> {
  const data = await graphql<DeleteBoardMutation>(DELETE_BOARD_MUTATION, {
    id,
  } as DeleteBoardMutationVariables);
  return data.deleteBoard;
}

// Column operations
export async function createColumn(
  boardId: string,
  name: string,
  isBacklog?: boolean
): Promise<CreateColumnMutation['createColumn']> {
  const data = await graphql<CreateColumnMutation>(CREATE_COLUMN_MUTATION, {
    input: { boardId, name, isBacklog },
  } as CreateColumnMutationVariables);
  return data.createColumn;
}

export async function updateColumn(
  id: string,
  name?: string,
  color?: string,
  wipLimit?: number | null,
  clearWipLimit?: boolean
): Promise<UpdateColumnMutation['updateColumn']> {
  const data = await graphql<UpdateColumnMutation>(UPDATE_COLUMN_MUTATION, {
    input: { id, name, color, wipLimit, clearWipLimit },
  } as UpdateColumnMutationVariables);
  return data.updateColumn;
}

export async function reorderColumns(
  boardId: string,
  columnIds: string[]
): Promise<ReorderColumnsMutation['reorderColumns']> {
  const data = await graphql<ReorderColumnsMutation>(REORDER_COLUMNS_MUTATION, {
    input: { boardId, columnIds },
  } as ReorderColumnsMutationVariables);
  return data.reorderColumns;
}

export async function toggleColumnVisibility(
  id: string
): Promise<ToggleColumnVisibilityMutation['toggleColumnVisibility']> {
  const data = await graphql<ToggleColumnVisibilityMutation>(TOGGLE_COLUMN_VISIBILITY_MUTATION, {
    id,
  } as ToggleColumnVisibilityMutationVariables);
  return data.toggleColumnVisibility;
}

export async function deleteColumn(id: string): Promise<boolean> {
  const data = await graphql<DeleteColumnMutation>(DELETE_COLUMN_MUTATION, {
    id,
  } as DeleteColumnMutationVariables);
  return data.deleteColumn;
}

// Card operations
export async function createCard(
  columnId: string,
  title: string,
  description?: string,
  priority?: CardPriority,
  assigneeId?: string,
  tagIds?: string[],
  dueDate?: string
): Promise<CreateCardMutation['createCard']> {
  const data = await graphql<CreateCardMutation>(CREATE_CARD_MUTATION, {
    input: { columnId, title, description, priority, assigneeId, tagIds, dueDate },
  } as CreateCardMutationVariables);
  return data.createCard;
}

export async function updateCard(
  id: string,
  title?: string,
  description?: string,
  priority?: CardPriority,
  assigneeId?: string | null,
  tagIds?: string[],
  dueDate?: string | null
): Promise<UpdateCardMutation['updateCard']> {
  // When dueDate is explicitly null, we want to clear it
  const clearDueDate = dueDate === null;
  // When assigneeId is explicitly null, we want to clear it
  const clearAssignee = assigneeId === null;
  const data = await graphql<UpdateCardMutation>(UPDATE_CARD_MUTATION, {
    input: {
      id,
      title,
      description,
      priority,
      assigneeId: clearAssignee ? undefined : assigneeId,
      clearAssignee: clearAssignee ? true : undefined,
      tagIds,
      dueDate: clearDueDate ? undefined : dueDate,
      clearDueDate: clearDueDate ? true : undefined,
    },
  } as UpdateCardMutationVariables);
  return data.updateCard;
}

export async function moveCard(
  cardId: string,
  targetColumnId: string,
  afterCardId?: string
): Promise<MoveCardMutation['moveCard']> {
  const data = await graphql<MoveCardMutation>(MOVE_CARD_MUTATION, {
    input: { cardId, targetColumnId, afterCardId },
  } as MoveCardMutationVariables);
  return data.moveCard;
}

export async function deleteCard(id: string): Promise<boolean> {
  const data = await graphql<DeleteCardMutation>(DELETE_CARD_MUTATION, {
    id,
  } as DeleteCardMutationVariables);
  return data.deleteCard;
}

// Tag operations
export async function getTags(projectId: string): Promise<Tag[]> {
  const data = await graphql<TagsQuery>(TAGS_QUERY, { projectId } as TagsQueryVariables);
  return data.tags;
}

export async function createTag(
  projectId: string,
  name: string,
  color: string,
  description?: string
): Promise<CreateTagMutation['createTag']> {
  const data = await graphql<CreateTagMutation>(CREATE_TAG_MUTATION, {
    input: { projectId, name, color, description },
  } as CreateTagMutationVariables);
  return data.createTag;
}

export async function updateTag(
  id: string,
  name?: string,
  color?: string,
  description?: string
): Promise<UpdateTagMutation['updateTag']> {
  const data = await graphql<UpdateTagMutation>(UPDATE_TAG_MUTATION, {
    input: { id, name, color, description },
  } as UpdateTagMutationVariables);
  return data.updateTag;
}

export async function deleteTag(id: string): Promise<boolean> {
  const data = await graphql<DeleteTagMutation>(DELETE_TAG_MUTATION, {
    id,
  } as DeleteTagMutationVariables);
  return data.deleteTag;
}
