let selectedCardIds = $state(new Set<string>());
let isSelectionMode = $state(false);

export function getSelectedCardIds(): Set<string> {
  return selectedCardIds;
}

export function getIsSelectionMode(): boolean {
  return isSelectionMode;
}

export function getSelectedCount(): number {
  return selectedCardIds.size;
}

export function toggleSelectionMode() {
  isSelectionMode = !isSelectionMode;
  if (!isSelectionMode) {
    selectedCardIds = new Set();
  }
}

export function enterSelectionMode() {
  isSelectionMode = true;
}

export function exitSelectionMode() {
  isSelectionMode = false;
  selectedCardIds = new Set();
}

export function toggleCardSelection(cardId: string) {
  const next = new Set(selectedCardIds);
  if (next.has(cardId)) {
    next.delete(cardId);
  } else {
    next.add(cardId);
  }
  selectedCardIds = next;
}

export function selectCards(cardIds: string[]) {
  const next = new Set(selectedCardIds);
  for (const id of cardIds) {
    next.add(id);
  }
  selectedCardIds = next;
}

export function deselectCards(cardIds: string[]) {
  const next = new Set(selectedCardIds);
  for (const id of cardIds) {
    next.delete(id);
  }
  selectedCardIds = next;
}

export function clearSelection() {
  selectedCardIds = new Set();
}

export function isCardSelected(cardId: string): boolean {
  return selectedCardIds.has(cardId);
}

export function toggleSelectAll(allCardIds: string[]) {
  const allSelected = allCardIds.every(id => selectedCardIds.has(id));
  if (allSelected) {
    // Deselect all
    const next = new Set(selectedCardIds);
    for (const id of allCardIds) {
      next.delete(id);
    }
    selectedCardIds = next;
  } else {
    // Select all
    const next = new Set(selectedCardIds);
    for (const id of allCardIds) {
      next.add(id);
    }
    selectedCardIds = next;
  }
}

export function getSelectedCardIdsArray(): string[] {
  return Array.from(selectedCardIds);
}
