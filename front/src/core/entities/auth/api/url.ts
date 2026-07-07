export const AUTH_USER_ID_HEADER = 'X-User-Id';

export enum AuthPath {
  CHANGE_PASSWORD = '/api/v1/auth/change-password',
  LOGIN = '/api/v1/auth/login',
  LOGOUT = '/api/v1/auth/logout',
  SEND_VERIFICATION = '/api/v1/auth/send-verification',
  SIGN_UP = '/api/v1/auth/signup',
  VERIFY_EMAIL = '/api/v1/auth/verify-email',
}
