<script lang="ts">
  import Modal from './Modal.svelte';
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

  function handleOpenChange(open: boolean) {
    if (!open) {
      onCancel();
    }
  }
</script>

<Modal open={isOpen} onOpenChange={handleOpenChange} {title} size="sm">
  {#snippet children()}
    <div class="px-6 py-4">
      <p class="text-sm text-gray-600">{message}</p>
    </div>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button variant="secondary" onclick={onCancel}>
        {cancelText}
      </Button>
      <Button variant={variant} onclick={onConfirm}>
        {confirmText}
      </Button>
    </div>
  {/snippet}
</Modal>
