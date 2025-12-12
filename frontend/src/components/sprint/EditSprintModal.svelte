<script lang="ts">
  import { Modal, Button, Input, Textarea } from '../ui';
  import { updateSprint, type SprintData } from '../../lib/api/sprints';
  import { toast } from 'svelte-sonner';

  interface Props {
    open: boolean;
    sprint: SprintData | null;
    onClose: () => void;
    onUpdated: (sprint: SprintData) => void;
  }

  let { open, sprint, onClose, onUpdated }: Props = $props();

  let name = $state('');
  let goal = $state('');
  let startDate = $state('');
  let endDate = $state('');
  let loading = $state(false);
  let error = $state<string | null>(null);

  // Convert RFC3339 date string to YYYY-MM-DD format for input
  function toDateInput(dateStr: string | null | undefined): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toISOString().split('T')[0];
  }

  // Convert date string (YYYY-MM-DD) to RFC3339 format
  function toRFC3339(dateStr: string): string | undefined {
    if (!dateStr) return undefined;
    return `${dateStr}T00:00:00Z`;
  }

  // Reset form when modal opens with sprint data
  $effect(() => {
    if (open && sprint) {
      name = sprint.name;
      goal = sprint.goal ?? '';
      startDate = toDateInput(sprint.startDate);
      endDate = toDateInput(sprint.endDate);
      error = null;
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!sprint || !name.trim()) return;

    loading = true;
    error = null;
    try {
      const updated = await updateSprint(sprint.id, {
        name: name.trim(),
        goal: goal.trim() || undefined,
        startDate: toRFC3339(startDate),
        endDate: toRFC3339(endDate),
      });
      toast.success('Sprint updated');
      onUpdated(updated);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to update sprint';
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

<Modal {open} onOpenChange={handleOpenChange} title="Edit Sprint" size="md">
  {#snippet children()}
    <form id="edit-sprint-form" onsubmit={handleSubmit}>
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
      <Button type="submit" form="edit-sprint-form" disabled={loading || !name.trim()}>
        {loading ? 'Saving...' : 'Save Changes'}
      </Button>
    </div>
  {/snippet}
</Modal>
