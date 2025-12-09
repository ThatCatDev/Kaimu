<script lang="ts">
  import { createProject } from '../lib/api/projects';
  import { Input, Textarea, Button } from './ui';

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

    <Input
      id="name"
      label="Project Name"
      value={name}
      oninput={handleNameChange}
      placeholder="My Project"
      required
    />

    <Input
      id="key"
      label="Project Key"
      bind:value={key}
      placeholder="PROJ"
      maxlength={10}
      class="uppercase"
      hint="A short identifier for your project (max 10 characters)"
      required
    />

    <Textarea
      id="description"
      label="Description"
      bind:value={description}
      placeholder="A brief description of your project"
      rows={3}
    />

    <div class="flex gap-4">
      <a
        href={`/organizations/${organizationId}`}
        class="flex-1 py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 text-center"
      >
        Cancel
      </a>
      <Button type="submit" {loading} class="flex-1">
        {loading ? 'Creating...' : 'Create Project'}
      </Button>
    </div>
  </form>
</div>
