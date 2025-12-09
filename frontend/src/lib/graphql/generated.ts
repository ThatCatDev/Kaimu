export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Date: { input: any; output: any; }
  Time: { input: string; output: string; }
  _Any: { input: any; output: any; }
  _FieldSet: { input: any; output: any; }
};

export type AuthPayload = {
  __typename?: 'AuthPayload';
  user: User;
};

export type CreateOrganizationInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
};

export type CreateProjectInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  key: Scalars['String']['input'];
  name: Scalars['String']['input'];
  organizationId: Scalars['ID']['input'];
};

export type LoginInput = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Create a new organization */
  createOrganization: Organization;
  /** Create a new project */
  createProject: Project;
  /** Login with username and password */
  login: AuthPayload;
  /** Logout current user */
  logout: Scalars['Boolean']['output'];
  /** Register a new user */
  register: AuthPayload;
};


export type MutationCreateOrganizationArgs = {
  input: CreateOrganizationInput;
};


export type MutationCreateProjectArgs = {
  input: CreateProjectInput;
};


export type MutationLoginArgs = {
  input: LoginInput;
};


export type MutationRegisterArgs = {
  input: RegisterInput;
};

export type Organization = {
  __typename?: 'Organization';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  members: Array<OrganizationMember>;
  name: Scalars['String']['output'];
  owner: User;
  projects: Array<Project>;
  slug: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type OrganizationMember = {
  __typename?: 'OrganizationMember';
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  role: Scalars['String']['output'];
  user: User;
};

export type Project = {
  __typename?: 'Project';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  key: Scalars['String']['output'];
  name: Scalars['String']['output'];
  organization: Organization;
  updatedAt: Scalars['Time']['output'];
};

export type Query = {
  __typename?: 'Query';
  _service: _Service;
  /** Hello World query */
  helloWorld: Scalars['String']['output'];
  /** Get current authenticated user */
  me?: Maybe<User>;
  /** Get a specific organization by ID */
  organization?: Maybe<Organization>;
  /** Get all organizations for the current user */
  organizations: Array<Organization>;
  /** Get a specific project by ID */
  project?: Maybe<Project>;
};


export type QueryOrganizationArgs = {
  id: Scalars['ID']['input'];
};


export type QueryProjectArgs = {
  id: Scalars['ID']['input'];
};

export type RegisterInput = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type User = {
  __typename?: 'User';
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  username: Scalars['String']['output'];
};

export type _Service = {
  __typename?: '_Service';
  sdl?: Maybe<Scalars['String']['output']>;
};

export type RegisterMutationVariables = Exact<{
  input: RegisterInput;
}>;


export type RegisterMutation = { __typename?: 'Mutation', register: { __typename?: 'AuthPayload', user: { __typename?: 'User', id: string, username: string, createdAt: string } } };

export type LoginMutationVariables = Exact<{
  input: LoginInput;
}>;


export type LoginMutation = { __typename?: 'Mutation', login: { __typename?: 'AuthPayload', user: { __typename?: 'User', id: string, username: string, createdAt: string } } };

export type LogoutMutationVariables = Exact<{ [key: string]: never; }>;


export type LogoutMutation = { __typename?: 'Mutation', logout: boolean };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me?: { __typename?: 'User', id: string, username: string, createdAt: string } | null };

export type CreateOrganizationMutationVariables = Exact<{
  input: CreateOrganizationInput;
}>;


export type CreateOrganizationMutation = { __typename?: 'Mutation', createOrganization: { __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string } };

export type OrganizationsQueryVariables = Exact<{ [key: string]: never; }>;


export type OrganizationsQuery = { __typename?: 'Query', organizations: Array<{ __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string }> };

export type OrganizationQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type OrganizationQuery = { __typename?: 'Query', organization?: { __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string, projects: Array<{ __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string }> } | null };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string } };

export type ProjectQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ProjectQuery = { __typename?: 'Query', project?: { __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string, organization: { __typename?: 'Organization', id: string, name: string, slug: string } } | null };
