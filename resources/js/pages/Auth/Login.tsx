import { Head } from "@inertiajs/react";
import { ArrowRight, LockKeyhole, ShieldCheck, Terminal } from "lucide-react";
import { Button } from "@/components/ui/button";
export default function Login({ configured }: { configured: boolean }) {
  return (
    <>
      <Head title="Sign in · vukem-trading" />
      <div className="grid min-h-dvh bg-background lg:grid-cols-2">
        <section className="relative hidden flex-col justify-between overflow-hidden border-r bg-[#0b1514] p-14 text-white lg:flex">
          <div className="absolute inset-0 bg-[linear-gradient(rgba(52,211,153,.035)_1px,transparent_1px),linear-gradient(90deg,rgba(52,211,153,.035)_1px,transparent_1px)] bg-[size:48px_48px]" />
          <div className="relative flex items-center gap-3">
            <Terminal className="text-emerald-400" />
            <span className="text-2xl font-semibold tracking-tight">
              vukem-trading<span className="text-emerald-400">.</span>
            </span>
          </div>
          <div className="relative max-w-lg">
            <p className="mb-5 text-xs uppercase tracking-[.3em] text-emerald-400">
              Your quant lab
            </p>
            <h1 className="text-5xl leading-[1.12] font-medium tracking-tight xl:text-6xl">
              Less intuition.
              <br />
              More <span className="text-emerald-400">evidence.</span>
            </h1>
            <p className="mt-7 max-w-md text-base leading-relaxed text-zinc-400">
              Develop your strategies. Understand the risks. Test your hypotheses before any
              exposure.
            </p>
            <div className="mt-10 flex gap-6 text-xs text-zinc-400">
              <span>01 / Brazil & US</span>
              <span>02 / Paper trading</span>
              <span>03 / Risk control</span>
            </div>
            <svg
              className="mt-12 w-full text-emerald-400/40"
              viewBox="0 0 400 70"
              aria-hidden="true"
            >
              <path
                d="M0 60 L30 54 L50 57 L80 39 L100 45 L130 20 L150 30 L175 23 L190 38 L220 28 L250 14 L270 20 L300 10 L330 18 L365 4 L400 9"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
              />
            </svg>
            <p className="text-[10px] text-zinc-600">
              Illustration. Does not represent financial results.
            </p>
          </div>
          <p className="relative text-[11px] text-zinc-500">
            PERSONAL TERMINAL / RESTRICTED ACCESS
          </p>
        </section>
        <section className="flex items-center justify-center p-8">
          <div className="w-full max-w-sm">
            <div className="mb-10 flex size-12 items-center justify-center rounded-2xl border bg-muted/40">
              <LockKeyhole className="size-5 text-emerald-500" />
            </div>
            <p className="text-xs uppercase tracking-widest text-muted-foreground">
              Welcome to vukem-trading
            </p>
            <h2 className="mt-3 text-3xl font-semibold tracking-tight">
              Sign in to your workspace
            </h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              Sign in with the Google account authorized by the terminal owner.
            </p>
            <Button asChild disabled={!configured} className="mt-8 h-12 w-full">
              <a
                href={configured ? "/auth/google/redirect" : "#"}
                aria-disabled={!configured}
                onClick={(e) => {
                  if (!configured) e.preventDefault();
                }}
              >
                <svg viewBox="0 0 24 24" className="mr-2 size-4" aria-hidden="true">
                  <path
                    fill="currentColor"
                    d="M21.6 12.2c0-.7-.1-1.5-.2-2.2H12v4.2h5.4c-.3 1.3-1 2.3-2 2.9v2.4h3.2c1.9-1.7 3-4.2 3-7.3ZM12 22c2.7 0 5-.9 6.6-2.4l-3.2-2.4c-.9.6-2 .9-3.4.9-2.6 0-4.9-1.8-5.7-4.2H3v2.5C4.6 19.7 8 22 12 22ZM6.3 13.9A6 6 0 0 1 6 12c0-.7.1-1.3.3-1.9V7.6H3A10 10 0 0 0 2 12c0 1.6.4 3.1 1 4.4l3.3-2.5ZM12 5.9c1.5 0 2.8.5 3.9 1.5l2.9-2.9A10 10 0 0 0 12 2c-4 0-7.4 2.3-9 5.6l3.3 2.5c.8-2.4 3.1-4.2 5.7-4.2Z"
                  />
                </svg>
                Continue with Google
                <ArrowRight className="ml-auto size-4" />
              </a>
            </Button>
            {!configured && (
              <p
                role="status"
                className="mt-4 rounded-lg border border-amber-500/20 bg-amber-500/5 p-3 text-xs text-amber-600"
              >
                Login waiting for OAuth configuration and an authorized account on the server.
              </p>
            )}
            <div className="mt-10 flex items-start gap-3 border-t pt-5 text-xs leading-relaxed text-muted-foreground">
              <ShieldCheck className="size-4 shrink-0 text-emerald-500" />
              <p>
                Simulation only. Login does not grant brokerage access. Credentials stay on the
                server.
              </p>
            </div>
          </div>
        </section>
      </div>
    </>
  );
}
