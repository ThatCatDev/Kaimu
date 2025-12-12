<script lang="ts">
  import type { MetricMode } from '../../lib/api/metrics';

  interface Props {
    mode: MetricMode;
    onModeChange: (mode: MetricMode) => void;
  }

  let { mode, onModeChange }: Props = $props();
</script>

<div class="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
  <button
    type="button"
    onclick={() => onModeChange('CARD_COUNT')}
    class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {mode === 'CARD_COUNT'
      ? 'bg-white text-gray-900 shadow-sm'
      : 'text-gray-600 hover:text-gray-900'}"
  >
    Cards
  </button>
  <button
    type="button"
    onclick={() => onModeChange('STORY_POINTS')}
    class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {mode === 'STORY_POINTS'
      ? 'bg-white text-gray-900 shadow-sm'
      : 'text-gray-600 hover:text-gray-900'}"
  >
    Story Points
  </button>
</div>
