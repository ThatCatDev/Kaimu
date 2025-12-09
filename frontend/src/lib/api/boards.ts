import { graphql } from './client';
import type {
  BoardQuery,
  BoardQueryVariables,
  BoardsQuery,
  BoardsQueryVariables,
  LabelsQuery,
  LabelsQueryVariables,
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
  CreateLabelMutation,
  CreateLabelMutationVariables,
  UpdateLabelMutation,
  UpdateLabelMutationVariables,
  DeleteLabelMutation,
  DeleteLabelMutationVariables,
  CardPriority,
} from '../graphql/generated';

export type BoardWithColumns = NonNullable<BoardQuery['board']>;
export type BoardColumn = BoardWithColumns['columns'][0];
export type BoardCard = BoardColumn['cards'][0];
export type Label = LabelsQuery['labels'][0];

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
          labels {
            id
            name
            color
          }
          assignee {
            id
            username
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

const LABELS_QUERY = `
  query Labels($projectId: ID!) {
    labels(projectId: $projectId) {
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
      labels {
        id
        name
        color
      }
      assignee {
        id
        username
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
      labels {
        id
        name
        color
      }
      assignee {
        id
        username
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

const CREATE_LABEL_MUTATION = `
  mutation CreateLabel($input: CreateLabelInput!) {
    createLabel(input: $input) {
      id
      name
      color
      description
      createdAt
    }
  }
`;

const UPDATE_LABEL_MUTATION = `
  mutation UpdateLabel($input: UpdateLabelInput!) {
    updateLabel(input: $input) {
      id
      name
      color
      description
    }
  }
`;

const DELETE_LABEL_MUTATION = `
  mutation DeleteLabel($id: ID!) {
    deleteLabel(id: $id)
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
  wipLimit?: number | null
): Promise<UpdateColumnMutation['updateColumn']> {
  const data = await graphql<UpdateColumnMutation>(UPDATE_COLUMN_MUTATION, {
    input: { id, name, color, wipLimit },
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
  labelIds?: string[],
  dueDate?: string
): Promise<CreateCardMutation['createCard']> {
  const data = await graphql<CreateCardMutation>(CREATE_CARD_MUTATION, {
    input: { columnId, title, description, priority, assigneeId, labelIds, dueDate },
  } as CreateCardMutationVariables);
  return data.createCard;
}

export async function updateCard(
  id: string,
  title?: string,
  description?: string,
  priority?: CardPriority,
  assigneeId?: string,
  labelIds?: string[],
  dueDate?: string | null
): Promise<UpdateCardMutation['updateCard']> {
  const data = await graphql<UpdateCardMutation>(UPDATE_CARD_MUTATION, {
    input: { id, title, description, priority, assigneeId, labelIds, dueDate },
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

// Label operations
export async function getLabels(projectId: string): Promise<Label[]> {
  const data = await graphql<LabelsQuery>(LABELS_QUERY, { projectId } as LabelsQueryVariables);
  return data.labels;
}

export async function createLabel(
  projectId: string,
  name: string,
  color: string,
  description?: string
): Promise<CreateLabelMutation['createLabel']> {
  const data = await graphql<CreateLabelMutation>(CREATE_LABEL_MUTATION, {
    input: { projectId, name, color, description },
  } as CreateLabelMutationVariables);
  return data.createLabel;
}

export async function updateLabel(
  id: string,
  name?: string,
  color?: string,
  description?: string
): Promise<UpdateLabelMutation['updateLabel']> {
  const data = await graphql<UpdateLabelMutation>(UPDATE_LABEL_MUTATION, {
    input: { id, name, color, description },
  } as UpdateLabelMutationVariables);
  return data.updateLabel;
}

export async function deleteLabel(id: string): Promise<boolean> {
  const data = await graphql<DeleteLabelMutation>(DELETE_LABEL_MUTATION, {
    id,
  } as DeleteLabelMutationVariables);
  return data.deleteLabel;
}
