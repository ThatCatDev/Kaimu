<script lang="ts">
  import type { HTMLTextareaAttributes } from 'svelte/elements';

  interface Props extends HTMLTextareaAttributes {
    label?: string;
    error?: string | null;
    hint?: string;
    value?: string;
  }

  let {
    label,
    error,
    hint,
    id,
    rows = 3,
    class: className = '',
    value = $bindable(''),
    ...rest
  }: Props = $props();

  const textareaId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);
</script>

{#if label}
  <div class="w-full">
    <label for={textareaId} class="block text-sm font-medium text-gray-700 mb-1">
      {label}
      {#if rest.required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
    <textarea
      id={textareaId}
      {rows}
      bind:value
      class="block w-full px-3 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
      {...rest}
    ></textarea>
    {#if hint && !error}
      <p class="mt-1 text-xs text-gray-500">{hint}</p>
    {/if}
    {#if error}
      <p class="mt-1 text-xs text-red-600">{error}</p>
    {/if}
  </div>
{:else}
  <textarea
    {id}
    {rows}
    bind:value
    class="block w-full px-3 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
    {...rest}
  ></textarea>
{/if}
