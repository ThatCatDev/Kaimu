import { graphql } from './client';
import type {
  CreateOrganizationMutation,
  CreateOrganizationMutationVariables,
  UpdateOrganizationMutation,
  UpdateOrganizationMutationVariables,
  DeleteOrganizationMutation,
  DeleteOrganizationMutationVariables,
  OrganizationsQuery,
  OrganizationQuery,
  OrganizationQueryVariables,
} from '../graphql/generated';

type CreatedOrganization = CreateOrganizationMutation['createOrganization'];
type UpdatedOrganization = UpdateOrganizationMutation['updateOrganization'];
type OrganizationListItem = OrganizationsQuery['organizations'][number];
type OrganizationWithProjects = NonNullable<OrganizationQuery['organization']>;

const CREATE_ORGANIZATION_MUTATION = `
  mutation CreateOrganization($input: CreateOrganizationInput!) {
    createOrganization(input: $input) {
      id
      name
      slug
      description
      createdAt
      updatedAt
    }
  }
`;

const ORGANIZATIONS_QUERY = `
  query Organizations {
    organizations {
      id
      name
      slug
      description
      createdAt
      updatedAt
      projects {
        id
        name
        key
        boards {
          id
          name
          isDefault
        }
      }
    }
  }
`;

const ORGANIZATION_QUERY = `
  query Organization($id: ID!) {
    organization(id: $id) {
      id
      name
      slug
      description
      createdAt
      updatedAt
      projects {
        id
        name
        key
        description
        createdAt
        updatedAt
      }
    }
  }
`;

export async function createOrganization(
  name: string,
  description?: string
): Promise<CreatedOrganization> {
  const data = await graphql<CreateOrganizationMutation>(
    CREATE_ORGANIZATION_MUTATION,
    {
      input: { name, description },
    } as CreateOrganizationMutationVariables
  );
  return data.createOrganization;
}

export async function getOrganizations(): Promise<OrganizationListItem[]> {
  const data = await graphql<OrganizationsQuery>(ORGANIZATIONS_QUERY);
  return data.organizations;
}

export async function getOrganization(id: string): Promise<OrganizationWithProjects | null> {
  const data = await graphql<OrganizationQuery>(ORGANIZATION_QUERY, {
    id,
  } as OrganizationQueryVariables);
  return data.organization ?? null;
}

const UPDATE_ORGANIZATION_MUTATION = `
  mutation UpdateOrganization($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      id
      name
      slug
      description
      updatedAt
    }
  }
`;

export async function updateOrganization(
  id: string,
  updates: { name?: string; description?: string }
): Promise<UpdatedOrganization> {
  const data = await graphql<UpdateOrganizationMutation>(
    UPDATE_ORGANIZATION_MUTATION,
    {
      input: { id, ...updates },
    } as UpdateOrganizationMutationVariables
  );
  return data.updateOrganization;
}

const DELETE_ORGANIZATION_MUTATION = `
  mutation DeleteOrganization($id: ID!) {
    deleteOrganization(id: $id)
  }
`;

export async function deleteOrganization(id: string): Promise<boolean> {
  const data = await graphql<DeleteOrganizationMutation>(
    DELETE_ORGANIZATION_MUTATION,
    { id } as DeleteOrganizationMutationVariables
  );
  return data.deleteOrganization;
}
