<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Editor } from '@tiptap/core';
  import StarterKit from '@tiptap/starter-kit';
  import Placeholder from '@tiptap/extension-placeholder';
  import Link from '@tiptap/extension-link';
  import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
  import { common, createLowlight } from 'lowlight';
  import DOMPurify from 'dompurify';

  // Configure DOMPurify to allow only safe tags and attributes
  function sanitizeHtml(html: string): string {
    if (typeof window === 'undefined') return html; // SSR fallback
    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: [
        'p', 'br', 'strong', 'b', 'em', 'i', 'u', 's', 'code', 'pre',
        'h1', 'h2', 'h3', 'ul', 'ol', 'li', 'blockquote', 'a', 'span'
      ],
      ALLOWED_ATTR: ['href', 'target', 'rel', 'class', 'data-language'],
      ALLOW_DATA_ATTR: true, // Allow data-* attributes for syntax highlighting
    });
  }

  interface Props {
    value?: string;
    placeholder?: string;
    label?: string;
    disabled?: boolean;
    readOnly?: boolean;
    class?: string;
    onUpdate?: (html: string) => void;
  }

  let {
    value = '',
    placeholder = 'Write something...',
    label,
    disabled = false,
    readOnly = false,
    class: className = '',
    onUpdate
  }: Props = $props();

  let element: HTMLDivElement;
  let editor: Editor | null = null;
  let isFocused = $state(false);
  let lastEmittedValue = $state(value); // Track what we last sent to parent to avoid flash on save

  const lowlight = createLowlight(common);

  onMount(() => {
    editor = new Editor({
      element,
      extensions: [
        StarterKit.configure({
          heading: {
            levels: [1, 2, 3]
          },
          codeBlock: false // Disable default codeBlock in favor of CodeBlockLowlight
        }),
        CodeBlockLowlight.configure({
          lowlight,
          defaultLanguage: 'plaintext'
        }),
        Placeholder.configure({
          placeholder,
          emptyEditorClass: 'is-editor-empty'
        }),
        Link.configure({
          openOnClick: false,
          HTMLAttributes: {
            class: 'text-indigo-600 underline hover:text-indigo-800',
            rel: 'noopener noreferrer nofollow'
          },
          validate: (href) => /^https?:\/\//.test(href) || href.startsWith('mailto:')
        })
      ],
      content: value,
      editable: !disabled && !readOnly,
      onUpdate: ({ editor }) => {
        const html = editor.getHTML();
        // Return empty string if only contains empty paragraph
        const cleanHtml = html === '<p></p>' ? '' : html;
        // Sanitize HTML before passing to parent
        const sanitized = sanitizeHtml(cleanHtml);
        lastEmittedValue = sanitized;
        onUpdate?.(sanitized);
      },
      onFocus: () => {
        isFocused = true;
      },
      onBlur: () => {
        isFocused = false;
      },
      editorProps: {
        attributes: {
          class: 'outline-none min-h-[80px] prose prose-sm max-w-none'
        }
      }
    });
  });

  onDestroy(() => {
    editor?.destroy();
  });

  // Update editor content when value prop changes externally (not from our own edits)
  $effect(() => {
    if (!editor) return;

    // Skip if this is our own change coming back
    if (value === lastEmittedValue) return;

    // Skip if editor is focused (user is typing)
    if (isFocused) return;

    // Compare normalized values
    const currentHtml = editor.getHTML();
    const normalizedCurrent = currentHtml === '<p></p>' ? '' : currentHtml;
    if (value !== normalizedCurrent) {
      editor.commands.setContent(value || '', false); // false = don't emit update event
      lastEmittedValue = value;
    }
  });

  // Update editable state when disabled/readOnly changes
  $effect(() => {
    if (editor) {
      editor.setEditable(!disabled && !readOnly);
    }
  });

  function toggleBold() {
    editor?.chain().focus().toggleBold().run();
  }

  function toggleItalic() {
    editor?.chain().focus().toggleItalic().run();
  }

  function toggleStrike() {
    editor?.chain().focus().toggleStrike().run();
  }

  function toggleCode() {
    editor?.chain().focus().toggleCode().run();
  }

  function toggleBulletList() {
    editor?.chain().focus().toggleBulletList().run();
  }

  function toggleOrderedList() {
    editor?.chain().focus().toggleOrderedList().run();
  }

  function toggleCodeBlock() {
    editor?.chain().focus().toggleCodeBlock().run();
  }

  function toggleBlockquote() {
    editor?.chain().focus().toggleBlockquote().run();
  }

  function setLink() {
    const url = window.prompt('Enter URL');
    if (url) {
      editor?.chain().focus().setLink({ href: url }).run();
    }
  }

  function unsetLink() {
    editor?.chain().focus().unsetLink().run();
  }

  const isActive = (name: string, attrs?: Record<string, unknown>) => {
    return editor?.isActive(name, attrs) ?? false;
  };
</script>

<div class="w-full {className}">
  {#if label}
    <label class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
      {label}
    </label>
  {/if}

  {#if readOnly}
    <div class="rich-text-content text-sm text-gray-900">
      {#if value}
        {@html sanitizeHtml(value)}
      {:else}
        <span class="text-gray-400">—</span>
      {/if}
    </div>
  {:else}
    <div
      class="rounded transition-all {isFocused ? 'bg-gray-50 ring-1 ring-indigo-500' : 'hover:bg-gray-50'} {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
    >
      <!-- Toolbar -->
      {#if !disabled}
        <div class="flex flex-wrap items-center gap-0.5 px-2 py-1.5 border-b border-gray-200">
          <button
            type="button"
            onclick={toggleBold}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('bold') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Bold"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 4h8a4 4 0 014 4 4 4 0 01-4 4H6z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12h9a4 4 0 014 4 4 4 0 01-4 4H6z" />
            </svg>
          </button>
          <button
            type="button"
            onclick={toggleItalic}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('italic') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Italic"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 4h4m-2 0v16m4-16h-4m-4 16h8" />
            </svg>
          </button>
          <button
            type="button"
            onclick={toggleStrike}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('strike') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Strikethrough"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v18M5 12h14" />
            </svg>
          </button>
          <button
            type="button"
            onclick={toggleCode}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('code') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Inline Code"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </button>

          <div class="w-px h-4 bg-gray-300 mx-1"></div>

          <button
            type="button"
            onclick={toggleBulletList}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('bulletList') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Bullet List"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <button
            type="button"
            onclick={toggleOrderedList}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('orderedList') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Numbered List"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20h14M7 12h14M7 4h14M3 20h.01M3 12h.01M3 4h.01" />
            </svg>
          </button>

          <div class="w-px h-4 bg-gray-300 mx-1"></div>

          <button
            type="button"
            onclick={toggleBlockquote}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('blockquote') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Quote"
          >
            <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M6 17h3l2-4V7H5v6h3zm8 0h3l2-4V7h-6v6h3z" />
            </svg>
          </button>
          <button
            type="button"
            onclick={toggleCodeBlock}
            class="p-1.5 rounded hover:bg-gray-200 transition-colors {isActive('codeBlock') ? 'bg-gray-200 text-indigo-600' : 'text-gray-600'}"
            title="Code Block"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </button>

          <div class="w-px h-4 bg-gray-300 mx-1"></div>

          {#if isActive('link')}
            <button
              type="button"
              onclick={unsetLink}
              class="p-1.5 rounded hover:bg-gray-200 transition-colors bg-gray-200 text-indigo-600"
              title="Remove Link"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </button>
          {:else}
            <button
              type="button"
              onclick={setLink}
              class="p-1.5 rounded hover:bg-gray-200 transition-colors text-gray-600"
              title="Add Link"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
              </svg>
            </button>
          {/if}
        </div>
      {/if}

      <!-- Editor -->
      <div bind:this={element} class="px-2 py-1.5"></div>
    </div>
  {/if}
</div>

<style>
  :global(.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    float: left;
    color: #9ca3af;
    pointer-events: none;
    height: 0;
  }

  :global(.ProseMirror) {
    min-height: 80px;
  }

  /* Shared styles for both edit mode (.ProseMirror) and read-only mode (.rich-text-content) */
  :global(.ProseMirror p),
  :global(.rich-text-content p) {
    margin: 0.5em 0;
  }

  :global(.ProseMirror p:first-child),
  :global(.rich-text-content p:first-child) {
    margin-top: 0;
  }

  :global(.ProseMirror p:last-child),
  :global(.rich-text-content p:last-child) {
    margin-bottom: 0;
  }

  :global(.ProseMirror ul),
  :global(.rich-text-content ul) {
    padding-left: 1.5em;
    margin: 0.5em 0;
    list-style-type: disc !important;
    list-style-position: outside;
  }

  :global(.ProseMirror ol),
  :global(.rich-text-content ol) {
    padding-left: 1.5em;
    margin: 0.5em 0;
    list-style-type: decimal !important;
    list-style-position: outside;
  }

  :global(.ProseMirror li),
  :global(.rich-text-content li) {
    margin: 0.25em 0;
    display: list-item !important;
  }

  :global(.ProseMirror ul li),
  :global(.rich-text-content ul li) {
    list-style-type: disc !important;
  }

  :global(.ProseMirror ol li),
  :global(.rich-text-content ol li) {
    list-style-type: decimal !important;
  }

  :global(.ProseMirror code),
  :global(.rich-text-content code) {
    background-color: #f3f4f6;
    padding: 0.125em 0.25em;
    border-radius: 0.25em;
    font-size: 0.875em;
  }

  :global(.ProseMirror pre),
  :global(.rich-text-content pre) {
    background-color: #1f2937;
    color: #f9fafb;
    padding: 0.75em 1em;
    border-radius: 0.375em;
    overflow-x: auto;
    margin: 0.5em 0;
  }

  :global(.ProseMirror pre code),
  :global(.rich-text-content pre code) {
    background: none;
    padding: 0;
    color: inherit;
  }

  :global(.ProseMirror blockquote),
  :global(.rich-text-content blockquote) {
    border-left: 3px solid #d1d5db;
    padding-left: 1em;
    margin: 0.5em 0;
    color: #6b7280;
  }

  :global(.ProseMirror h1),
  :global(.ProseMirror h2),
  :global(.ProseMirror h3),
  :global(.rich-text-content h1),
  :global(.rich-text-content h2),
  :global(.rich-text-content h3) {
    font-weight: 600;
    margin: 0.75em 0 0.5em;
  }

  :global(.ProseMirror h1),
  :global(.rich-text-content h1) {
    font-size: 1.5em;
  }

  :global(.ProseMirror h2),
  :global(.rich-text-content h2) {
    font-size: 1.25em;
  }

  :global(.ProseMirror h3),
  :global(.rich-text-content h3) {
    font-size: 1.125em;
  }

  :global(.rich-text-content strong),
  :global(.rich-text-content b) {
    font-weight: 600;
  }

  :global(.rich-text-content em),
  :global(.rich-text-content i) {
    font-style: italic;
  }

  /* Syntax highlighting - GitHub-inspired dark theme */
  :global(.hljs-comment),
  :global(.hljs-quote) {
    color: #8b949e;
    font-style: italic;
  }

  :global(.hljs-keyword),
  :global(.hljs-selector-tag),
  :global(.hljs-addition) {
    color: #ff7b72;
  }

  :global(.hljs-number),
  :global(.hljs-string),
  :global(.hljs-meta .hljs-meta-string),
  :global(.hljs-literal),
  :global(.hljs-doctag),
  :global(.hljs-regexp) {
    color: #a5d6ff;
  }

  :global(.hljs-title),
  :global(.hljs-section),
  :global(.hljs-name),
  :global(.hljs-selector-id),
  :global(.hljs-selector-class) {
    color: #d2a8ff;
  }

  :global(.hljs-attribute),
  :global(.hljs-attr),
  :global(.hljs-variable),
  :global(.hljs-template-variable),
  :global(.hljs-class .hljs-title),
  :global(.hljs-type) {
    color: #7ee787;
  }

  :global(.hljs-symbol),
  :global(.hljs-bullet),
  :global(.hljs-subst),
  :global(.hljs-meta),
  :global(.hljs-meta .hljs-keyword),
  :global(.hljs-selector-attr),
  :global(.hljs-selector-pseudo),
  :global(.hljs-link) {
    color: #ffa657;
  }

  :global(.hljs-built_in),
  :global(.hljs-deletion) {
    color: #ffa198;
  }

  :global(.hljs-formula) {
    background-color: #3b4048;
  }

  :global(.hljs-emphasis) {
    font-style: italic;
  }

  :global(.hljs-strong) {
    font-weight: bold;
  }
</style>
