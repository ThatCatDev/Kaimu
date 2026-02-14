<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import { ConfirmModal } from '../ui';
  import { CardPriority, TagOperation } from '../../lib/graphql/generated';
  import type { BoardColumn, Tag } from '../../lib/api/boards';
  import type { SprintData } from '../../lib/api/sprints';
  interface MemberOption {
    user: { id: string; email: string; displayName?: string | null };
  }

  interface Props {
    selectedCount: number;
    columns?: BoardColumn[];
    sprints?: SprintData[];
    tags?: Tag[];
    members?: MemberOption[];
    canMoveCard?: boolean;
    canEditCard?: boolean;
    canDeleteCard?: boolean;
    onClearSelection: () => void;
    onMoveToColumn?: (columnId: string) => void;
    onAddToSprint?: (sprintId: string) => void;
    onRemoveFromSprint?: (sprintId: string) => void;
    onSetPriority?: (priority: CardPriority) => void;
    onSetAssignee?: (userId: string | null) => void;
    onAddTags?: (tagIds: string[]) => void;
    onRemoveTags?: (tagIds: string[]) => void;
    onMoveToBacklog?: () => void;
    onDelete?: () => void;
  }

  let {
    selectedCount,
    columns = [],
    sprints = [],
    tags = [],
    members = [],
    canMoveCard = false,
    canEditCard = false,
    canDeleteCard = false,
    onClearSelection,
    onMoveToColumn,
    onAddToSprint,
    onRemoveFromSprint,
    onSetPriority,
    onSetAssignee,
    onAddTags,
    onRemoveTags,
    onMoveToBacklog,
    onDelete,
  }: Props = $props();

  let showDeleteConfirm = $state(false);

  const priorityOptions: { value: CardPriority; label: string; color: string }[] = [
    { value: CardPriority.Urgent, label: 'Urgent', color: 'text-red-600' },
    { value: CardPriority.High, label: 'High', color: 'text-orange-600' },
    { value: CardPriority.Medium, label: 'Medium', color: 'text-yellow-600' },
    { value: CardPriority.Low, label: 'Low', color: 'text-blue-600' },
    { value: CardPriority.None, label: 'None', color: 'text-gray-500' },
  ];
</script>

{#if selectedCount > 0}
  <div class="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-gray-200 shadow-lg px-6 py-3">
    <div class="max-w-7xl mx-auto flex items-center justify-between">
      <!-- Left side: count and clear -->
      <div class="flex items-center gap-3">
        <span class="text-sm font-medium text-gray-900">
          {selectedCount} selected
        </span>
        <button
          type="button"
          class="text-sm text-gray-500 hover:text-gray-700 underline"
          onclick={onClearSelection}
        >
          Clear
        </button>
      </div>

      <!-- Right side: action buttons -->
      <div class="flex items-center gap-2">
        <!-- Move to Column -->
        {#if canMoveCard && columns.length > 0 && onMoveToColumn}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
              Column
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg py-1"
              side="top"
              align="end"
            >
              {#each columns as col}
                <DropdownMenu.Item
                  class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer"
                  onclick={() => onMoveToColumn?.(col.id)}
                >
                  {col.name}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

        <!-- Sprint -->
        {#if canMoveCard && sprints.length > 0 && (onAddToSprint || onRemoveFromSprint)}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              Sprint
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              class="z-50 min-w-[200px] bg-white border border-gray-200 rounded-lg shadow-lg py-1"
              side="top"
              align="end"
            >
              {#if onAddToSprint}
                <div class="px-3 py-1.5 text-xs font-semibold text-gray-400 uppercase">Add to Sprint</div>
                {#each sprints as sprint}
                  <DropdownMenu.Item
                    class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer"
                    onclick={() => onAddToSprint?.(sprint.id)}
                  >
                    {sprint.name}
                  </DropdownMenu.Item>
                {/each}
              {/if}
              {#if onRemoveFromSprint}
                <DropdownMenu.Separator class="my-1 border-t border-gray-100" />
                <div class="px-3 py-1.5 text-xs font-semibold text-gray-400 uppercase">Remove from Sprint</div>
                {#each sprints as sprint}
                  <DropdownMenu.Item
                    class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer"
                    onclick={() => onRemoveFromSprint?.(sprint.id)}
                  >
                    {sprint.name}
                  </DropdownMenu.Item>
                {/each}
              {/if}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

        <!-- Priority -->
        {#if canEditCard && onSetPriority}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 6a3 3 0 013-3h10a1 1 0 01.8 1.6L14.25 8l2.55 3.4A1 1 0 0116 13H6a1 1 0 00-1 1v3a1 1 0 11-2 0V6z" clip-rule="evenodd" />
              </svg>
              Priority
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              class="z-50 min-w-[140px] bg-white border border-gray-200 rounded-lg shadow-lg py-1"
              side="top"
              align="end"
            >
              {#each priorityOptions as opt}
                <DropdownMenu.Item
                  class="px-3 py-2 text-sm hover:bg-gray-100 cursor-pointer {opt.color}"
                  onclick={() => onSetPriority?.(opt.value)}
                >
                  {opt.label}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

        <!-- Assignee -->
        {#if canEditCard && members.length > 0 && onSetAssignee}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              Assignee
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              class="z-50 min-w-[200px] max-h-64 overflow-y-auto bg-white border border-gray-200 rounded-lg shadow-lg py-1"
              side="top"
              align="end"
            >
              <DropdownMenu.Item
                class="px-3 py-2 text-sm text-gray-500 hover:bg-gray-100 cursor-pointer"
                onclick={() => onSetAssignee?.(null)}
              >
                Unassign
              </DropdownMenu.Item>
              <DropdownMenu.Separator class="my-1 border-t border-gray-100" />
              {#each members as member}
                <DropdownMenu.Item
                  class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer"
                  onclick={() => onSetAssignee?.(member.user.id)}
                >
                  {member.user.displayName || member.user.email}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

        <!-- Tags -->
        {#if canEditCard && tags.length > 0 && (onAddTags || onRemoveTags)}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
              Tags
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              class="z-50 min-w-[200px] max-h-64 overflow-y-auto bg-white border border-gray-200 rounded-lg shadow-lg py-1"
              side="top"
              align="end"
            >
              {#if onAddTags}
                <div class="px-3 py-1.5 text-xs font-semibold text-gray-400 uppercase">Add Tag</div>
                {#each tags as tag}
                  <DropdownMenu.Item
                    class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center gap-2"
                    onclick={() => onAddTags?.([tag.id])}
                  >
                    <span class="w-3 h-3 rounded-full flex-shrink-0" style="background-color: {tag.color}"></span>
                    {tag.name}
                  </DropdownMenu.Item>
                {/each}
              {/if}
              {#if onRemoveTags}
                <DropdownMenu.Separator class="my-1 border-t border-gray-100" />
                <div class="px-3 py-1.5 text-xs font-semibold text-gray-400 uppercase">Remove Tag</div>
                {#each tags as tag}
                  <DropdownMenu.Item
                    class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center gap-2"
                    onclick={() => onRemoveTags?.([tag.id])}
                  >
                    <span class="w-3 h-3 rounded-full flex-shrink-0" style="background-color: {tag.color}"></span>
                    {tag.name}
                  </DropdownMenu.Item>
                {/each}
              {/if}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}

        <!-- Move to Backlog -->
        {#if canMoveCard && onMoveToBacklog}
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            onclick={onMoveToBacklog}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            Backlog
          </button>
        {/if}

        <!-- Delete -->
        {#if canDeleteCard && onDelete}
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-red-600 bg-white border border-red-300 rounded-md hover:bg-red-50"
            onclick={() => showDeleteConfirm = true}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Delete
          </button>
        {/if}
      </div>
    </div>
  </div>

  <ConfirmModal
    isOpen={showDeleteConfirm}
    title="Delete {selectedCount} Card{selectedCount !== 1 ? 's' : ''}"
    message="Are you sure you want to delete {selectedCount} card{selectedCount !== 1 ? 's' : ''}? This action cannot be undone."
    confirmText="Delete"
    cancelText="Cancel"
    variant="danger"
    onConfirm={() => { showDeleteConfirm = false; onDelete?.(); }}
    onCancel={() => showDeleteConfirm = false}
  />
{/if}
