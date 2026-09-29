"use client";
import { ArrowUpRight, LockKeyhole } from "lucide-react";
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
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="text-xl tracking-tight">
            A few things worth knowing.
          </DialogTitle>
          <DialogDescription>
            OpenDownload saves media you own or have permission to download.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-5 py-3 text-sm leading-6">
          <div>
            <h3 className="font-medium">Public links, actual choices</h3>
            <p className="mt-1 text-muted-foreground">
              The extractor checks each link and returns the available formats.
              Platform support can change. Photo gallery coverage is currently
              limited to direct images and collections returned by the
              extractor.
            </p>
          </div>
          <div>
            <h3 className="font-medium">Your instance, temporary files</h3>
            <p className="mt-1 text-muted-foreground">
              Processing happens on your server. Files expire automatically;
              save them to your device before the expiry shown. Delete removes a
              job and its server files.
            </p>
          </div>
          <div className="rounded-xl bg-secondary p-4">
            <h3 className="flex items-center gap-2 font-medium text-primary">
              <LockKeyhole size={16} aria-hidden="true" />
              Access boundaries stay in place.
            </h3>
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              No private or paid content, login cookies, DRM bypass, live
              streams, or watermark removal. Use one public media link at a
              time.
            </p>
          </div>
          <a
            className="inline-flex min-h-10 items-center gap-2 text-primary underline underline-offset-4"
            href="https://github.com/Lord-shaban/OpenDownload"
            target="_blank"
            rel="noreferrer"
          >
            Read the project documentation{" "}
            <ArrowUpRight size={14} aria-hidden="true" />
          </a>
        </div>
      </DialogContent>
    </Dialog>
  );
}
