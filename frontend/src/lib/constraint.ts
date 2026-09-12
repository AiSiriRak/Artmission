export const accountConstraints = {
  username: {
    minLength: 3,
    maxLength: 20,
  },
  password: {
    minLength: 8,
    maxLength: 16,
  },
  email: {
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  },
  bankAccount: {
    bankName: {
      minLength: 1,
    },
    accountHolderName: {
      minLength: 1,
    },
    accountNumber: {
      minLength: 6,
      maxLength: 20,
      pattern: /^\d{6,20}$/,
    },
  },
  artistProfile: {
    description: {
      minLength: 0,
    },
  },
} as const;
