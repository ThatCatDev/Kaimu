<script lang="ts">
  interface Props {
    id: string;
    name: string;
    slug: string;
    description?: string | null;
    projectCount?: number;
  }

  let { id, name, slug, description, projectCount = 0 }: Props = $props();
</script>

<a
  href={`/organizations/${id}`}
  class="block bg-white rounded-lg shadow hover:shadow-md transition-shadow duration-200 p-6"
>
  <div class="flex items-start justify-between">
    <div class="flex-1 min-w-0">
      <h3 class="text-lg font-semibold text-gray-900 truncate">{name}</h3>
      <p class="text-sm text-gray-500">/{slug}</p>
    </div>
  </div>
  {#if description}
    <p class="mt-2 text-sm text-gray-600 line-clamp-2">{description}</p>
  {/if}
  <div class="mt-4 flex items-center text-sm text-gray-500">
    <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
    </svg>
    {projectCount} {projectCount === 1 ? 'project' : 'projects'}
  </div>
</a>
