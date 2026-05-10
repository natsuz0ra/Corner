import type { MenuItem, Settings } from "../types.js";

export const CLI_SANDBOX_MODES = [
  {
    value: "read-only",
    title: "Read Only",
    desc: "Commands can read files but cannot write through sandboxed tools",
  },
  {
    value: "workspace-write",
    title: "Workspace Write",
    desc: "Commands can write inside configured workspace roots",
  },
  {
    value: "danger-full-access",
    title: "Danger Full Access",
    desc: "Commands run without sandbox restrictions",
  },
] as const;

export type CliSandboxMode = (typeof CLI_SANDBOX_MODES)[number]["value"];

export type SandboxMenuAction =
  | { type: "mode"; mode: CliSandboxMode }
  | { type: "network"; enabled: boolean };

export function normalizeCliSandboxMode(mode?: string): CliSandboxMode {
  return CLI_SANDBOX_MODES.some((item) => item.value === mode)
    ? mode as CliSandboxMode
    : "workspace-write";
}

export function buildSandboxMenuItems(settings: Settings): MenuItem[] {
  const currentMode = normalizeCliSandboxMode(settings.cliSandboxMode);
  const networkEnabled = settings.cliSandboxNetworkEnabled !== false;
  return [
    ...CLI_SANDBOX_MODES.map((mode) => ({
      title: mode.title,
      desc: `${mode.value === currentMode ? "current · " : ""}${mode.desc}`,
      data: { type: "mode", mode: mode.value } satisfies SandboxMenuAction,
    })),
    {
      title: networkEnabled ? "Network Access: On" : "Network Access: Off",
      desc: "Toggle sandboxed command network access",
      data: { type: "network", enabled: !networkEnabled } satisfies SandboxMenuAction,
    },
  ];
}
