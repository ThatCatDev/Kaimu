<script lang="ts">
  import { createProject } from '../lib/api/projects';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  let name = $state('');
  let key = $state('');
  let description = $state('');
  let error = $state<string | null>(null);
  let loading = $state(false);

  function generateKey(projectName: string): string {
    return projectName
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, '')
      .slice(0, 6);
  }

  function handleNameChange(e: Event) {
    const target = e.target as HTMLInputElement;
    name = target.value;
    if (!key || key === generateKey(name.slice(0, -1))) {
      key = generateKey(name);
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = null;

    if (!name.trim()) {
      error = 'Project name is required';
      return;
    }

    if (!key.trim()) {
      error = 'Project key is required';
      return;
    }

    if (key.length > 10) {
      error = 'Project key must be 10 characters or less';
      return;
    }

    loading = true;
    try {
      const project = await createProject(
        organizationId,
        name.trim(),
        key.trim().toUpperCase(),
        description.trim() || undefined
      );
      window.location.href = `/projects/${project.id}`;
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
        Project Name
      </label>
      <input
        id="name"
        name="name"
        type="text"
        required
        value={name}
        oninput={handleNameChange}
        class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
        placeholder="My Project"
      />
    </div>

    <div>
      <label for="key" class="block text-sm font-medium text-gray-700">
        Project Key
      </label>
      <input
        id="key"
        name="key"
        type="text"
        required
        maxlength="10"
        bind:value={key}
        class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm uppercase"
        placeholder="PROJ"
      />
      <p class="mt-1 text-xs text-gray-500">
        A short identifier for your project (max 10 characters)
      </p>
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
        placeholder="A brief description of your project"
      ></textarea>
    </div>

    <div class="flex gap-4">
      <a
        href={`/organizations/${organizationId}`}
        class="flex-1 py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 text-center"
      >
        Cancel
      </a>
      <button
        type="submit"
        disabled={loading}
        class="flex-1 py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? 'Creating...' : 'Create Project'}
      </button>
    </div>
  </form>
</div>
