<script lang="ts">
  import { createOrganization } from '../lib/api/organizations';

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

    <div>
      <label for="name" class="block text-sm font-medium text-gray-700">
        Organization Name
      </label>
      <input
        id="name"
        name="name"
        type="text"
        required
        bind:value={name}
        class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
        placeholder="My Organization"
      />
    </div>

    <div>
      <label for="description" class="block text-sm font-medium text-gray-700">
        Description <span class="text-gray-400">(optional)</span>
      </label>
      <textarea
        id="description"
        name="description"
        rows="3"
        bind:value={description}
        class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
        placeholder="A brief description of your organization"
      ></textarea>
    </div>

    <div class="flex gap-4">
      <a
        href="/dashboard"
        class="flex-1 py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 text-center"
      >
        Cancel
      </a>
      <button
        type="submit"
        disabled={loading}
        class="flex-1 py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? 'Creating...' : 'Create Organization'}
      </button>
    </div>
  </form>
</div>
