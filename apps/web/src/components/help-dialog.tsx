"use client";
import { ArrowUpRightIcon as ArrowUpRight } from "@phosphor-icons/react/dist/csr/ArrowUpRight";
import { LockKeyIcon as LockKeyhole } from "@phosphor-icons/react/dist/csr/LockKey";
import { useLocale } from "@/components/locale-provider";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
export function HelpDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { t } = useLocale();
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="text-xl tracking-tight">
            {t("helpTitle")}
          </DialogTitle>
          <DialogDescription>{t("helpIntro")}</DialogDescription>
        </DialogHeader>
        <div className="space-y-5 py-3 text-sm leading-6">
          <div>
            <h3 className="font-medium">{t("publicChoices")}</h3>
            <p className="mt-1 text-muted-foreground">
              {t("publicChoicesExplanation")}
            </p>
          </div>
          <div>
            <h3 className="font-medium">{t("instanceFiles")}</h3>
            <p className="mt-1 text-muted-foreground">
              {t("instanceFilesExplanation")}
            </p>
          </div>
          <div className="rounded-xl bg-secondary p-4">
            <h3 className="flex items-center gap-2 font-medium text-primary">
              <LockKeyhole weight="duotone" size={16} aria-hidden="true" />
              {t("accessBoundary")}
            </h3>
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              {t("accessExplanation")}
            </p>
          </div>
          <a
            className="inline-flex min-h-10 items-center gap-2 text-primary underline underline-offset-4"
            href="https://github.com/Lord-shaban/OpenDownload"
            target="_blank"
            rel="noreferrer"
          >
            {t("readDocs")} <ArrowUpRight size={14} aria-hidden="true" />
          </a>
        </div>
      </DialogContent>
    </Dialog>
  );
}
