<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  interface Props extends Omit<HTMLInputAttributes, 'readonly'> {
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
    class: className = '',
    value = $bindable(''),
    readOnly = false,
    ...rest
  }: Props = $props();

  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);
</script>

{#if label}
  <div class="w-full">
    <label for={inputId} class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
      {label}
      {#if rest.required && !readOnly}
        <span class="text-red-500">*</span>
      {/if}
    </label>
    {#if readOnly}
      <p class="text-sm text-gray-900">{value || '—'}</p>
    {:else}
      <input
        {id}
        bind:value
        class="block w-full px-4 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
        {...rest}
      />
      {#if hint && !error}
        <p class="mt-1 text-xs text-gray-500">{hint}</p>
      {/if}
      {#if error}
        <p class="mt-1 text-xs text-red-600">{error}</p>
      {/if}
    {/if}
  </div>
{:else}
  {#if readOnly}
    <span class="text-sm text-gray-900">{value || '—'}</span>
  {:else}
    <input
      {id}
      bind:value
      class="block w-full px-4 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm transition-colors {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300'} {className}"
      {...rest}
    />
  {/if}
{/if}
