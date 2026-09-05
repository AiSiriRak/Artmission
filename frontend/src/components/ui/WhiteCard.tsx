interface WhiteCard {
  children: React.ReactNode;
}

export function WhiteCard({ children }: WhiteCard) {
  return (
    <div className="m-10 w-full max-w-md rounded-lg bg-white p-16 shadow-card">
      {children}
    </div>
  );
}
