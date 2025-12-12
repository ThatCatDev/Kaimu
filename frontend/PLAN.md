# Metrics Dashboard Implementation Plan

## Summary
Add sprint metrics dashboard with burn down/up charts, velocity, and cumulative flow. Support story points and card counts. Accessible from board header.

---

## Step 1: Add Story Points to Cards

**Backend changes:**

1. Create migration `000018_add_story_points.up.sql`:
```sql
ALTER TABLE cards ADD COLUMN story_points INTEGER;
```

2. Update `card_entity.go` - add `StoryPoints *int`

3. Update `types.graphqls`:
   - Add `storyPoints: Int` to Card type
   - Add `storyPoints: Int` to CreateCardInput
   - Add `storyPoints: Int` to UpdateCardInput

4. Update card service and resolvers

---

## Step 2: Add Completed Timestamp to Cards

**Backend changes:**

1. Create migration `000019_add_completed_at.up.sql`:
```sql
ALTER TABLE cards ADD COLUMN completed_at TIMESTAMPTZ;
```

2. Update `card_entity.go` - add `CompletedAt *time.Time`

3. Update card service to set `completed_at` when card moves to a "done" column

---

## Step 3: Create Metrics GraphQL Types

Add to `types.graphqls`:
```graphql
type SprintMetrics {
  totalCards: Int!
  completedCards: Int!
  totalPoints: Int!
  completedPoints: Int!
  burndown: [DailyMetric!]!
  burnup: [DailyMetric!]!
}

type DailyMetric {
  date: Time!
  remaining: Int!
  completed: Int!
  total: Int!
}

type VelocityPoint {
  sprintName: String!
  completedCards: Int!
  completedPoints: Int!
}

type ColumnSnapshot {
  columnName: String!
  cardCount: Int!
}

type FlowSnapshot {
  date: Time!
  columns: [ColumnSnapshot!]!
}
```

Add to `schema.graphqls`:
```graphql
sprintMetrics(sprintId: ID!): SprintMetrics
velocityData(boardId: ID!, sprintCount: Int): [VelocityPoint!]!
cumulativeFlow(boardId: ID!, days: Int): [FlowSnapshot!]!
```

---

## Step 4: Create Metrics Service

New file: `backend/internal/services/metrics/metrics_service.go`

Methods:
- `GetSprintMetrics(sprintID)` - calculates daily burn data from card completion dates
- `GetVelocityData(boardID, count)` - completed cards/points per closed sprint
- `GetCumulativeFlow(boardID, days)` - card counts per column over time

---

## Step 5: Add Chart.js to Frontend

```bash
cd frontend && bun add chart.js
```

---

## Step 6: Create Chart Components

New files in `frontend/src/components/metrics/`:

1. `BurnDownChart.svelte` - line chart, remaining work over time
2. `BurnUpChart.svelte` - line chart, completed vs scope
3. `VelocityChart.svelte` - bar chart, velocity per sprint
4. `CumulativeFlowChart.svelte` - stacked area, cards per column
5. `MetricModeToggle.svelte` - toggle points vs cards

---

## Step 7: Create Metrics Page

New file: `frontend/src/pages/projects/[id]/board/[boardId]/metrics.astro`

Layout:
- Sprint selector dropdown
- Metric mode toggle (points/cards)
- 2x2 grid of charts
- Summary stats row

---

## Step 8: Add Navigation

Add "Metrics" button to board header (in `BoardWithSprints.svelte` or board header component).

---

## Files to Create
- `backend/db/migrations/000018_add_story_points.up.sql`
- `backend/db/migrations/000018_add_story_points.down.sql`
- `backend/db/migrations/000019_add_completed_at.up.sql`
- `backend/db/migrations/000019_add_completed_at.down.sql`
- `backend/internal/services/metrics/metrics_service.go`
- `backend/internal/resolvers/metrics.go`
- `frontend/src/components/metrics/BurnDownChart.svelte`
- `frontend/src/components/metrics/BurnUpChart.svelte`
- `frontend/src/components/metrics/VelocityChart.svelte`
- `frontend/src/components/metrics/CumulativeFlowChart.svelte`
- `frontend/src/components/metrics/MetricModeToggle.svelte`
- `frontend/src/lib/api/metrics.ts`
- `frontend/src/pages/projects/[id]/board/[boardId]/metrics.astro`

## Files to Modify
- `backend/graph/types.graphqls`
- `backend/graph/schema.graphqls`
- `backend/internal/db/repositories/card/card_entity.go`
- `backend/internal/services/card/card_service.go`
- `backend/graph/resolver.go`
- `frontend/package.json`
- `frontend/src/components/kanban/BoardWithSprints.svelte` (add metrics link)
