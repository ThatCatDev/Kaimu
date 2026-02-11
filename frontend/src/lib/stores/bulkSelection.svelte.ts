import {
  bulkUpdateCards,
  bulkDeleteCards,
  bulkAddCardsToSprint,
  bulkRemoveCardsFromSprint,
  type BulkUpdateCardsInput,
  type BoardCard,
} from '../api/boards';
import type { CardPriority } from '../graphql/generated';

export interface BulkSelectionState {
  selectionMode: boolean;
  selectedCardIds: Set<string>;
  isLoading: boolean;
  error: string | null;
}

class BulkSelectionStore {
  selectionMode = $state(false);
  selectedCardIds = $state<Set<string>>(new Set());
  isLoading = $state(false);
  error = $state<string | null>(null);

  // Derived state
  get selectedCount() {
    return this.selectedCardIds.size;
  }

  get hasSelection() {
    return this.selectedCardIds.size > 0;
  }

  // Toggle selection mode
  toggleSelectionMode() {
    this.selectionMode = !this.selectionMode;
    if (!this.selectionMode) {
      this.clearSelection();
    }
  }

  // Enable selection mode
  enableSelectionMode() {
    this.selectionMode = true;
  }

  // Disable selection mode and clear selection
  disableSelectionMode() {
    this.selectionMode = false;
    this.clearSelection();
  }

  // Toggle a card's selection
  toggleCard(cardId: string, event?: MouseEvent) {
    // If Ctrl/Cmd+click outside selection mode, enable selection mode
    if (!this.selectionMode && event && (event.ctrlKey || event.metaKey)) {
      this.selectionMode = true;
    }

    const newSet = new Set(this.selectedCardIds);
    if (newSet.has(cardId)) {
      newSet.delete(cardId);
    } else {
      newSet.add(cardId);
    }
    this.selectedCardIds = newSet;
  }

  // Check if a card is selected
  isSelected(cardId: string): boolean {
    return this.selectedCardIds.has(cardId);
  }

  // Select multiple cards
  selectCards(cardIds: string[]) {
    const newSet = new Set(this.selectedCardIds);
    for (const id of cardIds) {
      newSet.add(id);
    }
    this.selectedCardIds = newSet;
  }

  // Deselect multiple cards
  deselectCards(cardIds: string[]) {
    const newSet = new Set(this.selectedCardIds);
    for (const id of cardIds) {
      newSet.delete(id);
    }
    this.selectedCardIds = newSet;
  }

  // Select all cards from a list
  selectAll(cardIds: string[]) {
    this.selectedCardIds = new Set(cardIds);
  }

  // Clear all selections
  clearSelection() {
    this.selectedCardIds = new Set();
    this.error = null;
  }

  // Get selected card IDs as array
  getSelectedIds(): string[] {
    return Array.from(this.selectedCardIds);
  }

  // Bulk update operations
  async bulkUpdate(updates: {
    columnId?: string;
    assigneeId?: string | null;
    clearAssignee?: boolean;
    tagIds?: string[];
    priority?: CardPriority;
    dueDate?: string | null;
    clearDueDate?: boolean;
    storyPoints?: number | null;
    clearStoryPoints?: boolean;
  }): Promise<BoardCard[]> {
    if (this.selectedCardIds.size === 0) return [];

    this.isLoading = true;
    this.error = null;

    try {
      const input: BulkUpdateCardsInput = {
        cardIds: this.getSelectedIds(),
        ...updates,
      };
      const result = await bulkUpdateCards(input);
      this.clearSelection();
      this.disableSelectionMode();
      return result;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Failed to update cards';
      throw e;
    } finally {
      this.isLoading = false;
    }
  }

  // Bulk delete
  async bulkDelete(): Promise<number> {
    if (this.selectedCardIds.size === 0) return 0;

    this.isLoading = true;
    this.error = null;

    try {
      const count = await bulkDeleteCards(this.getSelectedIds());
      this.clearSelection();
      this.disableSelectionMode();
      return count;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Failed to delete cards';
      throw e;
    } finally {
      this.isLoading = false;
    }
  }

  // Bulk add to sprint
  async bulkAddToSprint(sprintId: string): Promise<BoardCard[]> {
    if (this.selectedCardIds.size === 0) return [];

    this.isLoading = true;
    this.error = null;

    try {
      const result = await bulkAddCardsToSprint(this.getSelectedIds(), sprintId);
      this.clearSelection();
      this.disableSelectionMode();
      return result;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Failed to add cards to sprint';
      throw e;
    } finally {
      this.isLoading = false;
    }
  }

  // Bulk remove from sprint
  async bulkRemoveFromSprint(sprintId: string): Promise<BoardCard[]> {
    if (this.selectedCardIds.size === 0) return [];

    this.isLoading = true;
    this.error = null;

    try {
      const result = await bulkRemoveCardsFromSprint(this.getSelectedIds(), sprintId);
      this.clearSelection();
      this.disableSelectionMode();
      return result;
    } catch (e) {
      this.error = e instanceof Error ? e.message : 'Failed to remove cards from sprint';
      throw e;
    } finally {
      this.isLoading = false;
    }
  }
}

// Export singleton instance
export const bulkSelection = new BulkSelectionStore();
