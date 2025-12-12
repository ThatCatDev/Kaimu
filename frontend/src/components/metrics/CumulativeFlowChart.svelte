<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables, type ChartConfiguration } from 'chart.js';
  import type { CumulativeFlowData } from '../../lib/api/metrics';

  Chart.register(...registerables);

  interface Props {
    data: CumulativeFlowData;
    mode: 'CARD_COUNT' | 'STORY_POINTS';
  }

  let { data, mode }: Props = $props();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  // Generate colors for columns that don't have one
  const defaultColors = [
    '#3B82F6', '#10B981', '#F59E0B', '#EF4444',
    '#8B5CF6', '#EC4899', '#06B6D4', '#84CC16'
  ];

  function createChart() {
    if (chart) {
      chart.destroy();
    }

    // Create plain arrays to avoid Svelte 5 reactive proxy issues with Chart.js
    const labels = [...data.dates.map(d => formatDate(d))];

    const datasets = data.columns.map((col, index) => ({
      label: col.columnName,
      data: [...col.values],
      backgroundColor: col.color || defaultColors[index % defaultColors.length],
      fill: true,
      tension: 0.3,
    }));

    const config: ChartConfiguration = {
      type: 'line',
      data: {
        labels,
        datasets: datasets as any,
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'top',
            labels: {
              boxWidth: 12,
              padding: 10,
            }
          },
          tooltip: {
            mode: 'index',
            intersect: false,
          }
        },
        scales: {
          y: {
            stacked: true,
            beginAtZero: true,
            title: {
              display: true,
              text: mode === 'STORY_POINTS' ? 'Story Points' : 'Cards',
            }
          },
          x: {
            title: {
              display: true,
              text: 'Date',
            }
          }
        }
      }
    };

    chart = new Chart(canvas, config);
  }

  onMount(() => {
    createChart();
  });

  $effect(() => {
    if (data && canvas) {
      createChart();
    }
  });

  onDestroy(() => {
    if (chart) {
      chart.destroy();
    }
  });
</script>

<div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
  <h3 class="text-sm font-medium text-gray-700 mb-3">
    Cumulative Flow ({mode === 'STORY_POINTS' ? 'Story Points' : 'Card Count'})
  </h3>
  <div class="h-64">
    <canvas bind:this={canvas}></canvas>
  </div>
</div>
