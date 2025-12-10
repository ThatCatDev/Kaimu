import { useSWR, mutate } from 'sswr';
import type { User } from '../graphql/generated';
import * as authApi from '../api/auth';

const ME_KEY = 'me';

// Fetcher that only runs on client side
async function fetchMe(): Promise<User | null> {
  // Skip fetch during SSR - no cookies available
  if (typeof window === 'undefined') {
    return null;
  }
  return authApi.getMe();
}

export function useMe() {
  return useSWR<User | null>(ME_KEY, fetchMe);
}

export async function login(username: string, password: string): Promise<User> {
  const user = await authApi.login(username, password);
  mutate(ME_KEY, user);
  return user;
}

export async function register(username: string, email: string, password: string): Promise<User> {
  const user = await authApi.register(username, email, password);
  mutate(ME_KEY, user);
  return user;
}

export async function logout(): Promise<void> {
  await authApi.logout();
  mutate(ME_KEY, null);
}
