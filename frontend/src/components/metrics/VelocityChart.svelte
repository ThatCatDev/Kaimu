<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables, type ChartConfiguration } from 'chart.js';
  import type { VelocityData } from '../../lib/api/metrics';

  Chart.register(...registerables);

  interface Props {
    data: VelocityData;
    mode: 'CARD_COUNT' | 'STORY_POINTS';
  }

  let { data, mode }: Props = $props();

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;

  function createChart() {
    if (chart) {
      chart.destroy();
    }

    // Create plain arrays to avoid Svelte 5 reactive proxy issues with Chart.js
    const labels = [...data.sprints.map(s => s.sprintName)];
    const values = [...data.sprints.map(s =>
      mode === 'STORY_POINTS' ? s.completedPoints : s.completedCards
    )];

    // Calculate average
    const average = values.length > 0
      ? values.reduce((a, b) => a + b, 0) / values.length
      : 0;
    const averageLine = [...values.map(() => average)];

    const config: ChartConfiguration = {
      type: 'bar',
      data: {
        labels,
        datasets: [
          {
            label: mode === 'STORY_POINTS' ? 'Story Points' : 'Cards',
            data: values,
            backgroundColor: '#6366F1',
            borderRadius: 4,
          },
          {
            label: 'Average',
            data: averageLine,
            type: 'line',
            borderColor: '#F59E0B',
            borderDash: [5, 5],
            fill: false,
            pointRadius: 0,
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
              text: 'Sprint',
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
    Velocity ({mode === 'STORY_POINTS' ? 'Story Points' : 'Card Count'})
  </h3>
  <div class="h-64">
    <canvas bind:this={canvas}></canvas>
  </div>
</div>
