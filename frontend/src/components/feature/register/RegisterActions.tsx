import type { MouseEvent } from "react";
import { Button } from "@/components/ui/Button";
import type { RegisterStep } from "./registerConfig";

interface RegisterActionsProps {
  step: RegisterStep;
  submitError: string;
  isSubmitting: boolean;
  onBack: (event: MouseEvent<HTMLButtonElement>) => void;
  onNext: (event: MouseEvent<HTMLButtonElement>) => void;
}

export function RegisterActions({
  step,
  submitError,
  isSubmitting,
  onBack,
  onNext,
}: RegisterActionsProps) {
  return (
    <div className="mt-20">
      <p
        role={submitError ? "alert" : undefined}
        className="mb-2 min-h-5 text-center text-caption text-error"
      >
        {submitError}
      </p>

      <div className="grid grid-cols-2 items-center">
        <div className="justify-self-start">
          {step > 1 && (
            <Button
              type="button"
              variant="light"
              className="flex h-9 w-[120px] items-center justify-center gap-2 rounded-xl border-primary-500 px-0 py-0 text-subtle text-primary-500"
              onClick={onBack}
            >
              <span aria-hidden="true">←</span>
              Back
            </Button>
          )}
        </div>

        <div className="justify-self-end">
          {step < 3 ? (
            <Button
              key="next-button"
              type="button"
              variant="dark"
              className="flex h-9 w-[120px] items-center justify-center gap-2 rounded-xl px-0 py-0 text-button text-white"
              onClick={onNext}
            >
              Next
              <span aria-hidden="true">→</span>
            </Button>
          ) : (
            <Button
              key="submit-button"
              type="submit"
              variant="dark"
              disabled={isSubmitting}
              aria-disabled={isSubmitting}
              className={`flex h-9 w-[148px] items-center justify-center rounded-xl px-0 py-0 text-subtle text-white transition ${
                isSubmitting ? "opacity-50" : "opacity-100"
              }`}
            >
              {isSubmitting ? "Creating..." : "Create account"}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
