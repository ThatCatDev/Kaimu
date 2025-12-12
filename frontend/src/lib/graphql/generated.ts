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

export type AssignProjectRoleInput = {
  projectId: Scalars['ID']['input'];
  roleId?: InputMaybe<Scalars['ID']['input']>;
  userId: Scalars['ID']['input'];
};

export type AuthPayload = {
  __typename?: 'AuthPayload';
  user: User;
};

export type Board = {
  __typename?: 'Board';
  activeSprint?: Maybe<Sprint>;
  columns: Array<BoardColumn>;
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  isDefault: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  project: Project;
  sprints: Array<Sprint>;
  updatedAt: Scalars['Time']['output'];
};

export type BoardColumn = {
  __typename?: 'BoardColumn';
  board: Board;
  cards: Array<Card>;
  color?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  isBacklog: Scalars['Boolean']['output'];
  isHidden: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  position: Scalars['Int']['output'];
  updatedAt: Scalars['Time']['output'];
  wipLimit?: Maybe<Scalars['Int']['output']>;
};

export type Card = {
  __typename?: 'Card';
  assignee?: Maybe<User>;
  board: Board;
  column: BoardColumn;
  createdAt: Scalars['Time']['output'];
  createdBy?: Maybe<User>;
  description?: Maybe<Scalars['String']['output']>;
  dueDate?: Maybe<Scalars['Time']['output']>;
  id: Scalars['ID']['output'];
  position: Scalars['Float']['output'];
  priority: CardPriority;
  sprints: Array<Sprint>;
  tags: Array<Tag>;
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export enum CardPriority {
  High = 'HIGH',
  Low = 'LOW',
  Medium = 'MEDIUM',
  None = 'NONE',
  Urgent = 'URGENT'
}

export type ChangeMemberRoleInput = {
  roleId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};

export type CreateBoardInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};

export type CreateCardInput = {
  assigneeId?: InputMaybe<Scalars['ID']['input']>;
  columnId: Scalars['ID']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  dueDate?: InputMaybe<Scalars['Time']['input']>;
  priority?: InputMaybe<CardPriority>;
  tagIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  title: Scalars['String']['input'];
};

export type CreateColumnInput = {
  boardId: Scalars['ID']['input'];
  isBacklog?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
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

export type CreateRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  organizationId: Scalars['ID']['input'];
  permissionCodes: Array<Scalars['String']['input']>;
};

export type CreateSprintInput = {
  boardId: Scalars['ID']['input'];
  endDate?: InputMaybe<Scalars['Time']['input']>;
  goal?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  startDate?: InputMaybe<Scalars['Time']['input']>;
};

export type CreateTagInput = {
  color: Scalars['String']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  projectId: Scalars['ID']['input'];
};

export type Invitation = {
  __typename?: 'Invitation';
  createdAt: Scalars['Time']['output'];
  email: Scalars['String']['output'];
  expiresAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  invitedBy: User;
  organization: Organization;
  role: Role;
  token: Scalars['String']['output'];
};

export type InviteMemberInput = {
  email: Scalars['String']['input'];
  organizationId: Scalars['ID']['input'];
  roleId: Scalars['ID']['input'];
};

export type LoginInput = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type MoveCardInput = {
  afterCardId?: InputMaybe<Scalars['ID']['input']>;
  cardId: Scalars['ID']['input'];
  targetColumnId: Scalars['ID']['input'];
};

export type MoveCardToSprintInput = {
  cardId: Scalars['ID']['input'];
  sprintId: Scalars['ID']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Accept an invitation (for the invited user) */
  acceptInvitation: Organization;
  /** Add a card to a sprint (cards can be in multiple sprints) */
  addCardToSprint: Card;
  /** Assign/change a project-specific role */
  assignProjectRole: ProjectMember;
  /** Cancel a pending invitation */
  cancelInvitation: Scalars['Boolean']['output'];
  /** Change a member's role in an organization */
  changeMemberRole: OrganizationMember;
  /** Complete a sprint (sets status to closed) */
  completeSprint: Sprint;
  /** Create a new board */
  createBoard: Board;
  /** Create a new card */
  createCard: Card;
  /** Create a new column */
  createColumn: BoardColumn;
  /** Create a new organization */
  createOrganization: Organization;
  /** Create a new project */
  createProject: Project;
  /** Create a custom role */
  createRole: Role;
  /** Create a new sprint */
  createSprint: Sprint;
  /** Create a new tag */
  createTag: Tag;
  /** Delete a board */
  deleteBoard: Scalars['Boolean']['output'];
  /** Delete a card */
  deleteCard: Scalars['Boolean']['output'];
  /** Delete a column */
  deleteColumn: Scalars['Boolean']['output'];
  /** Delete an organization */
  deleteOrganization: Scalars['Boolean']['output'];
  /** Delete a project */
  deleteProject: Scalars['Boolean']['output'];
  /** Delete a custom role */
  deleteRole: Scalars['Boolean']['output'];
  /** Delete a sprint */
  deleteSprint: Scalars['Boolean']['output'];
  /** Delete a tag */
  deleteTag: Scalars['Boolean']['output'];
  /** Invite a user to an organization */
  inviteMember: Invitation;
  /** Login with username and password */
  login: AuthPayload;
  /** Logout current user */
  logout: Scalars['Boolean']['output'];
  /** Move a card to a different column */
  moveCard: Card;
  /** Move a card to backlog (remove from all sprints) */
  moveCardToBacklog: Card;
  /** Register a new user (sends verification email) */
  register: AuthPayload;
  /** Remove a card from a sprint */
  removeCardFromSprint: Card;
  /** Remove a member from an organization */
  removeMember: Scalars['Boolean']['output'];
  /** Remove a member from a project */
  removeProjectMember: Scalars['Boolean']['output'];
  /** Reorder columns */
  reorderColumns: Array<BoardColumn>;
  /** Resend an invitation */
  resendInvitation: Invitation;
  /** Resend verification email */
  resendVerificationEmail: Scalars['Boolean']['output'];
  /** Set all sprints for a card (replaces existing sprint assignments) */
  setCardSprints: Card;
  /** Start a sprint (sets status to active) */
  startSprint: Sprint;
  /** Toggle column visibility */
  toggleColumnVisibility: BoardColumn;
  /** Update a board */
  updateBoard: Board;
  /** Update a card */
  updateCard: Card;
  /** Update a column */
  updateColumn: BoardColumn;
  /** Update current user's profile */
  updateMe: User;
  /** Update an organization */
  updateOrganization: Organization;
  /** Update a project */
  updateProject: Project;
  /** Update a custom role */
  updateRole: Role;
  /** Update a sprint */
  updateSprint: Sprint;
  /** Update a tag */
  updateTag: Tag;
  /** Verify email with token */
  verifyEmail: AuthPayload;
};


export type MutationAcceptInvitationArgs = {
  token: Scalars['String']['input'];
};


export type MutationAddCardToSprintArgs = {
  input: MoveCardToSprintInput;
};


export type MutationAssignProjectRoleArgs = {
  input: AssignProjectRoleInput;
};


export type MutationCancelInvitationArgs = {
  id: Scalars['ID']['input'];
};


export type MutationChangeMemberRoleArgs = {
  input: ChangeMemberRoleInput;
  organizationId: Scalars['ID']['input'];
};


export type MutationCompleteSprintArgs = {
  id: Scalars['ID']['input'];
  moveIncompleteToBacklog?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationCreateBoardArgs = {
  input: CreateBoardInput;
};


export type MutationCreateCardArgs = {
  input: CreateCardInput;
};


export type MutationCreateColumnArgs = {
  input: CreateColumnInput;
};


export type MutationCreateOrganizationArgs = {
  input: CreateOrganizationInput;
};


export type MutationCreateProjectArgs = {
  input: CreateProjectInput;
};


export type MutationCreateRoleArgs = {
  input: CreateRoleInput;
};


export type MutationCreateSprintArgs = {
  input: CreateSprintInput;
};


export type MutationCreateTagArgs = {
  input: CreateTagInput;
};


export type MutationDeleteBoardArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteCardArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteColumnArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteOrganizationArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteProjectArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteRoleArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteSprintArgs = {
  id: Scalars['ID']['input'];
};


export type MutationDeleteTagArgs = {
  id: Scalars['ID']['input'];
};


export type MutationInviteMemberArgs = {
  input: InviteMemberInput;
};


export type MutationLoginArgs = {
  input: LoginInput;
};


export type MutationMoveCardArgs = {
  input: MoveCardInput;
};


export type MutationMoveCardToBacklogArgs = {
  cardId: Scalars['ID']['input'];
};


export type MutationRegisterArgs = {
  input: RegisterInput;
};


export type MutationRemoveCardFromSprintArgs = {
  input: MoveCardToSprintInput;
};


export type MutationRemoveMemberArgs = {
  organizationId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationRemoveProjectMemberArgs = {
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationReorderColumnsArgs = {
  input: ReorderColumnsInput;
};


export type MutationResendInvitationArgs = {
  id: Scalars['ID']['input'];
};


export type MutationSetCardSprintsArgs = {
  cardId: Scalars['ID']['input'];
  sprintIds: Array<Scalars['ID']['input']>;
};


export type MutationStartSprintArgs = {
  id: Scalars['ID']['input'];
};


export type MutationToggleColumnVisibilityArgs = {
  id: Scalars['ID']['input'];
};


export type MutationUpdateBoardArgs = {
  input: UpdateBoardInput;
};


export type MutationUpdateCardArgs = {
  input: UpdateCardInput;
};


export type MutationUpdateColumnArgs = {
  input: UpdateColumnInput;
};


export type MutationUpdateMeArgs = {
  input: UpdateMeInput;
};


export type MutationUpdateOrganizationArgs = {
  input: UpdateOrganizationInput;
};


export type MutationUpdateProjectArgs = {
  input: UpdateProjectInput;
};


export type MutationUpdateRoleArgs = {
  input: UpdateRoleInput;
};


export type MutationUpdateSprintArgs = {
  id: Scalars['ID']['input'];
  input: UpdateSprintInput;
};


export type MutationUpdateTagArgs = {
  input: UpdateTagInput;
};


export type MutationVerifyEmailArgs = {
  token: Scalars['String']['input'];
};

export type OidcProvider = {
  __typename?: 'OIDCProvider';
  name: Scalars['String']['output'];
  slug: Scalars['String']['output'];
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
  /** @deprecated Use role field instead */
  legacyRole: Scalars['String']['output'];
  role: Role;
  user: User;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor?: Maybe<Scalars['String']['output']>;
  totalCount: Scalars['Int']['output'];
};

export type Permission = {
  __typename?: 'Permission';
  code: Scalars['String']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  resourceType: Scalars['String']['output'];
};

export type Project = {
  __typename?: 'Project';
  boards: Array<Board>;
  createdAt: Scalars['Time']['output'];
  defaultBoard?: Maybe<Board>;
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  key: Scalars['String']['output'];
  name: Scalars['String']['output'];
  organization: Organization;
  tags: Array<Tag>;
  updatedAt: Scalars['Time']['output'];
};

export type ProjectMember = {
  __typename?: 'ProjectMember';
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  project: Project;
  role?: Maybe<Role>;
  user: User;
};

export type Query = {
  __typename?: 'Query';
  _service: _Service;
  /** Get the active sprint for a board */
  activeSprint?: Maybe<Sprint>;
  /** Get backlog cards (cards not assigned to any sprint) */
  backlogCards: Array<Card>;
  /** Get a board by ID */
  board?: Maybe<Board>;
  /** Get all boards for a project */
  boards: Array<Board>;
  /** Get a card by ID */
  card?: Maybe<Card>;
  /** Get closed sprints for a board (paginated) */
  closedSprints: SprintConnection;
  /** Get future sprints for a board */
  futureSprints: Array<Sprint>;
  /** Check if current user has a specific permission */
  hasPermission: Scalars['Boolean']['output'];
  /** Hello World query */
  helloWorld: Scalars['String']['output'];
  /** Get pending invitations for an organization */
  invitations: Array<Invitation>;
  /** Get current authenticated user */
  me?: Maybe<User>;
  /** Get all cards assigned to the current user */
  myCards: Array<Card>;
  /** Get current user's permissions for a resource */
  myPermissions: Array<Scalars['String']['output']>;
  /** Get available OIDC providers */
  oidcProviders: Array<OidcProvider>;
  /** Get a specific organization by ID */
  organization?: Maybe<Organization>;
  /** Get organization members with roles */
  organizationMembers: Array<OrganizationMember>;
  /** Get all organizations for the current user */
  organizations: Array<Organization>;
  /** Get all available permissions */
  permissions: Array<Permission>;
  /** Get a specific project by ID */
  project?: Maybe<Project>;
  /** Get project members */
  projectMembers: Array<ProjectMember>;
  /** Get a specific role by ID */
  role?: Maybe<Role>;
  /** Get roles for an organization (includes system roles) */
  roles: Array<Role>;
  /** Search across organizations, projects, boards, cards, and users */
  search: SearchResults;
  /** Get a sprint by ID */
  sprint?: Maybe<Sprint>;
  /** Get cards in a sprint */
  sprintCards: Array<Card>;
  /** Get all sprints for a board */
  sprints: Array<Sprint>;
  /** Get all tags for a project */
  tags: Array<Tag>;
};


export type QueryActiveSprintArgs = {
  boardId: Scalars['ID']['input'];
};


export type QueryBacklogCardsArgs = {
  boardId: Scalars['ID']['input'];
};


export type QueryBoardArgs = {
  id: Scalars['ID']['input'];
};


export type QueryBoardsArgs = {
  projectId: Scalars['ID']['input'];
};


export type QueryCardArgs = {
  id: Scalars['ID']['input'];
};


export type QueryClosedSprintsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  boardId: Scalars['ID']['input'];
  first?: InputMaybe<Scalars['Int']['input']>;
};


export type QueryFutureSprintsArgs = {
  boardId: Scalars['ID']['input'];
};


export type QueryHasPermissionArgs = {
  permission: Scalars['String']['input'];
  resourceId: Scalars['ID']['input'];
  resourceType: Scalars['String']['input'];
};


export type QueryInvitationsArgs = {
  organizationId: Scalars['ID']['input'];
};


export type QueryMyPermissionsArgs = {
  resourceId: Scalars['ID']['input'];
  resourceType: Scalars['String']['input'];
};


export type QueryOrganizationArgs = {
  id: Scalars['ID']['input'];
};


export type QueryOrganizationMembersArgs = {
  organizationId: Scalars['ID']['input'];
};


export type QueryProjectArgs = {
  id: Scalars['ID']['input'];
};


export type QueryProjectMembersArgs = {
  projectId: Scalars['ID']['input'];
};


export type QueryRoleArgs = {
  id: Scalars['ID']['input'];
};


export type QueryRolesArgs = {
  organizationId: Scalars['ID']['input'];
};


export type QuerySearchArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>;
  query: Scalars['String']['input'];
  scope?: InputMaybe<SearchScope>;
};


export type QuerySprintArgs = {
  id: Scalars['ID']['input'];
};


export type QuerySprintCardsArgs = {
  sprintId: Scalars['ID']['input'];
};


export type QuerySprintsArgs = {
  boardId: Scalars['ID']['input'];
};


export type QueryTagsArgs = {
  projectId: Scalars['ID']['input'];
};

export type RegisterInput = {
  email: Scalars['String']['input'];
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type ReorderColumnsInput = {
  boardId: Scalars['ID']['input'];
  columnIds: Array<Scalars['ID']['input']>;
};

export type Role = {
  __typename?: 'Role';
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  isSystem: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  permissions: Array<Permission>;
  scope: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export enum SearchEntityType {
  Board = 'BOARD',
  Card = 'CARD',
  Organization = 'ORGANIZATION',
  Project = 'PROJECT',
  User = 'USER'
}

export type SearchResult = {
  __typename?: 'SearchResult';
  boardId?: Maybe<Scalars['ID']['output']>;
  boardName?: Maybe<Scalars['String']['output']>;
  description?: Maybe<Scalars['String']['output']>;
  highlight: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  organizationId: Scalars['ID']['output'];
  organizationName: Scalars['String']['output'];
  projectId?: Maybe<Scalars['ID']['output']>;
  projectName?: Maybe<Scalars['String']['output']>;
  score: Scalars['Float']['output'];
  title: Scalars['String']['output'];
  type: SearchEntityType;
  url: Scalars['String']['output'];
};

export type SearchResults = {
  __typename?: 'SearchResults';
  query: Scalars['String']['output'];
  results: Array<SearchResult>;
  totalCount: Scalars['Int']['output'];
};

export type SearchScope = {
  organizationId?: InputMaybe<Scalars['ID']['input']>;
  projectId?: InputMaybe<Scalars['ID']['input']>;
};

export type Sprint = {
  __typename?: 'Sprint';
  board: Board;
  cards: Array<Card>;
  createdAt: Scalars['Time']['output'];
  createdBy?: Maybe<User>;
  endDate?: Maybe<Scalars['Time']['output']>;
  goal?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  position: Scalars['Int']['output'];
  startDate?: Maybe<Scalars['Time']['output']>;
  status: SprintStatus;
  updatedAt: Scalars['Time']['output'];
};

export type SprintConnection = {
  __typename?: 'SprintConnection';
  edges: Array<SprintEdge>;
  pageInfo: PageInfo;
};

export type SprintEdge = {
  __typename?: 'SprintEdge';
  cursor: Scalars['String']['output'];
  node: Sprint;
};

export enum SprintStatus {
  Active = 'ACTIVE',
  Closed = 'CLOSED',
  Future = 'FUTURE'
}

export type Tag = {
  __typename?: 'Tag';
  color: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  description?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  project: Project;
};

export type UpdateBoardInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateCardInput = {
  assigneeId?: InputMaybe<Scalars['ID']['input']>;
  clearAssignee?: InputMaybe<Scalars['Boolean']['input']>;
  clearDueDate?: InputMaybe<Scalars['Boolean']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  dueDate?: InputMaybe<Scalars['Time']['input']>;
  id: Scalars['ID']['input'];
  priority?: InputMaybe<CardPriority>;
  tagIds?: InputMaybe<Array<Scalars['ID']['input']>>;
  title?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateColumnInput = {
  clearWipLimit?: InputMaybe<Scalars['Boolean']['input']>;
  color?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  name?: InputMaybe<Scalars['String']['input']>;
  wipLimit?: InputMaybe<Scalars['Int']['input']>;
};

export type UpdateMeInput = {
  displayName?: InputMaybe<Scalars['String']['input']>;
  email?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateOrganizationInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateProjectInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  key?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateRoleInput = {
  description?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  name?: InputMaybe<Scalars['String']['input']>;
  permissionCodes?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type UpdateSprintInput = {
  endDate?: InputMaybe<Scalars['Time']['input']>;
  goal?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  startDate?: InputMaybe<Scalars['Time']['input']>;
};

export type UpdateTagInput = {
  color?: InputMaybe<Scalars['String']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  id: Scalars['ID']['input'];
  name?: InputMaybe<Scalars['String']['input']>;
};

export type User = {
  __typename?: 'User';
  avatarUrl?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['Time']['output'];
  displayName?: Maybe<Scalars['String']['output']>;
  email?: Maybe<Scalars['String']['output']>;
  emailVerified: Scalars['Boolean']['output'];
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

export type UpdateMeMutationVariables = Exact<{
  input: UpdateMeInput;
}>;


export type UpdateMeMutation = { __typename?: 'Mutation', updateMe: { __typename?: 'User', id: string, username: string, email?: string | null, displayName?: string | null, avatarUrl?: string | null, createdAt: string } };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me?: { __typename?: 'User', id: string, username: string, email?: string | null, displayName?: string | null, avatarUrl?: string | null, createdAt: string } | null };

export type OidcProvidersQueryVariables = Exact<{ [key: string]: never; }>;


export type OidcProvidersQuery = { __typename?: 'Query', oidcProviders: Array<{ __typename?: 'OIDCProvider', slug: string, name: string }> };

export type BoardQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type BoardQuery = { __typename?: 'Query', board?: { __typename?: 'Board', id: string, name: string, description?: string | null, isDefault: boolean, createdAt: string, updatedAt: string, project: { __typename?: 'Project', id: string, name: string, key: string, organization: { __typename?: 'Organization', id: string, name: string, slug: string } }, columns: Array<{ __typename?: 'BoardColumn', id: string, name: string, position: number, isBacklog: boolean, isHidden: boolean, color?: string | null, wipLimit?: number | null, cards: Array<{ __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null } | null }> }> } | null };

export type BoardsQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type BoardsQuery = { __typename?: 'Query', boards: Array<{ __typename?: 'Board', id: string, name: string, description?: string | null, isDefault: boolean, createdAt: string }> };

export type ProjectDefaultBoardQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type ProjectDefaultBoardQuery = { __typename?: 'Query', boards: Array<{ __typename?: 'Board', id: string, name: string, isDefault: boolean }> };

export type TagsQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type TagsQuery = { __typename?: 'Query', tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string, description?: string | null, createdAt: string }> };

export type MyCardsQueryVariables = Exact<{ [key: string]: never; }>;


export type MyCardsQuery = { __typename?: 'Query', myCards: Array<{ __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, board: { __typename?: 'Board', id: string, name: string, project: { __typename?: 'Project', id: string, name: string, key: string } }, column: { __typename?: 'BoardColumn', id: string, name: string }, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }> }> };

export type CreateBoardMutationVariables = Exact<{
  input: CreateBoardInput;
}>;


export type CreateBoardMutation = { __typename?: 'Mutation', createBoard: { __typename?: 'Board', id: string, name: string, description?: string | null, isDefault: boolean, createdAt: string } };

export type UpdateBoardMutationVariables = Exact<{
  input: UpdateBoardInput;
}>;


export type UpdateBoardMutation = { __typename?: 'Mutation', updateBoard: { __typename?: 'Board', id: string, name: string, description?: string | null, updatedAt: string } };

export type DeleteBoardMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteBoardMutation = { __typename?: 'Mutation', deleteBoard: boolean };

export type CreateColumnMutationVariables = Exact<{
  input: CreateColumnInput;
}>;


export type CreateColumnMutation = { __typename?: 'Mutation', createColumn: { __typename?: 'BoardColumn', id: string, name: string, position: number, isBacklog: boolean, isHidden: boolean, color?: string | null, wipLimit?: number | null, createdAt: string } };

export type UpdateColumnMutationVariables = Exact<{
  input: UpdateColumnInput;
}>;


export type UpdateColumnMutation = { __typename?: 'Mutation', updateColumn: { __typename?: 'BoardColumn', id: string, name: string, color?: string | null, wipLimit?: number | null, updatedAt: string } };

export type ReorderColumnsMutationVariables = Exact<{
  input: ReorderColumnsInput;
}>;


export type ReorderColumnsMutation = { __typename?: 'Mutation', reorderColumns: Array<{ __typename?: 'BoardColumn', id: string, position: number }> };

export type ToggleColumnVisibilityMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ToggleColumnVisibilityMutation = { __typename?: 'Mutation', toggleColumnVisibility: { __typename?: 'BoardColumn', id: string, isHidden: boolean } };

export type DeleteColumnMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteColumnMutation = { __typename?: 'Mutation', deleteColumn: boolean };

export type CreateCardMutationVariables = Exact<{
  input: CreateCardInput;
}>;


export type CreateCardMutation = { __typename?: 'Mutation', createCard: { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null } | null } };

export type UpdateCardMutationVariables = Exact<{
  input: UpdateCardInput;
}>;


export type UpdateCardMutation = { __typename?: 'Mutation', updateCard: { __typename?: 'Card', id: string, title: string, description?: string | null, priority: CardPriority, dueDate?: string | null, updatedAt: string, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null } | null } };

export type MoveCardMutationVariables = Exact<{
  input: MoveCardInput;
}>;


export type MoveCardMutation = { __typename?: 'Mutation', moveCard: { __typename?: 'Card', id: string, position: number, column: { __typename?: 'BoardColumn', id: string } } };

export type DeleteCardMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteCardMutation = { __typename?: 'Mutation', deleteCard: boolean };

export type CreateTagMutationVariables = Exact<{
  input: CreateTagInput;
}>;


export type CreateTagMutation = { __typename?: 'Mutation', createTag: { __typename?: 'Tag', id: string, name: string, color: string, description?: string | null, createdAt: string } };

export type UpdateTagMutationVariables = Exact<{
  input: UpdateTagInput;
}>;


export type UpdateTagMutation = { __typename?: 'Mutation', updateTag: { __typename?: 'Tag', id: string, name: string, color: string, description?: string | null } };

export type DeleteTagMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteTagMutation = { __typename?: 'Mutation', deleteTag: boolean };

export type CreateOrganizationMutationVariables = Exact<{
  input: CreateOrganizationInput;
}>;


export type CreateOrganizationMutation = { __typename?: 'Mutation', createOrganization: { __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string } };

export type OrganizationsQueryVariables = Exact<{ [key: string]: never; }>;


export type OrganizationsQuery = { __typename?: 'Query', organizations: Array<{ __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string, projects: Array<{ __typename?: 'Project', id: string, name: string, key: string, boards: Array<{ __typename?: 'Board', id: string, name: string, isDefault: boolean }> }> }> };

export type OrganizationQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type OrganizationQuery = { __typename?: 'Query', organization?: { __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, createdAt: string, updatedAt: string, projects: Array<{ __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string, boards: Array<{ __typename?: 'Board', id: string }> }> } | null };

export type UpdateOrganizationMutationVariables = Exact<{
  input: UpdateOrganizationInput;
}>;


export type UpdateOrganizationMutation = { __typename?: 'Mutation', updateOrganization: { __typename?: 'Organization', id: string, name: string, slug: string, description?: string | null, updatedAt: string } };

export type DeleteOrganizationMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteOrganizationMutation = { __typename?: 'Mutation', deleteOrganization: boolean };

export type CreateProjectMutationVariables = Exact<{
  input: CreateProjectInput;
}>;


export type CreateProjectMutation = { __typename?: 'Mutation', createProject: { __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string } };

export type ProjectQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ProjectQuery = { __typename?: 'Query', project?: { __typename?: 'Project', id: string, name: string, key: string, description?: string | null, createdAt: string, updatedAt: string, organization: { __typename?: 'Organization', id: string, name: string, slug: string } } | null };

export type UpdateProjectMutationVariables = Exact<{
  input: UpdateProjectInput;
}>;


export type UpdateProjectMutation = { __typename?: 'Mutation', updateProject: { __typename?: 'Project', id: string, name: string, key: string, description?: string | null, updatedAt: string } };

export type DeleteProjectMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteProjectMutation = { __typename?: 'Mutation', deleteProject: boolean };

export type PermissionsQueryVariables = Exact<{ [key: string]: never; }>;


export type PermissionsQuery = { __typename?: 'Query', permissions: Array<{ __typename?: 'Permission', id: string, code: string, name: string, description?: string | null, resourceType: string }> };

export type RolesQueryVariables = Exact<{
  organizationId: Scalars['ID']['input'];
}>;


export type RolesQuery = { __typename?: 'Query', roles: Array<{ __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean, scope: string, createdAt: string, updatedAt: string, permissions: Array<{ __typename?: 'Permission', id: string, code: string, name: string }> }> };

export type RoleQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type RoleQuery = { __typename?: 'Query', role?: { __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean, scope: string, createdAt: string, updatedAt: string, permissions: Array<{ __typename?: 'Permission', id: string, code: string, name: string, description?: string | null, resourceType: string }> } | null };

export type OrganizationMembersQueryVariables = Exact<{
  organizationId: Scalars['ID']['input'];
}>;


export type OrganizationMembersQuery = { __typename?: 'Query', organizationMembers: Array<{ __typename?: 'OrganizationMember', id: string, legacyRole: string, createdAt: string, user: { __typename?: 'User', id: string, email?: string | null, displayName?: string | null }, role: { __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean } }> };

export type ProjectMembersQueryVariables = Exact<{
  projectId: Scalars['ID']['input'];
}>;


export type ProjectMembersQuery = { __typename?: 'Query', projectMembers: Array<{ __typename?: 'ProjectMember', id: string, createdAt: string, user: { __typename?: 'User', id: string, email?: string | null, displayName?: string | null }, role?: { __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean } | null, project: { __typename?: 'Project', id: string, name: string } }> };

export type InvitationsQueryVariables = Exact<{
  organizationId: Scalars['ID']['input'];
}>;


export type InvitationsQuery = { __typename?: 'Query', invitations: Array<{ __typename?: 'Invitation', id: string, email: string, expiresAt: string, createdAt: string, role: { __typename?: 'Role', id: string, name: string }, invitedBy: { __typename?: 'User', id: string, email?: string | null, displayName?: string | null } }> };

export type HasPermissionQueryVariables = Exact<{
  permission: Scalars['String']['input'];
  resourceType: Scalars['String']['input'];
  resourceId: Scalars['ID']['input'];
}>;


export type HasPermissionQuery = { __typename?: 'Query', hasPermission: boolean };

export type MyPermissionsQueryVariables = Exact<{
  resourceType: Scalars['String']['input'];
  resourceId: Scalars['ID']['input'];
}>;


export type MyPermissionsQuery = { __typename?: 'Query', myPermissions: Array<string> };

export type CreateRoleMutationVariables = Exact<{
  input: CreateRoleInput;
}>;


export type CreateRoleMutation = { __typename?: 'Mutation', createRole: { __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean, scope: string, createdAt: string, updatedAt: string, permissions: Array<{ __typename?: 'Permission', id: string, code: string, name: string }> } };

export type UpdateRoleMutationVariables = Exact<{
  input: UpdateRoleInput;
}>;


export type UpdateRoleMutation = { __typename?: 'Mutation', updateRole: { __typename?: 'Role', id: string, name: string, description?: string | null, isSystem: boolean, scope: string, updatedAt: string, permissions: Array<{ __typename?: 'Permission', id: string, code: string, name: string }> } };

export type DeleteRoleMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteRoleMutation = { __typename?: 'Mutation', deleteRole: boolean };

export type InviteMemberMutationVariables = Exact<{
  input: InviteMemberInput;
}>;


export type InviteMemberMutation = { __typename?: 'Mutation', inviteMember: { __typename?: 'Invitation', id: string, email: string, token: string, expiresAt: string, createdAt: string, role: { __typename?: 'Role', id: string, name: string } } };

export type CancelInvitationMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type CancelInvitationMutation = { __typename?: 'Mutation', cancelInvitation: boolean };

export type ResendInvitationMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type ResendInvitationMutation = { __typename?: 'Mutation', resendInvitation: { __typename?: 'Invitation', id: string, email: string, expiresAt: string, createdAt: string } };

export type AcceptInvitationMutationVariables = Exact<{
  token: Scalars['String']['input'];
}>;


export type AcceptInvitationMutation = { __typename?: 'Mutation', acceptInvitation: { __typename?: 'Organization', id: string, name: string, slug: string } };

export type ChangeMemberRoleMutationVariables = Exact<{
  organizationId: Scalars['ID']['input'];
  input: ChangeMemberRoleInput;
}>;


export type ChangeMemberRoleMutation = { __typename?: 'Mutation', changeMemberRole: { __typename?: 'OrganizationMember', id: string, legacyRole: string, user: { __typename?: 'User', id: string, email?: string | null, displayName?: string | null }, role: { __typename?: 'Role', id: string, name: string } } };

export type RemoveMemberMutationVariables = Exact<{
  organizationId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
}>;


export type RemoveMemberMutation = { __typename?: 'Mutation', removeMember: boolean };

export type AssignProjectRoleMutationVariables = Exact<{
  input: AssignProjectRoleInput;
}>;


export type AssignProjectRoleMutation = { __typename?: 'Mutation', assignProjectRole: { __typename?: 'ProjectMember', id: string, user: { __typename?: 'User', id: string, email?: string | null, displayName?: string | null }, role?: { __typename?: 'Role', id: string, name: string } | null, project: { __typename?: 'Project', id: string, name: string } } };

export type RemoveProjectMemberMutationVariables = Exact<{
  projectId: Scalars['ID']['input'];
  userId: Scalars['ID']['input'];
}>;


export type RemoveProjectMemberMutation = { __typename?: 'Mutation', removeProjectMember: boolean };

export type SprintFieldsFragment = { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string };

export type CardFieldsFragment = { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } };

export type GetSprintQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetSprintQuery = { __typename?: 'Query', sprint?: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string, board: { __typename?: 'Board', id: string, name: string } } | null };

export type GetSprintsQueryVariables = Exact<{
  boardId: Scalars['ID']['input'];
}>;


export type GetSprintsQuery = { __typename?: 'Query', sprints: Array<{ __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string }> };

export type GetActiveSprintQueryVariables = Exact<{
  boardId: Scalars['ID']['input'];
}>;


export type GetActiveSprintQuery = { __typename?: 'Query', activeSprint?: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } | null };

export type GetFutureSprintsQueryVariables = Exact<{
  boardId: Scalars['ID']['input'];
}>;


export type GetFutureSprintsQuery = { __typename?: 'Query', futureSprints: Array<{ __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string }> };

export type GetClosedSprintsQueryVariables = Exact<{
  boardId: Scalars['ID']['input'];
  first?: InputMaybe<Scalars['Int']['input']>;
  after?: InputMaybe<Scalars['String']['input']>;
}>;


export type GetClosedSprintsQuery = { __typename?: 'Query', closedSprints: { __typename?: 'SprintConnection', edges: Array<{ __typename?: 'SprintEdge', cursor: string, node: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor?: string | null, endCursor?: string | null, totalCount: number } } };

export type GetSprintCardsQueryVariables = Exact<{
  sprintId: Scalars['ID']['input'];
}>;


export type GetSprintCardsQuery = { __typename?: 'Query', sprintCards: Array<{ __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } }> };

export type GetBacklogCardsQueryVariables = Exact<{
  boardId: Scalars['ID']['input'];
}>;


export type GetBacklogCardsQuery = { __typename?: 'Query', backlogCards: Array<{ __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } }> };

export type CreateSprintMutationVariables = Exact<{
  input: CreateSprintInput;
}>;


export type CreateSprintMutation = { __typename?: 'Mutation', createSprint: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } };

export type UpdateSprintMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  input: UpdateSprintInput;
}>;


export type UpdateSprintMutation = { __typename?: 'Mutation', updateSprint: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } };

export type DeleteSprintMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type DeleteSprintMutation = { __typename?: 'Mutation', deleteSprint: boolean };

export type StartSprintMutationVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type StartSprintMutation = { __typename?: 'Mutation', startSprint: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } };

export type CompleteSprintMutationVariables = Exact<{
  id: Scalars['ID']['input'];
  moveIncompleteToBacklog?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type CompleteSprintMutation = { __typename?: 'Mutation', completeSprint: { __typename?: 'Sprint', id: string, name: string, goal?: string | null, startDate?: string | null, endDate?: string | null, status: SprintStatus, position: number, createdAt: string, updatedAt: string } };

export type AddCardToSprintMutationVariables = Exact<{
  input: MoveCardToSprintInput;
}>;


export type AddCardToSprintMutation = { __typename?: 'Mutation', addCardToSprint: { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, sprints: Array<{ __typename?: 'Sprint', id: string, name: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } } };

export type RemoveCardFromSprintMutationVariables = Exact<{
  input: MoveCardToSprintInput;
}>;


export type RemoveCardFromSprintMutation = { __typename?: 'Mutation', removeCardFromSprint: { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, sprints: Array<{ __typename?: 'Sprint', id: string, name: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } } };

export type SetCardSprintsMutationVariables = Exact<{
  cardId: Scalars['ID']['input'];
  sprintIds: Array<Scalars['ID']['input']> | Scalars['ID']['input'];
}>;


export type SetCardSprintsMutation = { __typename?: 'Mutation', setCardSprints: { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, sprints: Array<{ __typename?: 'Sprint', id: string, name: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } } };

export type MoveCardToBacklogMutationVariables = Exact<{
  cardId: Scalars['ID']['input'];
}>;


export type MoveCardToBacklogMutation = { __typename?: 'Mutation', moveCardToBacklog: { __typename?: 'Card', id: string, title: string, description?: string | null, position: number, priority: CardPriority, dueDate?: string | null, createdAt: string, updatedAt: string, sprints: Array<{ __typename?: 'Sprint', id: string, name: string }>, assignee?: { __typename?: 'User', id: string, username: string, displayName?: string | null, avatarUrl?: string | null } | null, tags: Array<{ __typename?: 'Tag', id: string, name: string, color: string }>, column: { __typename?: 'BoardColumn', id: string, name: string } } };
