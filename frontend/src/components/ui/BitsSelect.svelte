<script lang="ts">
  import { Select } from 'bits-ui';

  interface SelectOption {
    value: string;
    label: string;
    disabled?: boolean;
  }

  interface Props {
    options: SelectOption[];
    value?: string;
    placeholder?: string;
    label?: string;
    error?: string | null;
    disabled?: boolean;
    required?: boolean;
    id?: string;
  }

  let {
    options,
    value = $bindable(''),
    placeholder = 'Select an option...',
    label,
    error,
    disabled = false,
    required = false,
    id
  }: Props = $props();

  const selectedLabel = $derived(
    value ? options.find(o => o.value === value)?.label : placeholder
  );
</script>

<div class="w-full">
  {#if label}
    <label for={id} class="block text-sm font-medium text-gray-700 mb-1">
      {label}
      {#if required}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  <Select.Root type="single" bind:value {disabled}>
    <Select.Trigger
      {id}
      class="flex h-10 w-full items-center justify-between rounded-md border bg-white px-4 py-2 text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 disabled:cursor-not-allowed disabled:opacity-50 {error ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 hover:border-gray-400'}"
    >
      <span class={value ? 'text-gray-900' : 'text-gray-500'}>
        {selectedLabel}
      </span>
      <svg class="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </Select.Trigger>

    <Select.Portal>
      <Select.Content
        class="z-50 w-[var(--bits-select-anchor-width)] rounded-md border border-gray-200 bg-white shadow-lg animate-in fade-in-0 zoom-in-95"
        sideOffset={4}
        avoidCollisions={true}
        collisionPadding={16}
      >
        <Select.Viewport class="p-1 max-h-60 overflow-y-auto">
          {#each options as option (option.value)}
            <Select.Item
              value={option.value}
              label={option.label}
              disabled={option.disabled}
              class="relative flex w-full cursor-pointer select-none items-center rounded-md px-4 py-2 text-sm text-gray-900 outline-none transition-colors data-[highlighted]:bg-indigo-50 data-[highlighted]:text-indigo-900 data-[selected]:bg-indigo-100 data-[disabled]:pointer-events-none data-[disabled]:opacity-50"
            >
              {#snippet children({ selected })}
                <span class="flex-1">{option.label}</span>
                {#if selected}
                  <svg class="h-4 w-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                {/if}
              {/snippet}
            </Select.Item>
          {/each}
        </Select.Viewport>
      </Select.Content>
    </Select.Portal>
  </Select.Root>

  {#if error}
    <p class="mt-1 text-xs text-red-600">{error}</p>
  {/if}
</div>

