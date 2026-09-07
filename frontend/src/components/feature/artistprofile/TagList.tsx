"use client";

interface TagListProps {
  items: string[];
  variant?: "category" | "style";
}

export default function TagList({ items, variant = "category" }: TagListProps) {
  if (!items || items.length === 0) return <p className="text-gray-400">-</p>;

  // กำหนดสีตามชนิดของ Tag
  const tagStyle =
    variant === "style"
      ? "bg-secondary-600 text-primary-400"
      : "bg-accent-200 text-primary-400";

  return (
    <div className="flex flex-wrap gap-2">
      {items.map((item, index) => (
        <span
          key={index}
          className={`px-3 py-1 ${tagStyle} text-sm rounded-full font-medium`}
        >
          {item.trim()}
        </span>
      ))}
    </div>
  );
}