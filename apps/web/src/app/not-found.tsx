import Link from "next/link";
export default function NotFound() {
  return (
    <main className="mx-auto max-w-lg p-12">
      <h1 className="text-2xl font-semibold">This page isn’t here.</h1>
      <Link className="mt-6 inline-block text-primary underline" href="/">
        Return to OpenDownload
      </Link>
    </main>
  );
}
