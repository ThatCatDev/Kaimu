---
title: Projects & Boards
description: Setting up projects and Kanban boards in Kaimu
---

Projects and boards help you organize and visualize your work in Kaimu.

## Projects

Projects are containers for related work within an organization. They might represent a product, a service, an initiative, or any logical grouping of tasks.

### Creating a Project

1. From the organization dashboard, click **New Project**
2. Enter a **name** (e.g., "Website Redesign")
3. Enter a **key** (e.g., "WEB") - this is used as a prefix for cards
4. Optionally add a **description**
5. Click **Create Project**

### Project Key

The project key is a short identifier (1-10 characters) that prefixes all cards in the project. For example, if your key is "WEB", your cards will be numbered WEB-1, WEB-2, etc.

:::tip
Choose a short, memorable key that's easy to reference in conversations and commits.
:::

### Project Settings

Access project settings by clicking the gear icon:

- **General**: Edit name, key, and description
- **Tags**: Manage tags for categorizing cards
- **Members**: View project access (inherited from organization)

## Boards

Boards visualize your workflow using the Kanban methodology. Each project can have multiple boards for different purposes (e.g., "Development", "Marketing", "Support").

### Creating a Board

1. From the project page, click **New Board**
2. Enter a **name** (e.g., "Development Board")
3. Optionally add a **description**
4. Click **Create Board**

New boards come with default columns: **Backlog**, **Todo**, **In Progress**, and **Done**.

### Board Views

Each board has three views, accessible via tabs:

| View | Purpose |
|------|---------|
| **Board** | Kanban view with drag-and-drop cards |
| **Planning** | Sprint planning with backlog and sprint sections |
| **Metrics** | Charts showing burndown, cumulative flow, etc. |

## Columns

Columns represent stages in your workflow. Cards move through columns from left to right.

### Managing Columns

Click **Board Settings** (gear icon) to manage columns:

- **Add Column**: Create a new workflow stage
- **Rename**: Change the column name
- **Reorder**: Drag columns to rearrange
- **Delete**: Remove unused columns (moves cards to first column)

### Column Properties

| Property | Description |
|----------|-------------|
| **Name** | Display name of the column |
| **Position** | Order from left to right |
| **Color** | Visual indicator color |
| **WIP Limit** | Maximum cards allowed (optional) |
| **Is Done** | Marks column as a "done" state for metrics |
| **Is Backlog** | Marks column as the backlog (first column) |

### WIP Limits

Work-In-Progress (WIP) limits help prevent bottlenecks:

1. Click **Board Settings**
2. Select a column
3. Set a **WIP Limit** number
4. The column header shows a warning when limit is exceeded

:::tip
Start with WIP limits of 3-5 cards per person in a column. Adjust based on your team's flow.
:::

### Done Column

Mark one column as "Done" to enable metrics:

1. Click **Board Settings**
2. Select your done column (e.g., "Done", "Closed", "Deployed")
3. Enable **Is Done**

This tells Kaimu which cards are complete for burndown charts and velocity calculations.

## Tags

Tags help categorize and filter cards. They're defined at the project level and can be used across all boards.

### Creating Tags

1. Go to **Project Settings** → **Tags**
2. Click **New Tag**
3. Enter a **name** (e.g., "Bug", "Feature", "Tech Debt")
4. Choose a **color**
5. Click **Create**

### Using Tags

- Add tags to cards from the card detail panel
- Filter the board by clicking on a tag
- Search for cards with specific tags using the search bar

## Best Practices

### Board Organization

- **One board per team**: Keep boards focused on one team's work
- **Clear column names**: Use verb forms (e.g., "Doing" not "Development")
- **Limit columns**: 4-6 columns is usually enough
- **Define "done"**: Be explicit about what "done" means

### Project Structure

- **One project per product**: Group related work together
- **Use meaningful keys**: Short keys are easier to remember
- **Document in description**: Explain the project's purpose

### Workflow Tips

- **Start simple**: Begin with basic columns and add complexity as needed
- **Review regularly**: Adjust your workflow based on what you learn
- **Use WIP limits**: Prevent overload and improve flow
- **Track metrics**: Use the Metrics view to identify bottlenecks

## Next Steps

- [Cards & Sprints](/usage/cards-sprints/) - Track individual work items
- [Core Concepts](/usage/concepts/) - Review the overall hierarchy
