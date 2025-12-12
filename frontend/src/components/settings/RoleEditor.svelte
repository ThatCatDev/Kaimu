<script lang="ts">
  import {
    createRole,
    updateRole,
    type Role,
    type Permission
  } from '../../lib/api/rbac';
  import { Modal, Input, Textarea, BitsSelect } from '../ui';

  interface Props {
    organizationId: string;
    role: Role | null;
    template: Role | null;  // Role to base new role on (for duplication)
    allRoles: Role[];       // All available roles for "based on" dropdown
    permissions: Permission[];
    open: boolean;
    onClose: () => void;
  }

  let { organizationId, role, template, allRoles, permissions, open, onClose }: Props = $props();

  let name = $state('');
  let description = $state('');
  let selectedPermissions = $state<Set<string>>(new Set());
  let selectedTemplateId = $state<string>('');  // For "based on" dropdown
  let error = $state<string | null>(null);
  let saving = $state(false);

  const isEditing = $derived(role !== null);
  const isSystem = $derived(role?.isSystem ?? false);
  const isCreating = $derived(!isEditing);

  // Reset form when modal opens or role changes
  $effect(() => {
    if (open) {
      if (role) {
        // Editing existing role
        name = role.name;
        description = role.description || '';
        selectedPermissions = new Set(role.permissions?.map(p => p.code) || []);
        selectedTemplateId = '';
      } else if (template) {
        // Creating new role based on template (duplicate)
        name = `${template.name} (Copy)`;
        description = template.description || '';
        selectedPermissions = new Set(template.permissions?.map(p => p.code) || []);
        selectedTemplateId = template.id;
      } else {
        // Creating new role from scratch
        name = '';
        description = '';
        selectedPermissions = new Set();
        selectedTemplateId = '';
      }
      error = null;
    }
  });

  // Build options for "based on" dropdown
  const templateOptions = $derived.by(() => {
    const options: Array<{ value: string; label: string }> = [
      { value: '', label: 'Start from scratch' }
    ];

    // Add system roles first
    const systemRoles = allRoles.filter(r => r.isSystem);
    for (const r of systemRoles) {
      options.push({ value: r.id, label: `${r.name} (System)` });
    }

    // Add custom roles
    const customRoles = allRoles.filter(r => !r.isSystem);
    for (const r of customRoles) {
      options.push({ value: r.id, label: r.name });
    }

    return options;
  });

  // Handle "based on" dropdown change
  $effect(() => {
    // Only apply template when creating (not editing) and when templateId changes
    if (isCreating && selectedTemplateId) {
      const selectedTemplate = allRoles.find(r => r.id === selectedTemplateId);
      if (selectedTemplate) {
        // Copy permissions from selected template
        selectedPermissions = new Set(selectedTemplate.permissions?.map(p => p.code) || []);
        // Optionally prefill description if empty
        if (!description.trim() && selectedTemplate.description) {
          description = selectedTemplate.description;
        }
      }
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!name.trim()) return;
    if (isSystem) return; // Can't edit system roles

    try {
      saving = true;
      error = null;

      const permissionCodes = Array.from(selectedPermissions);

      if (isEditing && role) {
        await updateRole(role.id, {
          name: name.trim(),
          description: description.trim() || undefined,
          permissionCodes
        });
      } else {
        await createRole(organizationId, name.trim(), description.trim() || undefined, permissionCodes);
      }

      onClose();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save role';
    } finally {
      saving = false;
    }
  }

  function togglePermission(code: string) {
    if (isSystem) return;
    const newSet = new Set(selectedPermissions);
    if (newSet.has(code)) {
      newSet.delete(code);
    } else {
      newSet.add(code);
    }
    selectedPermissions = newSet;
  }

  function selectAllInCategory(categoryPerms: Permission[]) {
    if (isSystem) return;
    const newSet = new Set(selectedPermissions);
    categoryPerms.forEach(p => newSet.add(p.code));
    selectedPermissions = newSet;
  }

  function deselectAllInCategory(categoryPerms: Permission[]) {
    if (isSystem) return;
    const newSet = new Set(selectedPermissions);
    categoryPerms.forEach(p => newSet.delete(p.code));
    selectedPermissions = newSet;
  }

  // Group permissions by resource type
  const permissionsByCategory = $derived.by(() => {
    const groups: Record<string, Permission[]> = {};
    for (const perm of permissions) {
      const category = perm.resourceType;
      if (!groups[category]) {
        groups[category] = [];
      }
      groups[category].push(perm);
    }
    return groups;
  });

  const categoryLabels: Record<string, string> = {
    organization: 'Organization',
    project: 'Project',
    board: 'Board',
    card: 'Card',
  };
</script>

<Modal
  {open}
  onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}
  title={isSystem ? `View Role: ${name}` : (isEditing ? 'Edit Role' : 'Create Role')}
  size="lg"
>
  <form onsubmit={handleSubmit} class="p-6 space-y-6">
    {#if error}
      <div class="rounded-md bg-red-50 p-3">
        <p class="text-sm text-red-700">{error}</p>
      </div>
    {/if}

    <div class="space-y-4">
      <Input
        label="Role Name"
        bind:value={name}
        placeholder="e.g., Developer, QA Tester"
        required
        disabled={isSystem}
      />

      <Textarea
        label="Description"
        bind:value={description}
        placeholder="Describe what this role is for..."
        rows={2}
        disabled={isSystem}
      />

      {#if isCreating && allRoles.length > 0}
        <BitsSelect
          id="basedOn"
          label="Based on"
          options={templateOptions}
          bind:value={selectedTemplateId}
          placeholder="Start from scratch"
        />
        <p class="-mt-3 text-xs text-gray-500">
          Copy permissions from an existing role to get started quickly
        </p>
      {/if}
    </div>

    <div class="space-y-4">
      <h3 class="text-sm font-medium text-gray-900">Permissions</h3>

      {#each Object.entries(permissionsByCategory) as [category, categoryPerms] (category)}
        <div class="border border-gray-200 rounded-lg overflow-hidden">
          <div class="bg-gray-50 px-4 py-2 flex items-center justify-between">
            <h4 class="text-sm font-medium text-gray-700">
              {categoryLabels[category] || category}
            </h4>
            {#if !isSystem}
              <div class="flex gap-2">
                <button
                  type="button"
                  onclick={() => selectAllInCategory(categoryPerms)}
                  class="text-xs text-indigo-600 hover:text-indigo-800"
                >
                  Select All
                </button>
                <span class="text-gray-300">|</span>
                <button
                  type="button"
                  onclick={() => deselectAllInCategory(categoryPerms)}
                  class="text-xs text-gray-600 hover:text-gray-800"
                >
                  Clear
                </button>
              </div>
            {/if}
          </div>
          <div class="p-4 space-y-2">
            {#each categoryPerms as perm (perm.id)}
              <label class="flex items-start gap-3 {isSystem ? 'cursor-default' : 'cursor-pointer'}">
                <input
                  type="checkbox"
                  checked={selectedPermissions.has(perm.code)}
                  onchange={() => togglePermission(perm.code)}
                  disabled={isSystem}
                  class="mt-0.5 h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 disabled:opacity-60"
                />
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium text-gray-900">{perm.name}</div>
                  {#if perm.description}
                    <div class="text-xs text-gray-500">{perm.description}</div>
                  {/if}
                  <div class="text-xs text-gray-400 font-mono">{perm.code}</div>
                </div>
              </label>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </form>

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <button
        type="button"
        onclick={onClose}
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        {isSystem ? 'Close' : 'Cancel'}
      </button>
      {#if !isSystem}
        <button
          type="button"
          onclick={handleSubmit}
          disabled={saving || !name.trim()}
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Saving...' : (isEditing ? 'Save Changes' : 'Create Role')}
        </button>
      {/if}
    </div>
  {/snippet}
</Modal>
