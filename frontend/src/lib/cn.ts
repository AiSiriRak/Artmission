import { clsx, type ClassValue } from "clsx";
import { extendTailwindMerge } from "tailwind-merge";

/**
 * Custom font-size utilities from `globals.css` / `@theme`
 * (`text-h1`, `text-small`, …). Register them so twMerge does not
 * treat them as conflicting with color utilities like `text-primary-500`.
 */
const fontSizeTokens = [
  "h1",
  "h2",
  "h3",
  "body",
  "small",
  "button",
  "caption",
  "subtle",
  "link",
] as const;

const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      text: [...fontSizeTokens],
    },
  },
});

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
