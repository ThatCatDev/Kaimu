<script lang="ts">
  import { verifyEmail } from '../lib/api/auth';
  import { Button } from './ui';

  interface Props {
    token: string | null;
  }

  let { token }: Props = $props();

  let status = $state<'loading' | 'success' | 'error' | 'no-token'>('loading');
  let errorMessage = $state<string>('');

  $effect(() => {
    if (!token) {
      status = 'no-token';
      return;
    }

    verifyEmail(token)
      .then(() => {
        status = 'success';
        // Redirect to home after 2 seconds
        setTimeout(() => {
          window.location.href = '/';
        }, 2000);
      })
      .catch((error) => {
        status = 'error';
        errorMessage = error instanceof Error ? error.message : 'Failed to verify email';
      });
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
  <div class="max-w-md w-full space-y-8 text-center">
    {#if status === 'loading'}
      <div>
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
        <h2 class="mt-6 text-2xl font-bold text-gray-900">Verifying your email...</h2>
        <p class="mt-2 text-gray-600">Please wait while we verify your email address.</p>
      </div>
    {:else if status === 'success'}
      <div>
        <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
          <svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 class="mt-6 text-2xl font-bold text-gray-900">Email Verified!</h2>
        <p class="mt-2 text-gray-600">Your email has been successfully verified. Redirecting you to the homepage...</p>
      </div>
    {:else if status === 'error'}
      <div>
        <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
          <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h2 class="mt-6 text-2xl font-bold text-gray-900">Verification Failed</h2>
        <p class="mt-2 text-gray-600">{errorMessage}</p>
        <div class="mt-6">
          <Button onclick={() => window.location.href = '/'}>Go to Homepage</Button>
        </div>
      </div>
    {:else if status === 'no-token'}
      <div>
        <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100">
          <svg class="h-6 w-6 text-yellow-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <h2 class="mt-6 text-2xl font-bold text-gray-900">No Verification Token</h2>
        <p class="mt-2 text-gray-600">No verification token was provided. Please check your email for the verification link.</p>
        <div class="mt-6">
          <Button onclick={() => window.location.href = '/'}>Go to Homepage</Button>
        </div>
      </div>
    {/if}
  </div>
</div>
