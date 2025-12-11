import { graphql } from './client';
import type {
  SearchResult as SearchResultType,
  SearchResults as SearchResultsType,
  SearchScope as SearchScopeType,
} from '../graphql/generated';
import { SearchEntityType } from '../graphql/generated';

export type SearchResult = SearchResultType;
export type SearchResults = SearchResultsType;
export type SearchScope = SearchScopeType;
export { SearchEntityType };

const SEARCH_QUERY = `
  query Search($query: String!, $scope: SearchScope, $limit: Int) {
    search(query: $query, scope: $scope, limit: $limit) {
      results {
        type
        id
        title
        description
        highlight
        organizationId
        organizationName
        projectId
        projectName
        boardId
        boardName
        url
        score
      }
      totalCount
      query
    }
  }
`;

interface SearchQueryResponse {
  search: SearchResults;
}

export async function search(
  query: string,
  scope?: SearchScope,
  limit: number = 20
): Promise<SearchResults> {
  const result = await graphql<SearchQueryResponse>(SEARCH_QUERY, {
    query,
    scope: scope || null,
    limit,
  });
  return result.search;
}
