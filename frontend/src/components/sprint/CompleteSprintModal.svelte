<script lang="ts">
  import { Modal, Button, Input } from '../ui';
  import {
    completeSprint,
    createSprint,
    getSprints,
    getFutureSprints,
    getClosedSprints,
    type SprintData,
  } from '../../lib/api/sprints';
  import { toast } from 'svelte-sonner';

  interface Props {
    open: boolean;
    sprint: SprintData | null;
    boardId: string;
    incompleteCardCount: number;
    onClose: () => void;
    onCompleted: () => void;
  }

  let { open, sprint, boardId, incompleteCardCount, onClose, onCompleted }: Props = $props();

  let loading = $state(false);
  let loadingData = $state(true);
  let error = $state<string | null>(null);

  // Sprint data
  let nextSprint = $state<SprintData | null>(null);
  let suggestedSprintName = $state('');
  let suggestedStartDate = $state('');
  let suggestedEndDate = $state('');

  // User choices
  let selectedOption = $state<'next' | 'create' | 'none'>('next');
  let newSprintName = $state('');
  let newSprintStartDate = $state('');
  let newSprintEndDate = $state('');

  // Load data when modal opens
  $effect(() => {
    if (open && sprint) {
      loadSprintData();
    }
  });

  async function loadSprintData() {
    loadingData = true;
    error = null;

    try {
      // Check for existing future sprints
      const futureSprints = await getFutureSprints(boardId);
      nextSprint = futureSprints.length > 0 ? futureSprints[0] : null;

      // Calculate suggested sprint name and dates
      const allSprints = await getSprints(boardId);

      // Guess next sprint name
      if (sprint) {
        suggestedSprintName = guessNextSprintName(sprint.name);
      }

      // Calculate average sprint length from closed sprints
      const closedResult = await getClosedSprints(boardId, 10);
      const avgDays = calculateAverageSprintLength(closedResult.sprints);

      // Set suggested dates
      const today = new Date();
      suggestedStartDate = formatDateForInput(today);

      const endDate = new Date(today);
      endDate.setDate(endDate.getDate() + avgDays);
      suggestedEndDate = formatDateForInput(endDate);

      // Initialize form with suggestions
      newSprintName = suggestedSprintName;
      newSprintStartDate = suggestedStartDate;
      newSprintEndDate = suggestedEndDate;

      // Default selection based on whether next sprint exists
      if (nextSprint) {
        selectedOption = 'next';
      } else if (incompleteCardCount > 0) {
        selectedOption = 'create';
      } else {
        selectedOption = 'none';
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load sprint data';
    } finally {
      loadingData = false;
    }
  }

  function guessNextSprintName(previousName: string): string {
    // Try to find a number at the end of the name
    const endNumberMatch = previousName.match(/^(.+?)(\d+)$/);
    if (endNumberMatch) {
      const prefix = endNumberMatch[1];
      const number = parseInt(endNumberMatch[2], 10);
      return `${prefix}${number + 1}`;
    }

    // Try to find a number anywhere in the name
    const anyNumberMatch = previousName.match(/^(.+?)(\d+)(.*)$/);
    if (anyNumberMatch) {
      const prefix = anyNumberMatch[1];
      const number = parseInt(anyNumberMatch[2], 10);
      const suffix = anyNumberMatch[3];
      return `${prefix}${number + 1}${suffix}`;
    }

    return `${previousName} 2`;
  }

  function calculateAverageSprintLength(sprints: SprintData[]): number {
    const sprintsWithDates = sprints.filter(s => s.startDate && s.endDate);

    if (sprintsWithDates.length === 0) {
      return 14; // Default to 2 weeks
    }

    let totalDays = 0;
    for (const s of sprintsWithDates) {
      const start = new Date(s.startDate!);
      const end = new Date(s.endDate!);
      const days = Math.round((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
      totalDays += days;
    }

    return Math.round(totalDays / sprintsWithDates.length);
  }

  function formatDateForInput(date: Date): string {
    return date.toISOString().split('T')[0];
  }

  function toRFC3339(dateStr: string): string | undefined {
    if (!dateStr) return undefined;
    return `${dateStr}T00:00:00Z`;
  }

  async function handleComplete() {
    if (!sprint) return;

    loading = true;
    error = null;

    try {
      // If user wants to create a new sprint first
      if (selectedOption === 'create' && incompleteCardCount > 0) {
        if (!newSprintName.trim()) {
          error = 'Please enter a name for the new sprint';
          loading = false;
          return;
        }

        // Create the new sprint
        await createSprint({
          boardId,
          name: newSprintName.trim(),
          startDate: toRFC3339(newSprintStartDate),
          endDate: toRFC3339(newSprintEndDate),
        });

        toast.success(`Created ${newSprintName.trim()}`);
      }

      // Complete the sprint
      // moveIncompleteToNextSprint: true if moving to next/new sprint, false if not moving
      const moveToNext = selectedOption !== 'none' && incompleteCardCount > 0;
      await completeSprint(sprint.id, moveToNext);

      if (incompleteCardCount > 0) {
        if (selectedOption === 'next' && nextSprint) {
          toast.success(`Completed ${sprint.name}. ${incompleteCardCount} incomplete card(s) moved to ${nextSprint.name}`);
        } else if (selectedOption === 'create') {
          toast.success(`Completed ${sprint.name}. ${incompleteCardCount} incomplete card(s) moved to ${newSprintName.trim()}`);
        } else {
          toast.success(`Completed ${sprint.name}`);
        }
      } else {
        toast.success(`Completed ${sprint.name}`);
      }

      onCompleted();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to complete sprint';
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

<Modal {open} onOpenChange={handleOpenChange} title="Complete Sprint" size="md">
  {#snippet children()}
    <div class="px-6 py-4 space-y-4">
      {#if error}
        <div class="rounded-md bg-red-50 p-3">
          <p class="text-sm text-red-700">{error}</p>
        </div>
      {/if}

      {#if loadingData}
        <div class="flex items-center justify-center py-8">
          <span class="text-gray-500">Loading...</span>
        </div>
      {:else}
        <div class="space-y-4">
          <!-- Sprint summary -->
          <div class="bg-gray-50 rounded-lg p-4">
            <h4 class="font-medium text-gray-900">{sprint?.name}</h4>
            {#if incompleteCardCount > 0}
              <p class="text-sm text-amber-600 mt-1">
                {incompleteCardCount} incomplete card{incompleteCardCount === 1 ? '' : 's'} (not in done columns)
              </p>
            {:else}
              <p class="text-sm text-green-600 mt-1">
                All cards are complete!
              </p>
            {/if}
          </div>

          {#if incompleteCardCount > 0}
            <div class="space-y-3">
              <p class="text-sm font-medium text-gray-700">What would you like to do with incomplete cards?</p>

              <!-- Option: Move to next sprint (if exists) -->
              {#if nextSprint}
                <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors
                  {selectedOption === 'next' ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:bg-gray-50'}">
                  <input
                    type="radio"
                    name="completeOption"
                    value="next"
                    bind:group={selectedOption}
                    class="mt-0.5"
                  />
                  <div>
                    <span class="text-sm font-medium text-gray-900">Move to {nextSprint.name}</span>
                    <p class="text-xs text-gray-500 mt-0.5">Incomplete cards will be added to the next sprint</p>
                  </div>
                </label>
              {/if}

              <!-- Option: Create new sprint -->
              <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors
                {selectedOption === 'create' ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:bg-gray-50'}">
                <input
                  type="radio"
                  name="completeOption"
                  value="create"
                  bind:group={selectedOption}
                  class="mt-0.5"
                />
                <div class="flex-1">
                  <span class="text-sm font-medium text-gray-900">Create a new sprint</span>
                  <p class="text-xs text-gray-500 mt-0.5">Create a new sprint and move incomplete cards there</p>
                </div>
              </label>

              {#if selectedOption === 'create'}
                <div class="ml-6 space-y-3 pt-2">
                  <Input
                    label="Sprint Name"
                    bind:value={newSprintName}
                    placeholder="e.g., Sprint 2"
                    disabled={loading}
                  />
                  <div class="grid grid-cols-2 gap-3">
                    <Input
                      label="Start Date"
                      type="date"
                      bind:value={newSprintStartDate}
                      disabled={loading}
                    />
                    <Input
                      label="End Date"
                      type="date"
                      bind:value={newSprintEndDate}
                      disabled={loading}
                    />
                  </div>
                </div>
              {/if}

              <!-- Option: Don't move -->
              <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors
                {selectedOption === 'none' ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:bg-gray-50'}">
                <input
                  type="radio"
                  name="completeOption"
                  value="none"
                  bind:group={selectedOption}
                  class="mt-0.5"
                />
                <div>
                  <span class="text-sm font-medium text-gray-900">Don't move cards</span>
                  <p class="text-xs text-gray-500 mt-0.5">Incomplete cards will stay only in this closed sprint</p>
                </div>
              </label>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button type="button" variant="secondary" onclick={onClose} disabled={loading}>
        Cancel
      </Button>
      <Button
        type="button"
        onclick={handleComplete}
        disabled={loading || loadingData || (selectedOption === 'create' && !newSprintName.trim())}
      >
        {loading ? 'Completing...' : 'Complete Sprint'}
      </Button>
    </div>
  {/snippet}
</Modal>
