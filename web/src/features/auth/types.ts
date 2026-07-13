export type AuthMode = "login" | "register";

export interface AuthFormValues {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
}
