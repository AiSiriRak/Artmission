import { cn } from "@/lib/cn";

interface WhiteCard {
  children: React.ReactNode;
  margin?: string;
  padding?: string;
  roundsize?: string;
  className?: string;
}

export function WhiteCard({
  children,
  roundsize = "rounded-lg",
  margin = "m-10",
  padding = "p-16",
  className,
}: WhiteCard) {
  return (
    <div
      className={cn(
        margin,
        roundsize,
        padding,
        "w-full max-w-md bg-white shadow-card",
        className,
      )}
    >
      {children}
    </div>
  );
}
