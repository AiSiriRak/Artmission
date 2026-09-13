import { accountConstraints } from "@/lib/constraint";
import { BANK_LABELS } from "@/lib/types";

export type RegisterStep = 1 | 2 | 3;
export type Role = "artist" | "customer" | "";

export interface RegisterFormValues {
  role: Role;
  email: string;
  username: string;
  password: string;
  confirmPassword: string;
  bank: string;
  accountHolderName: string;
  accountNumber: string;
  acceptedTerms: boolean;
}

export type RegisterValidationField = keyof RegisterFormValues;
export type RegisterValidationErrors = Partial<
  Record<RegisterValidationField, string>
>;

export const registerSteps: RegisterStep[] = [1, 2, 3];

export const registerBanks = Object.entries(BANK_LABELS).map(
  ([value, label]) => ({
    value,
    label,
  }),
);

export const fieldErrorIds: Record<RegisterValidationField, string> = {
  role: "role-error",
  email: "email-error",
  username: "username-error",
  password: "password-error",
  confirmPassword: "confirm-password-error",
  bank: "bank-error",
  accountHolderName: "account-holder-name-error",
  accountNumber: "account-number-error",
  acceptedTerms: "accepted-terms-error",
};

export const registerValidationMessages = {
  role: {
    required: "Please select your role.",
  },
  email: {
    required: "Please enter your email address.",
    invalid: "Please enter a valid email address.",
    duplicate: "This email is already registered.",
  },
  username: {
    required: "Please enter your username.",
    minLength: "Username must be at least 3 characters.",
    maxLength: "Username must be 20 characters or less.",
  },
  bank: {
    required: "Please select your bank.",
  },
  accountHolderName: {
    required: "Please enter the account holder's name.",
  },
  accountNumber: {
    required: "Please enter your bank account number.",
    invalid: "Please enter a valid bank account number.",
  },
  password: {
    required: "Please enter a password.",
    minLength: "Password must be at least 8 characters.",
    maxLength: "Password must be 16 characters or less.",
  },
  confirmPassword: {
    mismatch: "Passwords do not match.",
  },
  policy: {
    required:
      "Please agree to the Terms of Service and Privacy Policy to continue.",
  },
} as const;

export function validateRegisterStepFields(
  step: RegisterStep,
  values: RegisterFormValues,
): RegisterValidationErrors {
  const errors: RegisterValidationErrors = {};
  const email = values.email.trim();
  const username = values.username.trim();
  const accountHolderName = values.accountHolderName.trim();
  const accountNumber = values.accountNumber.trim();

  if (step === 1) {
    if (!values.role) {
      errors.role = registerValidationMessages.role.required;
    }

    return errors;
  }

  if (step === 2) {
    if (!email) {
      errors.email = registerValidationMessages.email.required;
    } else if (!accountConstraints.email.pattern.test(email)) {
      errors.email = registerValidationMessages.email.invalid;
    }

    if (!username) {
      errors.username = registerValidationMessages.username.required;
    } else if (username.length < accountConstraints.username.minLength) {
      errors.username = registerValidationMessages.username.minLength;
    } else if (username.length > accountConstraints.username.maxLength) {
      errors.username = registerValidationMessages.username.maxLength;
    }

    const hasValidPasswordLength =
      values.password.length >= accountConstraints.password.minLength &&
      values.password.length <= accountConstraints.password.maxLength;

    if (!values.password) {
      errors.password = registerValidationMessages.password.required;
    } else if (values.password.length < accountConstraints.password.minLength) {
      errors.password = registerValidationMessages.password.minLength;
    } else if (values.password.length > accountConstraints.password.maxLength) {
      errors.password = registerValidationMessages.password.maxLength;
    }

    if (hasValidPasswordLength && values.password !== values.confirmPassword) {
      errors.confirmPassword =
        registerValidationMessages.confirmPassword.mismatch;
    }

    return errors;
  }

  if (!values.bank) {
    errors.bank = registerValidationMessages.bank.required;
  }

  if (!accountHolderName) {
    errors.accountHolderName =
      registerValidationMessages.accountHolderName.required;
  }

  if (!accountNumber) {
    errors.accountNumber = registerValidationMessages.accountNumber.required;
  } else if (
    !accountConstraints.bankAccount.accountNumber.pattern.test(accountNumber)
  ) {
    errors.accountNumber = registerValidationMessages.accountNumber.invalid;
  }

  if (!values.acceptedTerms) {
    errors.acceptedTerms = registerValidationMessages.policy.required;
  }

  return errors;
}

export function hasValidationErrors(errors: RegisterValidationErrors) {
  return Object.keys(errors).length > 0;
}
