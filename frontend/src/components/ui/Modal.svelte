<script lang="ts">
  import { Dialog } from 'bits-ui';
  import { fly, fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import type { Snippet } from 'svelte';

  interface Props {
    open: boolean;
    onOpenChange?: (open: boolean) => void;
    title?: string;
    description?: string;
    size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
    children: Snippet;
    footer?: Snippet;
    headerActions?: Snippet;
  }

  let {
    open = $bindable(false),
    onOpenChange,
    title,
    description,
    size = 'md',
    children,
    footer,
    headerActions
  }: Props = $props();

  const sizeClasses: Record<string, string> = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    '2xl': 'max-w-2xl'
  };

  function handleOpenChange(newOpen: boolean) {
    open = newOpen;
    onOpenChange?.(newOpen);
  }
</script>

<Dialog.Root {open} onOpenChange={handleOpenChange}>
  <Dialog.Portal>
    <Dialog.Overlay forceMount>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
            transition:fade={{ duration: 150 }}
          ></div>
        {/if}
      {/snippet}
    </Dialog.Overlay>

    <Dialog.Content forceMount>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            class="fixed inset-0 z-50 flex items-center justify-center p-4"
            transition:fade={{ duration: 150 }}
          >
            <div
              class="bg-white rounded-xl shadow-2xl w-full {sizeClasses[size]} max-h-[90vh] flex flex-col"
              transition:fly={{ y: 10, duration: 200, easing: cubicOut }}
              onclick={(e) => e.stopPropagation()}
            >
              {#if title}
                <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between flex-shrink-0">
                  <div>
                    <Dialog.Title class="text-lg font-semibold text-gray-900">{title}</Dialog.Title>
                    {#if description}
                      <Dialog.Description class="text-sm text-gray-500 mt-1">{description}</Dialog.Description>
                    {/if}
                  </div>
                  <div class="flex items-center gap-1">
                    {#if headerActions}
                      {@render headerActions()}
                    {/if}
                    <Dialog.Close class="text-gray-400 hover:text-gray-600 transition-colors p-1.5 hover:bg-gray-100 rounded-md">
                      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </Dialog.Close>
                  </div>
                </div>
              {/if}

              <div class="flex-1 overflow-y-auto">
                {@render children()}
              </div>

              {#if footer}
                <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 rounded-b-xl flex-shrink-0">
                  {@render footer()}
                </div>
              {/if}
            </div>
          </div>
        {/if}
      {/snippet}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
