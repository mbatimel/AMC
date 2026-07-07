export type AuthUserId = string;

export type ChangePasswordParams = {
  newPassword: string;
  oldPassword: string;
  userId: AuthUserId;
};

export type LoginParams = {
  email: string;
  password: string;
};

export type LoginResult = {
  userId: AuthUserId;
};

export type SignUpParams = {
  email: string;
  name: string;
  password: string;
  surename: string;
};

export type SignUpResult = {
  userId: AuthUserId;
};

export type UserIdResponse = {
  userID: AuthUserId;
};

export type VerifyEmailParams = {
  code: number;
  userId: AuthUserId;
};
