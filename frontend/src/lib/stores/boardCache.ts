import type { BoardWithColumns } from '../api/boards';

// Simple in-memory cache for board data that persists across View Transitions
const boardCache = new Map<string, BoardWithColumns>();

export function getCachedBoard(boardId: string): BoardWithColumns | null {
  return boardCache.get(boardId) ?? null;
}

export function setCachedBoard(boardId: string, board: BoardWithColumns): void {
  boardCache.set(boardId, board);
}
