import { cn } from "@/lib/cn";

interface WhiteCardBorder {
  children: React.ReactNode;
  margin?: string;
  padding?: string;
  roundsize?: string;
  className?: string;
}

export function WhiteCardBorder({
  children,
  roundsize = "rounded-4xl",
  margin = "m-10",
  padding = "p-16",
  className,
}: WhiteCardBorder) {
  return (
    <div
      className={cn(
        margin,
        roundsize,
        padding,
        "w-full max-w-md bg-secondary-200 border-1 border-primary-500",
        className,
      )}
    >
      {children}
    </div>
  );
}
