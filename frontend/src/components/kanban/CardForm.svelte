<script lang="ts">
  import { CardPriority } from '../../lib/graphql/generated';
  import { Input, BitsSelect, DatePicker, RichTextEditor } from '../ui';
  import TagPicker from './TagPicker.svelte';
  import type { Tag } from '../../lib/api/boards';

  const priorityOptions = [
    { value: CardPriority.None, label: 'None' },
    { value: CardPriority.Low, label: 'Low' },
    { value: CardPriority.Medium, label: 'Medium' },
    { value: CardPriority.High, label: 'High' },
    { value: CardPriority.Urgent, label: 'Urgent' },
  ];

  const priorityBadgeStyles: Record<CardPriority, { bg: string; text: string; label: string }> = {
    [CardPriority.None]: { bg: 'bg-gray-100', text: 'text-gray-500', label: 'None' },
    [CardPriority.Low]: { bg: 'bg-blue-100', text: 'text-blue-700', label: 'Low' },
    [CardPriority.Medium]: { bg: 'bg-yellow-100', text: 'text-yellow-700', label: 'Medium' },
    [CardPriority.High]: { bg: 'bg-orange-100', text: 'text-orange-700', label: 'High' },
    [CardPriority.Urgent]: { bg: 'bg-red-100', text: 'text-red-700', label: 'Urgent' },
  };

  interface Props {
    title: string;
    description: string;
    priority: CardPriority;
    dueDate: string;
    storyPoints: number | null;
    selectedTagIds: string[];
    projectId: string;
    tags: Tag[];
    onTitleChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onPriorityChange: (value: CardPriority) => void;
    onDueDateChange: (value: string) => void;
    onStoryPointsChange: (value: number | null) => void;
    onTagSelectionChange: (ids: string[]) => void;
    onTagsChanged?: () => void;
    error?: string | null;
    disabled?: boolean;
    readOnly?: boolean;
    descriptionRows?: number;
    idPrefix?: string;
  }

  let {
    title,
    description,
    priority,
    dueDate,
    storyPoints,
    selectedTagIds,
    projectId,
    tags,
    onTitleChange,
    onDescriptionChange,
    onPriorityChange,
    onDueDateChange,
    onStoryPointsChange,
    onTagSelectionChange,
    onTagsChanged,
    error = null,
    disabled = false,
    readOnly = false,
    descriptionRows = 3,
    idPrefix = ''
  }: Props = $props();


  // Computed values for read-only display
  const selectedTags = $derived(tags.filter(t => selectedTagIds.includes(t.id)));

  function formatDisplayDate(dateStr: string): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
</script>

<div class="space-y-4">
  {#if error && !readOnly}
    <div class="rounded-md bg-red-50 p-4">
      <p class="text-sm text-red-700">{error}</p>
    </div>
  {/if}

  <Input
    id="{idPrefix}title"
    label="Title"
    value={title}
    onInput={(e) => onTitleChange(e.currentTarget.value)}
    placeholder="Enter card title"
    required
    {disabled}
    {readOnly}
    class="font-semibold text-base"
  />

  <RichTextEditor
    label="Description"
    value={description}
    placeholder="Add a description..."
    {disabled}
    {readOnly}
    onUpdate={onDescriptionChange}
  />

  <div class="grid grid-cols-3 gap-4">
    <BitsSelect
      id="{idPrefix}priority"
      label="Priority"
      options={priorityOptions}
      value={priority}
      onValueChange={onPriorityChange}
      placeholder="Select priority..."
      disabled={disabled}
      {readOnly}
    />

    <DatePicker
      id="{idPrefix}dueDate"
      label="Due Date"
      value={dueDate}
      onValueChange={onDueDateChange}
      disabled={disabled}
      {readOnly}
    />

    <div>
      <label for="{idPrefix}storyPoints" class="block text-sm font-medium text-gray-700 mb-1">
        Story Points
      </label>
      {#if readOnly}
        <div class="py-2 px-3 bg-gray-50 rounded-md text-sm text-gray-700 min-h-[38px] flex items-center">
          {storyPoints ?? '-'}
        </div>
      {:else}
        <input
          type="number"
          id="{idPrefix}storyPoints"
          value={storyPoints ?? ''}
          oninput={(e) => {
            const val = e.currentTarget.value;
            onStoryPointsChange(val ? parseInt(val, 10) : null);
          }}
          min="0"
          max="100"
          placeholder="0"
          {disabled}
          class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500 disabled:bg-gray-50 disabled:text-gray-500"
        />
      {/if}
    </div>
  </div>

  <TagPicker
    {projectId}
    {tags}
    {selectedTagIds}
    onSelectionChange={onTagSelectionChange}
    {onTagsChanged}
    {disabled}
    {readOnly}
  />
</div>
