import { type KeyboardEvent, useEffect, useMemo, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

export interface Command {
  id: string;
  label: string;
  hint?: string;
  run: () => void;
}

/**
 * A keyboard-first command list. Ctrl or Cmd plus K opens it; Alt plus a digit runs the matching
 * command of the list directly, so every page is one chord away.
 */
export function CommandPalette({ commands }: { commands: Command[] }) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [index, setIndex] = useState(0);

  useEffect(() => {
    function onKey(e: globalThis.KeyboardEvent) {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setQuery("");
        setIndex(0);
        setOpen((v) => !v);
        return;
      }
      if (e.altKey && !e.ctrlKey && !e.metaKey && /^[1-9]$/.test(e.key)) {
        const command = commands[Number(e.key) - 1];
        if (command) {
          e.preventDefault();
          command.run();
        }
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [commands]);

  const shown = useMemo(() => {
    const q = query.trim().toLowerCase();
    return commands
      .map((command, position) => ({ command, position }))
      .filter(({ command }) => !q || command.label.toLowerCase().includes(q));
  }, [commands, query]);

  function choose(i: number) {
    const picked = shown[i];
    if (!picked) return;
    setOpen(false);
    picked.command.run();
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setIndex((v) => Math.min(v + 1, shown.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setIndex((v) => Math.max(v - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      choose(index);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="gap-3 p-4 sm:max-w-md" aria-label="Command palette">
        <DialogHeader>
          <DialogTitle>Command palette</DialogTitle>
          <DialogDescription>
            Type to filter. Alt plus 1 to 9 jumps straight to a page.
          </DialogDescription>
        </DialogHeader>
        <Input
          autoFocus
          aria-controls="command-list"
          aria-label="Search commands"
          placeholder="Search pages and actions"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIndex(0);
          }}
          onKeyDown={onKeyDown}
        />
        <ul id="command-list" className="max-h-72 overflow-auto text-sm">
          {shown.length === 0 && <li className="px-2 py-3 text-muted-foreground">No match</li>}
          {shown.map(({ command, position }, i) => (
            <li key={command.id}>
              <button
                type="button"
                aria-current={i === index ? "true" : undefined}
                className={`flex w-full items-center justify-between rounded-md px-2 py-2 text-left ${
                  i === index ? "bg-muted" : ""
                }`}
                onMouseEnter={() => setIndex(i)}
                onClick={() => choose(i)}
              >
                <span>{command.label}</span>
                <span className="text-xs text-muted-foreground">
                  {command.hint ?? (position < 9 ? `Alt+${position + 1}` : "")}
                </span>
              </button>
            </li>
          ))}
        </ul>
      </DialogContent>
    </Dialog>
  );
}
