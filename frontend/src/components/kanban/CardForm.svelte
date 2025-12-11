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
    selectedTagIds: string[];
    projectId: string;
    tags: Tag[];
    onTitleChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onPriorityChange: (value: CardPriority) => void;
    onDueDateChange: (value: string) => void;
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
    selectedTagIds,
    projectId,
    tags,
    onTitleChange,
    onDescriptionChange,
    onPriorityChange,
    onDueDateChange,
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

  <div class="grid grid-cols-2 gap-4">
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
