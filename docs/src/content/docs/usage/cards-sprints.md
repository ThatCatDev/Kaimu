---
title: Cards & Sprints
description: Tracking work items and managing sprints in Kaimu
---

Cards are the core work items in Kaimu. Sprints help you organize cards into time-boxed iterations for agile workflows.

## Cards

Cards represent individual units of work: tasks, bugs, stories, or any work item your team needs to track.

### Creating Cards

**From the Board View:**
1. Click the **+** button in any column, or
2. Click **New Card** in the toolbar
3. Enter a **title**
4. Click **Create** (or press Enter)

**Quick creation:** Type in the "Add card" input at the bottom of any column.

### Card Properties

| Property | Description |
|----------|-------------|
| **Title** | Brief description of the work |
| **Description** | Detailed information (supports rich text) |
| **Priority** | None, Low, Medium, High, or Urgent |
| **Assignee** | Team member responsible |
| **Due Date** | Target completion date |
| **Story Points** | Estimation for sprint planning |
| **Tags** | Labels for categorization |
| **Sprints** | Which sprint(s) the card belongs to |

### Editing Cards

Click on any card to open the detail panel:

- **Title**: Click to edit inline
- **Description**: Rich text editor with markdown support
- **Properties**: Use the sidebar fields to set priority, assignee, etc.

Changes are saved automatically.

### Moving Cards

**Drag and drop:** Grab a card and drop it in another column to change its status.

**Keyboard:**
- Open a card
- Use the column dropdown to move it

### Card Priorities

Use priorities to highlight important work:

| Priority | When to Use |
|----------|-------------|
| **Urgent** | Blocking issues, production incidents |
| **High** | Important work that should be done soon |
| **Medium** | Normal priority work |
| **Low** | Nice-to-have, can wait |
| **None** | Unspecified priority |

### Story Points

Story points estimate the effort required for a card. They're used for sprint planning and velocity tracking.

Common scales:
- **Fibonacci**: 1, 2, 3, 5, 8, 13, 21
- **T-shirt sizes**: 1 (XS), 2 (S), 3 (M), 5 (L), 8 (XL)

:::tip
Story points measure relative effort, not time. A 2-point card should be roughly twice the effort of a 1-point card.
:::

## Sprints

Sprints are time-boxed periods (typically 1-4 weeks) where your team commits to completing a set of work.

### Creating Sprints

1. Go to the **Planning** view
2. Click **New Sprint**
3. Enter a **name** (e.g., "Sprint 23" or "Jan 2024 Sprint")
4. Set **start date** and **end date**
5. Optionally add a **goal**
6. Click **Create**

New sprints start with status **Future**.

### Sprint Status

| Status | Description |
|--------|-------------|
| **Future** | Planned but not started |
| **Active** | Currently in progress (only one active sprint per board) |
| **Closed** | Completed |

### Sprint Lifecycle

```
Future → Active → Closed
```

**Start a sprint:**
1. Go to Planning view
2. Find the future sprint
3. Click **Start Sprint**
4. Confirm the start date

**Complete a sprint:**
1. Click **Complete Sprint** on the active sprint
2. Choose what to do with incomplete cards:
   - Move to next sprint
   - Return to backlog

### Adding Cards to Sprints

**From the card detail panel:**
1. Open a card
2. In the **Sprints** section, check the sprint(s)
3. Changes save automatically

**From the Planning view:**
1. Find a card in the Backlog section
2. Click the menu (⋮) on the card
3. Select **Move to Sprint** → choose a sprint

**Drag and drop:** In Planning view, drag cards between sections.

:::note
Cards can belong to multiple sprints. This is useful for work that carries over or spans multiple sprints.
:::

### Sprint Planning

The **Planning** view is designed for sprint planning sessions:

1. **Active Sprint**: Shows cards in the current sprint
2. **Future Sprints**: Planned work for upcoming sprints
3. **Backlog**: Cards not assigned to any sprint
4. **Closed Sprints**: Historical reference

Each section shows:
- Number of cards
- Total story points
- Expandable card list

### Backlog Management

The backlog contains cards not assigned to any sprint:

- **Add to backlog**: Create cards without assigning a sprint
- **Prioritize**: Reorder cards by dragging (coming soon)
- **Groom**: Review and estimate cards before sprint planning

## Metrics

The **Metrics** view provides insights into sprint progress and team performance.

### Burndown Chart

Shows remaining work over time:
- **Ideal line**: Expected progress if work is done evenly
- **Actual line**: Real remaining work
- **Gap**: Indicates if you're ahead or behind

### Cumulative Flow Diagram

Shows cards in each column over time:
- **Width of bands**: WIP in each stage
- **Slope**: Overall throughput
- **Bottlenecks**: Wide bands indicate queues

### Velocity

Track story points completed per sprint:
- **Average velocity**: Plan future sprints
- **Trends**: See improvement over time

## Best Practices

### Card Writing

- **Clear titles**: Make titles specific and actionable
- **Definition of done**: Include acceptance criteria in description
- **Right size**: Cards should be completable in 1-2 days
- **Single responsibility**: One card = one piece of work

### Sprint Management

- **Consistent cadence**: Keep sprint length consistent
- **Realistic planning**: Don't overcommit based on velocity
- **Daily updates**: Move cards through columns as work progresses
- **Sprint reviews**: Reflect on what was completed

### Estimation

- **Team consensus**: Estimate as a team
- **Relative sizing**: Compare to known cards
- **Re-estimate if needed**: Update estimates when you learn more

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `n` | New card |
| `Escape` | Close card panel |
| `/` | Focus search |

## Next Steps

- [Core Concepts](/usage/concepts/) - Review the hierarchy
- [Organizations](/usage/organizations/) - Managing team access
- [Projects & Boards](/usage/projects-boards/) - Board configuration
