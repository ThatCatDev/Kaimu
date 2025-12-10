<script lang="ts">
  import Modal from '../ui/Modal.svelte';
  import Input from '../ui/Input.svelte';
  import Button from '../ui/Button.svelte';
  import { updateMe } from '../../lib/api/auth';
  import type { User } from '../../lib/graphql/generated';

  interface Props {
    open: boolean;
    onOpenChange?: (open: boolean) => void;
    user: User;
    onUpdate?: (user: User) => void;
  }

  let { open = $bindable(false), onOpenChange, user, onUpdate }: Props = $props();

  let displayName = $state(user.displayName || '');
  let email = $state(user.email || '');
  let error = $state<string | null>(null);
  let saving = $state(false);

  $effect(() => {
    if (open) {
      displayName = user.displayName || '';
      email = user.email || '';
      error = null;
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = null;
    saving = true;

    try {
      const updated = await updateMe({
        displayName: displayName || null,
        email: email || null,
      });
      onUpdate?.(updated);
      open = false;
      onOpenChange?.(false);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to update profile';
    } finally {
      saving = false;
    }
  }

  function handleCancel() {
    open = false;
    onOpenChange?.(false);
  }
</script>

<Modal bind:open {onOpenChange} title="Account Settings" description="Update your profile information" size="md">
  <form onsubmit={handleSubmit} class="p-6 space-y-4">
    {#if error}
      <div class="p-3 rounded-md bg-red-50 border border-red-200">
        <p class="text-sm text-red-600">{error}</p>
      </div>
    {/if}

    <div>
      <Input
        label="Username"
        value={user.username}
        disabled
        hint="Username cannot be changed"
      />
    </div>

    <div>
      <Input
        label="Display Name"
        bind:value={displayName}
        placeholder="Enter your display name"
        hint="This is how your name will appear to others"
      />
    </div>

    <div>
      <Input
        label="Email"
        type="email"
        bind:value={email}
        placeholder="Enter your email address"
      />
    </div>

    <div class="flex justify-end gap-3 pt-4">
      <Button type="button" variant="secondary" onclick={handleCancel}>
        Cancel
      </Button>
      <Button type="submit" variant="primary" loading={saving}>
        {saving ? 'Saving...' : 'Save Changes'}
      </Button>
    </div>
  </form>
</Modal>
