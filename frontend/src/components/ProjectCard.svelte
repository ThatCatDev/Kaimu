<script lang="ts">
  interface Props {
    id: string;
    name: string;
    projectKey: string;
    description?: string | null;
  }

  let { id, name, projectKey, description }: Props = $props();
</script>

<a
  href={`/projects/${id}`}
  class="block bg-white rounded-lg shadow hover:shadow-md transition-shadow duration-200 p-6"
>
  <div class="flex items-start justify-between">
    <div class="flex-1 min-w-0">
      <h3 class="text-lg font-semibold text-gray-900 truncate">{name}</h3>
      <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 mt-1">
        {projectKey}
      </span>
    </div>
  </div>
  {#if description}
    <p class="mt-2 text-sm text-gray-600 line-clamp-2">{description}</p>
  {/if}
</a>
