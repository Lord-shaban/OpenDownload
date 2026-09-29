"use client";
export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto max-w-lg p-12">
      <h1 className="text-2xl font-semibold">
        Something interrupted the workspace.
      </h1>
      <p className="my-4 text-muted-foreground">
        Your queued downloads are stored on the server. Reload to reconnect.
      </p>
      <button
        onClick={reset}
        className="rounded-lg bg-primary px-6 py-3 text-primary-foreground"
      >
        Try again
      </button>
    </main>
  );
}
