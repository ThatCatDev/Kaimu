<script lang="ts">
  import { Modal, Button, Input, Textarea } from '../ui';
  import { createSprint, getSprints } from '../../lib/api/sprints';
  import { toast } from 'svelte-sonner';

  interface Props {
    open: boolean;
    boardId: string;
    onClose: () => void;
    onCreated: () => void;
  }

  let { open, boardId, onClose, onCreated }: Props = $props();

  let name = $state('');
  let goal = $state('');
  let startDate = $state('');
  let endDate = $state('');
  let loading = $state(false);
  let error = $state<string | null>(null);

  // Try to guess the next sprint name from the previous sprint
  function guessNextSprintName(previousName: string): string {
    // Try to find a number at the end of the name
    // e.g., "Sprint 1" -> "Sprint 2", "Week 10" -> "Week 11"
    const endNumberMatch = previousName.match(/^(.+?)(\d+)$/);
    if (endNumberMatch) {
      const prefix = endNumberMatch[1];
      const number = parseInt(endNumberMatch[2], 10);
      return `${prefix}${number + 1}`;
    }

    // Try to find a number anywhere in the name
    // e.g., "Sprint 1 - Q4" -> "Sprint 2 - Q4"
    const anyNumberMatch = previousName.match(/^(.+?)(\d+)(.*)$/);
    if (anyNumberMatch) {
      const prefix = anyNumberMatch[1];
      const number = parseInt(anyNumberMatch[2], 10);
      const suffix = anyNumberMatch[3];
      return `${prefix}${number + 1}${suffix}`;
    }

    // No number found, just append " 2"
    return `${previousName} 2`;
  }

  // Reset form when modal opens and suggest next sprint name
  $effect(() => {
    if (open) {
      goal = '';
      startDate = '';
      endDate = '';
      error = null;

      // Fetch sprints to suggest next name
      loadSuggestedName();
    }
  });

  async function loadSuggestedName() {
    try {
      const sprints = await getSprints(boardId);
      if (sprints.length > 0) {
        // Sort by position (most recent/highest position first) or createdAt
        const sortedSprints = [...sprints].sort((a, b) => b.position - a.position);
        const lastSprint = sortedSprints[0];
        name = guessNextSprintName(lastSprint.name);
      } else {
        // No sprints yet, suggest "Sprint 1"
        name = 'Sprint 1';
      }
    } catch {
      // If we can't fetch sprints, just leave the name empty
      name = '';
    }
  }

  // Convert date string (YYYY-MM-DD) to RFC3339 format
  function toRFC3339(dateStr: string): string | undefined {
    if (!dateStr) return undefined;
    return `${dateStr}T00:00:00Z`;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!name.trim()) return;

    loading = true;
    error = null;
    try {
      await createSprint({
        boardId,
        name: name.trim(),
        goal: goal.trim() || undefined,
        startDate: toRFC3339(startDate),
        endDate: toRFC3339(endDate),
      });
      toast.success('Sprint created');
      onCreated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create sprint';
    } finally {
      loading = false;
    }
  }

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      onClose();
    }
  }
</script>

<Modal {open} onOpenChange={handleOpenChange} title="Create Sprint" size="md">
  {#snippet children()}
    <form id="create-sprint-form" onsubmit={handleSubmit}>
      <div class="px-6 py-4 space-y-4">
        {#if error}
          <div class="rounded-md bg-red-50 p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}

        <Input
          label="Sprint Name"
          bind:value={name}
          placeholder="e.g., Sprint 1"
          required
          disabled={loading}
        />

        <Textarea
          label="Goal"
          bind:value={goal}
          placeholder="What do you want to achieve in this sprint?"
          rows={3}
          disabled={loading}
          hint="Optional - describe the sprint objective"
        />

        <div class="grid grid-cols-2 gap-4">
          <Input
            label="Start Date"
            type="date"
            bind:value={startDate}
            disabled={loading}
          />
          <Input
            label="End Date"
            type="date"
            bind:value={endDate}
            disabled={loading}
          />
        </div>
      </div>
    </form>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button type="button" variant="secondary" onclick={onClose} disabled={loading}>
        Cancel
      </Button>
      <Button type="submit" form="create-sprint-form" disabled={loading || !name.trim()}>
        {loading ? 'Creating...' : 'Create Sprint'}
      </Button>
    </div>
  {/snippet}
</Modal>
