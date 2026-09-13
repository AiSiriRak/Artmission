"use client";

import TagList from "./TagList";
import { TextInput } from "@/components/ui/TextInput";

interface EditorFieldProps {
  label: string;
  name: string;
  value: string;
  isEditing: boolean;
  onChange: (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  isTextArea?: boolean;
  isTag?: boolean;
  tagVariant?: "category" | "style";
}

export default function EditorField({
  label,
  name,
  value,
  isEditing,
  onChange,
  isTextArea = false,
  isTag = false,
  tagVariant = "category",
}: EditorFieldProps) {
  return (
    <div className="mb-6">
      <label className="block text-body font-bold text-gray-800 mb-2">
        {label}
      </label>

      {isEditing ? (
        isTextArea ? (
          <textarea
            name={name}
            value={value}
            onChange={onChange}
            className="w-full border p-2 rounded h-24 focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        ) : (
          <TextInput
            name={name}
            value={value}
            onChange={(val) => {
              onChange({
                target: { name, value: val },
              } as React.ChangeEvent<HTMLInputElement>);
            }}
            placeholder={isTag ? "คั่นด้วยลูกน้ำ เช่น Pixel Art, Cartoon" : ""}
          />
        )
      ) : isTag ? (
        <TagList items={value.split(",")} variant={tagVariant} />
      ) : (
        <p className="text-gray-700 text-sm md:text-base leading-relaxed">
          {value}
        </p>
      )}
    </div>
  );
}
