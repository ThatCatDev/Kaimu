<script lang="ts">
  import type { HTMLSelectAttributes } from 'svelte/elements';
  import type { Snippet } from 'svelte';

  interface Props extends HTMLSelectAttributes {
    label?: string;
    error?: string | null;
    hint?: string;
    value?: string;
    children: Snippet;
  }

  let {
    label,
    error,
    hint,
    id,
    children,
    class: className = '',
    value = $bindable(''),
    ...rest
  }: Props = $props();

  const selectId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);
</script>

{#if label}
  <div class="w-full">
    <label for={selectId} class="block text-sm font-medium text-gray-700 mb-1">
      {label}
      {#if rest.required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
    <select
      id={selectId}
      bind:value
      class="block w-full px-4 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
      {...rest}
    >
      {@render children()}
    </select>
    {#if hint && !error}
      <p class="mt-1 text-xs text-gray-500">{hint}</p>
    {/if}
    {#if error}
      <p class="mt-1 text-xs text-red-600">{error}</p>
    {/if}
  </div>
{:else}
  <select
    {id}
    bind:value
    class="block w-full px-4 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
    {...rest}
  >
    {@render children()}
  </select>
{/if}
