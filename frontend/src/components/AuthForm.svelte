<script lang="ts">
  import { login, register } from '../lib/stores/auth.svelte';
  import { getOIDCProviders, getOIDCLoginURL } from '../lib/api/oidc';
  import type { OidcProvider } from '../lib/graphql/generated';
  import { Input, Button } from './ui';

  interface Props {
    mode: 'login' | 'register';
  }

  let { mode }: Props = $props();

  let username = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let error = $state<string | null>(null);
  let loading = $state(false);
  let oidcProviders = $state<OidcProvider[]>([]);

  // Load OIDC providers on mount
  $effect(() => {
    getOIDCProviders()
      .then((providers) => {
        oidcProviders = providers;
      })
      .catch((err) => {
        console.error('Failed to load OIDC providers:', err);
      });
  });

  const isLogin = $derived(mode === 'login');
  const title = $derived(isLogin ? 'Sign in to your account' : 'Create your account');
  const submitText = $derived(isLogin ? 'Sign in' : 'Register');
  const altText = $derived(isLogin ? "Don't have an account?" : 'Already have an account?');
  const altLink = $derived(isLogin ? '/register' : '/login');
  const altLinkText = $derived(isLogin ? 'Register' : 'Sign in');

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = null;

    if (!username.trim() || !password.trim()) {
      error = 'Please fill in all fields';
      return;
    }

    if (!isLogin && password !== confirmPassword) {
      error = 'Passwords do not match';
      return;
    }

    loading = true;
    try {
      if (isLogin) {
        await login(username, password);
      } else {
        await register(username, password);
      }
      window.location.href = '/';
    } catch (e) {
      error = e instanceof Error ? e.message : 'An error occurred';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md w-full space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
        {title}
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        {altText}
        <a href={altLink} class="font-medium text-indigo-600 hover:text-indigo-500">
          {altLinkText}
        </a>
      </p>
    </div>

    {#if oidcProviders.length > 0}
      <div class="mt-8 space-y-4">
        {#each oidcProviders as provider}
          <a
            href={getOIDCLoginURL(provider.slug)}
            class="w-full flex justify-center items-center gap-2 py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Continue with {provider.name}
          </a>
        {/each}
      </div>

      <div class="relative mt-6">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-gray-300"></div>
        </div>
        <div class="relative flex justify-center text-sm">
          <span class="px-2 bg-gray-50 text-gray-500">Or continue with</span>
        </div>
      </div>
    {/if}

    <form class="mt-8 space-y-6" onsubmit={handleSubmit}>
      {#if error}
        <div class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-700">{error}</p>
        </div>
      {/if}

      <div class="space-y-4">
        <Input
          id="username"
          label="Username"
          type="text"
          autocomplete="username"
          bind:value={username}
          placeholder="Enter your username"
          required
        />

        <Input
          id="password"
          label="Password"
          type="password"
          autocomplete={isLogin ? 'current-password' : 'new-password'}
          bind:value={password}
          placeholder="Enter your password"
          required
        />

        {#if !isLogin}
          <Input
            id="confirmPassword"
            label="Confirm Password"
            type="password"
            autocomplete="new-password"
            bind:value={confirmPassword}
            placeholder="Confirm your password"
            required
          />
        {/if}
      </div>

      <Button type="submit" {loading} class="w-full">
        {loading ? 'Please wait...' : submitText}
      </Button>
    </form>
  </div>
</div>
