<script lang="ts">
  import { createOrganization } from '../lib/api/organizations';
  import { Input, Textarea, Button } from './ui';

  let name = $state('');
  let description = $state('');
  let error = $state<string | null>(null);
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = null;

    if (!name.trim()) {
      error = 'Organization name is required';
      return;
    }

    loading = true;
    try {
      const org = await createOrganization(name.trim(), description.trim() || undefined);
      window.location.href = `/organizations/${org.id}`;
    } catch (e) {
      error = e instanceof Error ? e.message : 'An error occurred';
    } finally {
      loading = false;
    }
  }
</script>

<div class="max-w-lg mx-auto">
  <form class="space-y-6" onsubmit={handleSubmit}>
    {#if error}
      <div class="rounded-md bg-red-50 p-4">
        <p class="text-sm text-red-700">{error}</p>
      </div>
    {/if}

    <Input
      id="name"
      label="Organization Name"
      bind:value={name}
      placeholder="My Organization"
      required
    />

    <Textarea
      id="description"
      label="Description"
      bind:value={description}
      placeholder="A brief description of your organization"
      rows={3}
    />

    <div class="flex gap-4">
      <a
        href="/dashboard"
        class="flex-1 py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 text-center"
      >
        Cancel
      </a>
      <Button type="submit" {loading} class="flex-1">
        {loading ? 'Creating...' : 'Create Organization'}
      </Button>
    </div>
  </form>
</div>
