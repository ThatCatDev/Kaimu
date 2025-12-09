<script lang="ts">
  import Button from './Button.svelte';

  interface Props {
    isOpen: boolean;
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    variant?: 'danger' | 'primary';
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    isOpen,
    title,
    message,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    variant = 'danger',
    onConfirm,
    onCancel
  }: Props = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (!isOpen) return;
    if (e.key === 'Escape') {
      onCancel();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onCancel();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div
    class="fixed inset-0 bg-gray-900/60 backdrop-blur-sm z-[60] animate-fade-in flex items-center justify-center p-4"
    onclick={handleBackdropClick}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-white rounded-xl shadow-2xl max-w-md w-full animate-scale-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="px-6 py-5">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
        <p class="text-sm text-gray-600">{message}</p>
      </div>

      <div class="px-6 py-4 bg-gray-50 rounded-b-xl flex justify-end gap-3">
        <Button variant="secondary" onclick={onCancel}>
          {cancelText}
        </Button>
        <Button variant={variant} onclick={onConfirm}>
          {confirmText}
        </Button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scale-in {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  :global(.animate-fade-in) {
    animation: fade-in 0.15s ease-out;
  }

  :global(.animate-scale-in) {
    animation: scale-in 0.2s ease-out;
  }
</style>
