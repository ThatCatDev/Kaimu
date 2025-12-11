<script lang="ts">
  interface Props {
    value: string | null | undefined;
    onSave: (newValue: string) => Promise<void>;
    placeholder?: string;
    class?: string;
  }

  let {
    value,
    onSave,
    placeholder = 'Add description...',
    class: className = ''
  }: Props = $props();

  let editing = $state(false);
  let editValue = $state(value ?? '');
  let saving = $state(false);
  let textareaEl: HTMLTextAreaElement;

  function startEditing() {
    editValue = value ?? '';
    editing = true;
    setTimeout(() => {
      textareaEl?.focus();
      autoResize();
    }, 0);
  }

  async function save() {
    const trimmed = editValue.trim();
    // Allow saving empty string to clear description, but skip if unchanged
    if (trimmed === (value ?? '')) {
      editing = false;
      return;
    }

    try {
      saving = true;
      await onSave(trimmed);
      editing = false;
    } catch (e) {
      console.error('Failed to save:', e);
    } finally {
      saving = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      editing = false;
      editValue = value ?? '';
    }
  }

  function handleBlur() {
    if (!saving) {
      save();
    }
  }

  function autoResize() {
    if (textareaEl) {
      textareaEl.style.height = 'auto';
      textareaEl.style.height = Math.max(textareaEl.scrollHeight, 24) + 'px';
    }
  }
</script>

{#if editing}
  <textarea
    bind:this={textareaEl}
    bind:value={editValue}
    onkeydown={handleKeydown}
    onblur={handleBlur}
    oninput={autoResize}
    disabled={saving}
    placeholder={placeholder}
    rows="1"
    class="w-full bg-transparent border-b-2 border-indigo-500 outline-none text-sm text-gray-600 resize-none py-0.5 {className}"
  ></textarea>
{:else}
  <button
    type="button"
    onclick={startEditing}
    class="group inline-flex items-center gap-1 text-left hover:bg-gray-100 rounded px-1 -mx-1 py-0.5 transition-colors {className}"
    title="Click to edit"
  >
    {#if value}
      <span class="text-sm text-gray-500 whitespace-pre-wrap">{value}</span>
    {:else}
      <span class="text-sm text-gray-400 italic">{placeholder}</span>
    {/if}
    <svg class="w-3.5 h-3.5 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
    </svg>
  </button>
{/if}
