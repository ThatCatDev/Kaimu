<script lang="ts">
  interface Props {
    class?: string;
    variant?: 'text' | 'circular' | 'rectangular';
    width?: string;
    height?: string;
    lines?: number;
  }

  let {
    class: className = '',
    variant = 'text',
    width,
    height,
    lines = 1,
  }: Props = $props();

  const baseClasses = 'animate-pulse bg-gray-200 rounded';

  const variantClasses = {
    text: 'h-4 rounded',
    circular: 'rounded-full',
    rectangular: 'rounded-lg',
  };

  function getStyle(): string {
    const styles: string[] = [];
    if (width) styles.push(`width: ${width}`);
    if (height) styles.push(`height: ${height}`);
    return styles.join('; ');
  }
</script>

{#if lines > 1}
  <div class="space-y-2 {className}">
    {#each Array(lines) as _, i}
      <div
        class="{baseClasses} {variantClasses[variant]}"
        style="{getStyle()}{i === lines - 1 ? '; width: 75%' : ''}"
      ></div>
    {/each}
  </div>
{:else}
  <div
    class="{baseClasses} {variantClasses[variant]} {className}"
    style={getStyle()}
  ></div>
{/if}
