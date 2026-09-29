import { ArrowDownToLine } from "lucide-react";
export function Brand() {
  return (
    <span
      dir="ltr"
      className="flex items-center gap-2.5 [font-family:var(--font-geist-sans)]"
    >
      <span className="flex size-9 items-center justify-center rounded-[10px] bg-primary text-primary-foreground">
        <ArrowDownToLine size={21} strokeWidth={2} aria-hidden="true" />
      </span>
      <span className="text-[19px] font-semibold tracking-[-0.045em]">
        OpenDownload<span className="text-primary">.</span>
      </span>
    </span>
  );
}
