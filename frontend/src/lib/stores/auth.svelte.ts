import { useSWR, mutate } from 'sswr';
import type { User } from '../graphql/generated';
import * as authApi from '../api/auth';
import { clearTokenExpiration, setTokenExpiration, tryRefreshToken } from '../api/client';

const ME_KEY = 'me';

// Default access token expiration in seconds (5 minutes)
const DEFAULT_ACCESS_TOKEN_EXPIRATION = 300;

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
  // Set token expiration (default 5 minutes)
  setTokenExpiration(DEFAULT_ACCESS_TOKEN_EXPIRATION);
  return user;
}

export async function register(username: string, email: string, password: string): Promise<User> {
  const user = await authApi.register(username, email, password);
  mutate(ME_KEY, user);
  // Set token expiration (default 5 minutes)
  setTokenExpiration(DEFAULT_ACCESS_TOKEN_EXPIRATION);
  return user;
}

export async function logout(): Promise<void> {
  await authApi.logout();
  mutate(ME_KEY, null);
  clearTokenExpiration();
}

/**
 * Check if the user's session is still valid by attempting to refresh the token
 * Returns true if the session is valid, false otherwise
 */
export async function checkSession(): Promise<boolean> {
  if (typeof window === 'undefined') {
    return false;
  }
  return tryRefreshToken();
}
