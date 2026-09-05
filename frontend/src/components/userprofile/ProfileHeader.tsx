// Profile Picture + Name

export function ProfileHeader() {
  return (
    <div className="mt-12 flex flex-col items-center">
      <div className="relative">
        <div className="h-24 w-24 rounded-full bg-neutral-400" />

        <button className="absolute bottom-0 right-0 rounded-full bg-neutral p-2">
          📷
        </button>
      </div>

      <h1 className="mt-4 text-h2 text-primary-500">User's Name</h1>
    </div>
  );
}
