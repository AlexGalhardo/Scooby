import { Head, Link } from "@inertiajs/react";

import { demoInstalled, demoLink } from "../demo";

export default function Home() {
  return (
    <>
      <Head title="Scooby" />
      <section className="py-12 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">Scooby</h1>
        <p className="mt-3 text-zinc-600 dark:text-zinc-400">
          Vukem no servidor. React com Inertia, Tailwind CSS e PostgreSQL.
        </p>
        {demoInstalled && (
          <Link
            href={demoLink.href}
            className="mt-6 inline-block rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500"
          >
            Explorar posts
          </Link>
        )}
      </section>
    </>
  );
}
