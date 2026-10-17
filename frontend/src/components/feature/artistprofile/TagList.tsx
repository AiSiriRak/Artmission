"use client";

interface TagListProps {
  items: string[];
  variant?: "category" | "style";
}

function formatTagLabel(value: string) {
  const label = value.trim().toLowerCase();
  return label.charAt(0).toUpperCase() + label.slice(1);
}

export default function TagList({ items, variant = "category" }: TagListProps) {
  if (!items || items.length === 0) return <p className="text-gray-400">-</p>;

  // Category = pink
  // Style = brown
  const tagStyle =
    variant === "style"
      ? "bg-secondary-600 text-primary-400"
      : "bg-accent-200 text-primary-400";

  return (
    <div className="flex flex-wrap gap-2">
      {items.map((item, index) => (
        <span
          key={index}
          className={`px-3 py-1 ${tagStyle} text-caption rounded-[46px]`}
        >
          {formatTagLabel(item)}
        </span>
      ))}
    </div>
  );
}
