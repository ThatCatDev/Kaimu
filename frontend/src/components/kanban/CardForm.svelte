<script lang="ts">
  import { CardPriority } from '../../lib/graphql/generated';
  import { Input, Textarea, Select } from '../ui';
  import LabelPicker from './LabelPicker.svelte';
  import type { Label } from '../../lib/api/boards';

  interface Props {
    title: string;
    description: string;
    priority: CardPriority;
    dueDate: string;
    selectedLabelIds: string[];
    projectId: string;
    labels: Label[];
    onTitleChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onPriorityChange: (value: CardPriority) => void;
    onDueDateChange: (value: string) => void;
    onLabelSelectionChange: (ids: string[]) => void;
    onLabelsChanged?: () => void;
    error?: string | null;
    disabled?: boolean;
    descriptionRows?: number;
    idPrefix?: string;
  }

  let {
    title,
    description,
    priority,
    dueDate,
    selectedLabelIds,
    projectId,
    labels,
    onTitleChange,
    onDescriptionChange,
    onPriorityChange,
    onDueDateChange,
    onLabelSelectionChange,
    onLabelsChanged,
    error = null,
    disabled = false,
    descriptionRows = 3,
    idPrefix = ''
  }: Props = $props();

  // Local reactive bindings that call parent callbacks
  let localTitle = $state(title);
  let localDescription = $state(description);
  let localPriority = $state(priority);
  let localDueDate = $state(dueDate);

  // Sync from parent
  $effect(() => { localTitle = title; });
  $effect(() => { localDescription = description; });
  $effect(() => { localPriority = priority; });
  $effect(() => { localDueDate = dueDate; });

  // Notify parent on changes
  $effect(() => {
    if (localTitle !== title) onTitleChange(localTitle);
  });
  $effect(() => {
    if (localDescription !== description) onDescriptionChange(localDescription);
  });
  $effect(() => {
    if (localPriority !== priority) onPriorityChange(localPriority);
  });
  $effect(() => {
    if (localDueDate !== dueDate) onDueDateChange(localDueDate);
  });
</script>

<div class="space-y-4">
  {#if error}
    <div class="rounded-md bg-red-50 p-3">
      <p class="text-sm text-red-700">{error}</p>
    </div>
  {/if}

  <Input
    id="{idPrefix}title"
    label="Title"
    bind:value={localTitle}
    placeholder="Enter card title"
    required
    {disabled}
  />

  <Textarea
    id="{idPrefix}description"
    label="Description"
    bind:value={localDescription}
    rows={descriptionRows}
    placeholder="Add a description"
    {disabled}
  />

  <div class="grid grid-cols-2 gap-4">
    <Select id="{idPrefix}priority" label="Priority" bind:value={localPriority} {disabled}>
      <option value={CardPriority.None}>None</option>
      <option value={CardPriority.Low}>Low</option>
      <option value={CardPriority.Medium}>Medium</option>
      <option value={CardPriority.High}>High</option>
      <option value={CardPriority.Urgent}>Urgent</option>
    </Select>

    <Input
      type="date"
      id="{idPrefix}dueDate"
      label="Due Date"
      bind:value={localDueDate}
      {disabled}
    />
  </div>

  <LabelPicker
    {projectId}
    {labels}
    {selectedLabelIds}
    onSelectionChange={onLabelSelectionChange}
    {onLabelsChanged}
    {disabled}
  />
</div>
