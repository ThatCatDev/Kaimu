import { graphql } from './client';
import type {
  CreateProjectMutation,
  CreateProjectMutationVariables,
  DeleteProjectMutation,
  DeleteProjectMutationVariables,
  ProjectQuery,
  ProjectQueryVariables,
} from '../graphql/generated';

type CreatedProject = CreateProjectMutation['createProject'];
type ProjectWithOrg = NonNullable<ProjectQuery['project']>;

const CREATE_PROJECT_MUTATION = `
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      name
      key
      description
      createdAt
      updatedAt
    }
  }
`;

const PROJECT_QUERY = `
  query Project($id: ID!) {
    project(id: $id) {
      id
      name
      key
      description
      createdAt
      updatedAt
      organization {
        id
        name
        slug
      }
    }
  }
`;

export async function createProject(
  organizationId: string,
  name: string,
  key: string,
  description?: string
): Promise<CreatedProject> {
  const data = await graphql<CreateProjectMutation>(CREATE_PROJECT_MUTATION, {
    input: { organizationId, name, key, description },
  } as CreateProjectMutationVariables);
  return data.createProject;
}

export async function getProject(id: string): Promise<ProjectWithOrg | null> {
  const data = await graphql<ProjectQuery>(PROJECT_QUERY, {
    id,
  } as ProjectQueryVariables);
  return data.project ?? null;
}

const DELETE_PROJECT_MUTATION = `
  mutation DeleteProject($id: ID!) {
    deleteProject(id: $id)
  }
`;

export async function deleteProject(id: string): Promise<boolean> {
  const data = await graphql<DeleteProjectMutation>(
    DELETE_PROJECT_MUTATION,
    { id } as DeleteProjectMutationVariables
  );
  return data.deleteProject;
}
