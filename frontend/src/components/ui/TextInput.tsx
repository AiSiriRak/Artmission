interface TextInput {
  value: string;
  onChange: (value: string) => void;
  type?: string;
  placeholder?: string;
}

export function TextInput({
  value,
  onChange,
  type = "text",
  placeholder,
}: TextInput) {
  return (
    <input
      type={type}
      value={value}
      placeholder={placeholder}
      onChange={(e) => onChange(e.target.value)}
      className="mt-1 w-full rounded border px-3 py-2 text-small"
    />
  );
}
