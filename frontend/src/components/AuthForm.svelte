<script lang="ts">
  import { login, register } from '../lib/stores/auth.svelte';

  interface Props {
    mode: 'login' | 'register';
  }

  let { mode }: Props = $props();

  let username = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let error = $state<string | null>(null);
  let loading = $state(false);

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

    <form class="mt-8 space-y-6" onsubmit={handleSubmit}>
      {#if error}
        <div class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-700">{error}</p>
        </div>
      {/if}

      <div class="space-y-4">
        <div>
          <label for="username" class="block text-sm font-medium text-gray-700">
            Username
          </label>
          <input
            id="username"
            name="username"
            type="text"
            autocomplete="username"
            required
            bind:value={username}
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Enter your username"
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700">
            Password
          </label>
          <input
            id="password"
            name="password"
            type="password"
            autocomplete={isLogin ? 'current-password' : 'new-password'}
            required
            bind:value={password}
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            placeholder="Enter your password"
          />
        </div>

        {#if !isLogin}
          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-gray-700">
              Confirm Password
            </label>
            <input
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              autocomplete="new-password"
              required
              bind:value={confirmPassword}
              class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              placeholder="Confirm your password"
            />
          </div>
        {/if}
      </div>

      <div>
        <button
          type="submit"
          disabled={loading}
          class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Please wait...' : submitText}
        </button>
      </div>
    </form>
  </div>
</div>
