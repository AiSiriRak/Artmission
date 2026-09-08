export function Loading() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center">
      <div className="h-12 w-12 animate-spin rounded-full border-4 border-neutral-200 border-t-primary-500" />

      <p className="mt-4 text-body text-primary-500">Loading...</p>
    </div>
  );
}
