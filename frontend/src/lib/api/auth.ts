import { graphql } from './client';
import type {
  RegisterMutation,
  RegisterMutationVariables,
  LoginMutation,
  LoginMutationVariables,
  LogoutMutation,
  UpdateMeMutation,
  UpdateMeMutationVariables,
  MeQuery,
  User,
  UpdateMeInput,
} from '../graphql/generated';

const REGISTER_MUTATION = `
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      user {
        id
        username
        email
        emailVerified
        displayName
        avatarUrl
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
        email
        emailVerified
        displayName
        avatarUrl
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

const UPDATE_ME_MUTATION = `
  mutation UpdateMe($input: UpdateMeInput!) {
    updateMe(input: $input) {
      id
      username
      email
      displayName
      avatarUrl
      createdAt
    }
  }
`;

const ME_QUERY = `
  query Me {
    me {
      id
      username
      email
      emailVerified
      displayName
      avatarUrl
      createdAt
    }
  }
`;

const VERIFY_EMAIL_MUTATION = `
  mutation VerifyEmail($token: String!) {
    verifyEmail(token: $token) {
      user {
        id
        username
        email
        emailVerified
        displayName
        avatarUrl
        createdAt
      }
    }
  }
`;

const RESEND_VERIFICATION_EMAIL_MUTATION = `
  mutation ResendVerificationEmail {
    resendVerificationEmail
  }
`;

export async function register(username: string, email: string, password: string): Promise<User> {
  const data = await graphql<RegisterMutation>(REGISTER_MUTATION, {
    input: { username, email, password },
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

export async function updateMe(input: UpdateMeInput): Promise<User> {
  const data = await graphql<UpdateMeMutation>(UPDATE_ME_MUTATION, {
    input,
  } as UpdateMeMutationVariables);
  return data.updateMe;
}

export async function verifyEmail(token: string): Promise<User> {
  const data = await graphql<{ verifyEmail: { user: User } }>(VERIFY_EMAIL_MUTATION, {
    token,
  });
  return data.verifyEmail.user;
}

export async function resendVerificationEmail(): Promise<boolean> {
  const data = await graphql<{ resendVerificationEmail: boolean }>(RESEND_VERIFICATION_EMAIL_MUTATION);
  return data.resendVerificationEmail;
}
