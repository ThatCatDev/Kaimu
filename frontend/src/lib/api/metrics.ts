import { graphql } from './client';
import type {
  BurnDownData,
  BurnUpData,
  VelocityData,
  CumulativeFlowData,
  SprintStats,
  MetricMode
} from '../graphql/generated';

export type { BurnDownData, BurnUpData, VelocityData, CumulativeFlowData, SprintStats, MetricMode };

export async function getBurnDownData(sprintId: string, mode: MetricMode): Promise<BurnDownData | null> {
  const query = `
    query GetBurnDownData($sprintId: ID!, $mode: MetricMode!) {
      burnDownData(sprintId: $sprintId, mode: $mode) {
        sprintId
        sprintName
        startDate
        endDate
        idealLine { date value }
        actualLine { date value }
      }
    }
  `;
  const result = await graphql<{ burnDownData: BurnDownData | null }>(query, { sprintId, mode });
  return result.burnDownData;
}

export async function getBurnUpData(sprintId: string, mode: MetricMode): Promise<BurnUpData | null> {
  const query = `
    query GetBurnUpData($sprintId: ID!, $mode: MetricMode!) {
      burnUpData(sprintId: $sprintId, mode: $mode) {
        sprintId
        sprintName
        startDate
        endDate
        scopeLine { date value }
        doneLine { date value }
      }
    }
  `;
  const result = await graphql<{ burnUpData: BurnUpData | null }>(query, { sprintId, mode });
  return result.burnUpData;
}

export async function getVelocityData(boardId: string, mode: MetricMode, sprintCount?: number): Promise<VelocityData> {
  const query = `
    query GetVelocityData($boardId: ID!, $mode: MetricMode!, $sprintCount: Int) {
      velocityData(boardId: $boardId, mode: $mode, sprintCount: $sprintCount) {
        sprints {
          sprintId
          sprintName
          completedCards
          completedPoints
        }
      }
    }
  `;
  const result = await graphql<{ velocityData: VelocityData }>(query, { boardId, mode, sprintCount });
  return result.velocityData;
}

export async function getCumulativeFlowData(sprintId: string, mode: MetricMode): Promise<CumulativeFlowData | null> {
  const query = `
    query GetCumulativeFlowData($sprintId: ID!, $mode: MetricMode!) {
      cumulativeFlowData(sprintId: $sprintId, mode: $mode) {
        sprintId
        sprintName
        columns { columnId columnName color values }
        dates
      }
    }
  `;
  const result = await graphql<{ cumulativeFlowData: CumulativeFlowData | null }>(query, { sprintId, mode });
  return result.cumulativeFlowData;
}

export async function getSprintStats(sprintId: string): Promise<SprintStats | null> {
  const query = `
    query GetSprintStats($sprintId: ID!) {
      sprintStats(sprintId: $sprintId) {
        totalCards
        completedCards
        totalStoryPoints
        completedStoryPoints
        daysRemaining
        daysElapsed
      }
    }
  `;
  const result = await graphql<{ sprintStats: SprintStats | null }>(query, { sprintId });
  return result.sprintStats;
}
