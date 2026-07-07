import { createEvent, createStore } from 'effector';
import { readonly } from 'patronum';

export enum AuthStatus {
  Authenticated = 'authenticated',
  Unauthenticated = 'unauthenticated',
  Unknown = 'unknown',
}

export const authSessionReceived = createEvent();
export const authSessionCleared = createEvent();

export const $authStatus = createStore<AuthStatus>(AuthStatus.Unknown);

$authStatus
  .on(authSessionReceived, () => AuthStatus.Authenticated)
  .on(authSessionCleared, () => AuthStatus.Unauthenticated);

export const $isAuthenticated = readonly(
  $authStatus.map((status) => status === AuthStatus.Authenticated),
);
