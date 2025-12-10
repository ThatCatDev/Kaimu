import { graphql } from './client';
import type {
  RegisterMutation,
  RegisterMutationVariables,
  LoginMutation,
  LoginMutationVariables,
  LogoutMutation,
  MeQuery,
  User,
} from '../graphql/generated';

const REGISTER_MUTATION = `
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      user {
        id
        username
        createdAt
      }
    }
  }
`;

const LOGIN_MUTATION = `
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      user {
        id
        username
        createdAt
      }
    }
  }
`;

const LOGOUT_MUTATION = `
  mutation Logout {
    logout
  }
`;

const ME_QUERY = `
  query Me {
    me {
      id
      username
      email
      displayName
      avatarUrl
      createdAt
    }
  }
`;

export async function register(username: string, password: string): Promise<User> {
  const data = await graphql<RegisterMutation>(REGISTER_MUTATION, {
    input: { username, password },
  } as RegisterMutationVariables);
  return data.register.user;
}

export async function login(username: string, password: string): Promise<User> {
  const data = await graphql<LoginMutation>(LOGIN_MUTATION, {
    input: { username, password },
  } as LoginMutationVariables);
  return data.login.user;
}

export async function logout(): Promise<boolean> {
  const data = await graphql<LogoutMutation>(LOGOUT_MUTATION);
  return data.logout;
}

export async function getMe(): Promise<User | null> {
  const data = await graphql<MeQuery>(ME_QUERY);
  return data.me ?? null;
}
