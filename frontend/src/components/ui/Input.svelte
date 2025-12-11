<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  interface Props extends Omit<HTMLInputAttributes, 'readonly'> {
    label?: string;
    error?: string | null;
    hint?: string;
    value?: string;
    readOnly?: boolean;
    onInput?: (e: Event & { currentTarget: HTMLInputElement }) => void;
  }

  let {
    label,
    error,
    hint,
    id,
    class: className = '',
    value = $bindable(''),
    readOnly = false,
    onInput,
    ...rest
  }: Props = $props();

  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  // Handle input - update bindable value and call callback if provided
  function handleInput(e: Event & { currentTarget: HTMLInputElement }) {
    value = e.currentTarget.value;
    onInput?.(e);
  }
</script>

{#if label}
  <div class="w-full">
    <label for={inputId} class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
      {label}
      {#if rest.required && !readOnly}
        <span class="text-red-500 ml-0.5">*</span>
      {/if}
    </label>
    {#if readOnly}
      <p class="text-sm text-gray-900">{value || '—'}</p>
    {:else}
      <input
        {id}
        {value}
        oninput={handleInput}
        class="block w-full px-2 py-1.5 bg-transparent border-0 rounded text-sm text-gray-900 placeholder-gray-400 hover:bg-gray-50 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-indigo-500 transition-all {error ? 'ring-1 ring-red-500 bg-red-50' : ''} {className}"
        {...rest}
      />
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
    <span class="text-sm text-gray-900">{value || '—'}</span>
  {:else}
    <input
      {id}
      {value}
      oninput={handleInput}
      class="block w-full px-2 py-1.5 bg-transparent border-0 rounded text-sm text-gray-900 placeholder-gray-400 hover:bg-gray-50 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-indigo-500 transition-all {error ? 'ring-1 ring-red-500 bg-red-50' : ''} {className}"
      {...rest}
    />
  {/if}
{/if}
