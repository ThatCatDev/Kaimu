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

export type LoginInput = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Login with username and password */
  login: AuthPayload;
  /** Logout current user */
  logout: Scalars['Boolean']['output'];
  /** Register a new user */
  register: AuthPayload;
};


export type MutationLoginArgs = {
  input: LoginInput;
};


export type MutationRegisterArgs = {
  input: RegisterInput;
};

export type Query = {
  __typename?: 'Query';
  _service: _Service;
  /** Hello World query */
  helloWorld: Scalars['String']['output'];
  /** Get current authenticated user */
  me?: Maybe<User>;
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
