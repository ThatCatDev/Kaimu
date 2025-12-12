<script lang="ts">
  import { Modal, Button, Input, Textarea } from '../ui';
  import { createSprint } from '../../lib/api/sprints';
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

  // Reset form when modal opens
  $effect(() => {
    if (open) {
      name = '';
      goal = '';
      startDate = '';
      endDate = '';
      error = null;
    }
  });

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
