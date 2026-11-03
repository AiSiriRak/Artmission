// Accounts created by the backend seed (`go run . seed reseed`); all share one password.
const password = "Password123!";

export const accounts = {
  customerJane: { email: "jane@artmission.test", password },
  artistAnn: { email: "ann@artmission.test", password },
  artistBo: { email: "bo@artmission.test", password },
} as const;

export type AccountKey = keyof typeof accounts;

// Browser session saved by auth.setup.ts for each account.
export function authFile(key: AccountKey): string {
  return `e2e/.auth/${key}.json`;
}
