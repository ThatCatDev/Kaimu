<script lang="ts">
  import type { HTMLTextareaAttributes } from 'svelte/elements';

  interface Props extends HTMLTextareaAttributes {
    label?: string;
    error?: string | null;
    hint?: string;
    value?: string;
    readOnly?: boolean;
  }

  let {
    label,
    error,
    hint,
    id,
    rows = 3,
    class: className = '',
    value = $bindable(''),
    readOnly = false,
    ...rest
  }: Props = $props();

  const textareaId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);
</script>

{#if label}
  <div class="w-full">
    <label for={textareaId} class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
      {label}
      {#if rest.required && !readOnly}
        <span class="text-red-500 ml-0.5">*</span>
      {/if}
    </label>
    {#if readOnly}
      <p class="text-sm text-gray-900 whitespace-pre-wrap">{value || '—'}</p>
    {:else}
      <textarea
        id={textareaId}
        {rows}
        bind:value
        class="block w-full px-2 py-1.5 bg-transparent border-0 rounded text-sm text-gray-900 placeholder-gray-400 hover:bg-gray-50 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-indigo-500 transition-all resize-none {error ? 'ring-1 ring-red-500 bg-red-50' : ''} {className}"
        {...rest}
      ></textarea>
      {#if hint && !error}
        <p class="mt-1.5 text-xs text-gray-500">{hint}</p>
      {/if}
      {#if error}
        <p class="mt-1.5 text-xs text-red-600">{error}</p>
      {/if}
    {/if}
  </div>
{:else}
  {#if readOnly}
    <p class="text-sm text-gray-900 whitespace-pre-wrap">{value || '—'}</p>
  {:else}
    <textarea
      {id}
      {rows}
      bind:value
      class="block w-full px-2 py-1.5 bg-transparent border-0 rounded text-sm text-gray-900 placeholder-gray-400 hover:bg-gray-50 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-indigo-500 transition-all resize-none {error ? 'ring-1 ring-red-500 bg-red-50' : ''} {className}"
      {...rest}
    ></textarea>
  {/if}
{/if}
