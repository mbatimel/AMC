import { Manager } from '@/core/shared/api/Manager';
import { parseApiResponse } from '@/core/shared/api/parseResponse';

import type {
  ChangePasswordParams,
  LoginParams,
  LoginResult,
  SignUpParams,
  SignUpResult,
  UserIdResponse,
  VerifyEmailParams,
} from './types';

import { AUTH_USER_ID_HEADER, AuthPath } from './url';

export class AuthManager extends Manager {
  changePassword(params: ChangePasswordParams): Promise<void> {
    const { newPassword, oldPassword, userId } = params;

    return this.httpClient
      .post(AuthPath.CHANGE_PASSWORD, {
        headers: { [AUTH_USER_ID_HEADER]: userId },
        query: { newPassword, oldPassword },
      })
      .then(({ body, status }) => parseApiResponse(body, status));
  }

  login(params: LoginParams): Promise<LoginResult> {
    return this.httpClient
      .post(AuthPath.LOGIN, { query: params })
      .then(({ body, status }) => {
        const { userID } = parseApiResponse<UserIdResponse>(body, status);

        return { userId: userID };
      });
  }

  logout(userId: string): Promise<void> {
    return this.httpClient
      .post(AuthPath.LOGOUT, {
        headers: { [AUTH_USER_ID_HEADER]: userId },
      })
      .then(({ body, status }) => parseApiResponse(body, status));
  }

  sendVerification(userId: string): Promise<void> {
    return this.httpClient
      .post(AuthPath.SEND_VERIFICATION, {
        headers: { [AUTH_USER_ID_HEADER]: userId },
      })
      .then(({ body, status }) => parseApiResponse(body, status));
  }

  signUp(params: SignUpParams): Promise<SignUpResult> {
    return this.httpClient
      .post(AuthPath.SIGN_UP, { query: params })
      .then(({ body, status }) => {
        const { userID } = parseApiResponse<UserIdResponse>(body, status);

        return { userId: userID };
      });
  }

  verifyEmail(params: VerifyEmailParams): Promise<void> {
    const { code, userId } = params;

    return this.httpClient
      .post(AuthPath.VERIFY_EMAIL, {
        headers: { [AUTH_USER_ID_HEADER]: userId },
        query: { code },
      })
      .then(({ body, status }) => parseApiResponse(body, status));
  }
}
