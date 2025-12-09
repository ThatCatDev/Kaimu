<script lang="ts">
  interface Props {
    value: string;
    onSave: (newValue: string) => Promise<void>;
    placeholder?: string;
    class?: string;
  }

  let { value, onSave, placeholder = 'Enter name...', class: className = '' }: Props = $props();

  let editing = $state(false);
  let editValue = $state(value);
  let saving = $state(false);
  let inputEl: HTMLInputElement;

  function startEditing() {
    editValue = value;
    editing = true;
    setTimeout(() => inputEl?.focus(), 0);
  }

  async function save() {
    if (editValue.trim() === '' || editValue === value) {
      editing = false;
      return;
    }

    try {
      saving = true;
      await onSave(editValue.trim());
      editing = false;
    } catch (e) {
      // Keep editing mode open on error
      console.error('Failed to save:', e);
    } finally {
      saving = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      save();
    } else if (e.key === 'Escape') {
      editing = false;
      editValue = value;
    }
  }

  function handleBlur() {
    if (!saving) {
      save();
    }
  }
</script>

{#if editing}
  <input
    bind:this={inputEl}
    bind:value={editValue}
    onkeydown={handleKeydown}
    onblur={handleBlur}
    disabled={saving}
    {placeholder}
    class="bg-transparent border-b-2 border-indigo-500 outline-none py-0.5 {className}"
  />
{:else}
  <button
    type="button"
    onclick={startEditing}
    class="group inline-flex items-center gap-2 hover:bg-gray-100 rounded px-1 -mx-1 py-0.5 transition-colors {className}"
    title="Click to edit"
  >
    <span>{value}</span>
    <svg class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
    </svg>
  </button>
{/if}
