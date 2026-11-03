import nextJest from "next/jest.js";

// Provide the path to the Next.js app to load next.config.js and .env files.
const createJestConfig = nextJest({ dir: "./" });

/** @type {import('jest').Config} */
const config = {
  testEnvironment: "node",
  // Playwright specs under e2e/ are run by `npm run e2e`, not Jest.
  testPathIgnorePatterns: ["/node_modules/", "/.next/", "/e2e/"],
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/src/$1",
  },
};

export default createJestConfig(config);
