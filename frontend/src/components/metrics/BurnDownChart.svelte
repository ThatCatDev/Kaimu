<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables, type ChartConfiguration } from 'chart.js';
  import type { BurnDownData } from '../../lib/api/metrics';

  Chart.register(...registerables);

  interface Props {
    data: BurnDownData;
    mode: 'CARD_COUNT' | 'STORY_POINTS';
  }

  let { data, mode }: Props = $props();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function createChart() {
    if (chart) {
      chart.destroy();
    }

    // Create plain arrays to avoid Svelte 5 reactive proxy issues with Chart.js
    const idealLabels = [...data.idealLine.map(p => formatDate(p.date))];
    const idealData = [...data.idealLine.map(p => p.value)];
    const actualData = [...data.actualLine.map(p => p.value)];

    const config: ChartConfiguration = {
      type: 'line',
      data: {
        labels: idealLabels,
        datasets: [
          {
            label: 'Ideal',
            data: idealData,
            borderColor: '#9CA3AF',
            borderDash: [5, 5],
            fill: false,
            tension: 0,
            pointRadius: 0,
          },
          {
            label: 'Remaining',
            data: actualData,
            borderColor: '#3B82F6',
            backgroundColor: 'rgba(59, 130, 246, 0.1)',
            fill: true,
            tension: 0.1,
            pointRadius: 3,
            pointBackgroundColor: '#3B82F6',
          }
        ]
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
    Burn Down ({mode === 'STORY_POINTS' ? 'Story Points' : 'Card Count'})
  </h3>
  <div class="h-64">
    <canvas bind:this={canvas}></canvas>
  </div>
</div>
