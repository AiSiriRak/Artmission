interface WhiteCard {
  children: React.ReactNode;
  margin?: string;
  padding?: string;
  roundsize?: string;
}

export function WhiteCard({
  children,
  roundsize = "rounded-lg",
  margin = "m-10",
  padding = "p-16",
}: WhiteCard) {
  return (
    <div
      className={`${margin} ${roundsize} ${padding} w-full max-w-md bg-white shadow-card`}
    >
      {children}
    </div>
  );
}
