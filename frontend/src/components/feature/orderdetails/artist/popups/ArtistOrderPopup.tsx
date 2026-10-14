"use client";

import { useEffect, useId, useRef, type ReactNode } from "react";

interface ArtistOrderPopupProps {
  title: string;
  children: ReactNode;
  onClose: () => void;
  form?: boolean;
  icon?: ReactNode;
}

export default function ArtistOrderPopup({ title, children, onClose, form = false, icon }: ArtistOrderPopupProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  useEffect(() => {
    const dialog = dialogRef.current;
    const previousFocus = document.activeElement;
    const previousOverflow = document.body.style.overflow;
    const page = document.documentElement;
    const previousScrollbarGutter = page.style.scrollbarGutter;
    // Keep the scrollbar's space reserved while background scrolling is locked.
    page.style.scrollbarGutter = "stable";
    dialog?.showModal();
    document.body.style.overflow = "hidden";
    return () => {
      dialog?.close();
      document.body.style.overflow = previousOverflow;
      page.style.scrollbarGutter = previousScrollbarGutter;
      if (previousFocus instanceof HTMLElement && previousFocus.isConnected) previousFocus.focus();
    };
  }, []);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      onCancel={(event) => { event.preventDefault(); onClose(); }}
      className={`fixed inset-0 m-auto max-h-[calc(100dvh-32px)] w-[calc(100%-32px)] overflow-y-auto bg-white text-primary-500 shadow-card backdrop:bg-black/25 ${form ? "max-w-[600px] rounded-2xl p-8" : "max-w-[380px] rounded-xl px-4 py-7"}`}
    >
      {form && (
        <button type="button" aria-label="Close popup" onClick={onClose} className="absolute right-8 top-6 cursor-pointer text-h3">×</button>
      )}
      {icon && <div aria-hidden="true" className="mb-6 flex justify-center">{icon}</div>}
      <h2 id={titleId} className={`text-center ${form ? "text-h2" : "text-h3"}`}>{title}</h2>
      {children}
    </dialog>
  );
}
