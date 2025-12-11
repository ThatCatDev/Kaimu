<script lang="ts">
  import { DatePicker } from 'bits-ui';
  import { CalendarDate, parseDate, today, getLocalTimeZone } from '@internationalized/date';

  interface Props {
    value?: string; // ISO date string (YYYY-MM-DD)
    label?: string;
    error?: string | null;
    disabled?: boolean;
    required?: boolean;
    placeholder?: string;
    id?: string;
    minValue?: string;
    maxValue?: string;
    readOnly?: boolean;
    onValueChange?: (value: string) => void;
  }

  let {
    value = $bindable(''),
    label,
    error,
    disabled = false,
    required = false,
    placeholder = 'Select date...',
    id,
    minValue,
    maxValue,
    readOnly = false,
    onValueChange
  }: Props = $props();

  function formatDisplayDate(dateStr: string): string {
    if (!dateStr) return '—';
    try {
      // Parse the date parts to avoid timezone conversion issues
      // Input format is YYYY-MM-DD
      const [year, month, day] = dateStr.split('-').map(Number);
      const date = new Date(year, month - 1, day); // month is 0-indexed
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return dateStr;
    }
  }

  // Convert string to CalendarDate for Bits UI
  let calendarValue = $state<CalendarDate | undefined>(undefined);

  // Sync from string value to CalendarDate
  $effect(() => {
    if (value) {
      try {
        calendarValue = parseDate(value);
      } catch {
        calendarValue = undefined;
      }
    } else {
      calendarValue = undefined;
    }
  });

  // Sync from CalendarDate to string value
  function handleValueChange(newValue: CalendarDate | undefined) {
    calendarValue = newValue;
    const newStringValue = newValue ? newValue.toString() : '';
    value = newStringValue;
    onValueChange?.(newStringValue);
  }

  function clearDate(e: Event) {
    e.stopPropagation();
    calendarValue = undefined;
    value = '';
    onValueChange?.('');
  }

  const minDate = $derived(minValue ? parseDate(minValue) : undefined);
  const maxDate = $derived(maxValue ? parseDate(maxValue) : undefined);
  const placeholderDate = $derived(today(getLocalTimeZone()));
</script>

<div class="w-full">
  {#if label}
    <label for={id} class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
      {label}
      {#if required && !readOnly}
        <span class="text-red-500">*</span>
      {/if}
    </label>
  {/if}

  {#if readOnly}
    <p class="text-sm text-gray-900">{formatDisplayDate(value)}</p>
  {:else}
    <DatePicker.Root
    value={calendarValue}
    onValueChange={handleValueChange}
    placeholder={placeholderDate}
    weekdayFormat="short"
    fixedWeeks={true}
    minValue={minDate}
    maxValue={maxDate}
    {disabled}
  >
    <DatePicker.Input
      {id}
      class="flex h-auto w-full items-center rounded bg-transparent px-2 py-1.5 text-sm transition-colors hover:bg-gray-50 focus-within:outline-none focus-within:bg-gray-50 focus-within:ring-1 focus-within:ring-indigo-500 {error ? 'ring-1 ring-red-500 bg-red-50' : ''} {disabled ? 'cursor-not-allowed opacity-50' : ''}"
    >
      {#snippet children({ segments })}
        {#each segments as { part, value: segValue }}
          {#if part === 'literal'}
            <span class="text-gray-400">{segValue}</span>
          {:else}
            <DatePicker.Segment
              {part}
              class="rounded px-0.5 tabular-nums text-gray-900 outline-none focus:bg-indigo-100 data-[placeholder]:text-gray-400"
            >
              {segValue}
            </DatePicker.Segment>
          {/if}
        {/each}

        {#if value && !disabled}
          <button
            type="button"
            onclick={clearDate}
            class="ml-auto inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors"
            title="Clear date"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}

        <DatePicker.Trigger
          class="{value ? '' : 'ml-auto'} inline-flex h-8 w-8 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors disabled:pointer-events-none disabled:opacity-50"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
        </DatePicker.Trigger>
      {/snippet}
    </DatePicker.Input>

    <DatePicker.Content
      sideOffset={4}
      avoidCollisions={true}
      collisionPadding={16}
      class="z-50 rounded-lg border border-gray-200 bg-white p-4 shadow-lg animate-in fade-in-0 zoom-in-95"
    >
      <DatePicker.Calendar>
        {#snippet children({ months, weekdays })}
          <DatePicker.Header class="flex items-center justify-between mb-4">
            <DatePicker.PrevButton
              class="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-gray-100 transition-colors disabled:pointer-events-none disabled:opacity-50"
            >
              <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </DatePicker.PrevButton>

            <DatePicker.Heading class="text-sm font-semibold text-gray-900" />

            <DatePicker.NextButton
              class="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-gray-100 transition-colors disabled:pointer-events-none disabled:opacity-50"
            >
              <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </DatePicker.NextButton>
          </DatePicker.Header>

          {#each months as month (month.value.toString())}
            <DatePicker.Grid class="w-full">
              <DatePicker.GridHead>
                <DatePicker.GridRow class="flex">
                  {#each weekdays as day}
                    <DatePicker.HeadCell
                      class="w-10 text-center text-xs font-medium text-gray-500"
                    >
                      {day.slice(0, 2)}
                    </DatePicker.HeadCell>
                  {/each}
                </DatePicker.GridRow>
              </DatePicker.GridHead>

              <DatePicker.GridBody>
                {#each month.weeks as weekDates}
                  <DatePicker.GridRow class="flex">
                    {#each weekDates as date}
                      <DatePicker.Cell {date} month={month.value} class="p-0">
                        <DatePicker.Day
                          class="inline-flex h-10 w-10 items-center justify-center rounded-md text-sm transition-colors hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 data-[selected]:bg-indigo-600 data-[selected]:text-white data-[selected]:hover:bg-indigo-700 data-[outside-month]:text-gray-300 data-[outside-month]:pointer-events-none data-[disabled]:text-gray-300 data-[disabled]:pointer-events-none data-[today]:font-semibold data-[today]:text-indigo-600 data-[today]:data-[selected]:text-white"
                        >
                          {date.day}
                        </DatePicker.Day>
                      </DatePicker.Cell>
                    {/each}
                  </DatePicker.GridRow>
                {/each}
              </DatePicker.GridBody>
            </DatePicker.Grid>
          {/each}
        {/snippet}
      </DatePicker.Calendar>
    </DatePicker.Content>
    </DatePicker.Root>
  {/if}

  {#if error && !readOnly}
    <p class="mt-1 text-xs text-red-600">{error}</p>
  {/if}
</div>

