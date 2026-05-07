#!/usr/bin/env node

const { execSync } = require("node:child_process");

const defaultPorts = ["8080", "5173"];
const requestedPorts = process.argv.slice(2).filter((port) => /^\d+$/.test(port));
const ports = [...new Set(requestedPorts.length > 0 ? requestedPorts : defaultPorts)];

function getListeningPids(port) {
  try {
    const stdout = execSync(`lsof -tiTCP:${port} -sTCP:LISTEN`, {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    });
    return stdout
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
  } catch {
    return [];
  }
}

for (const port of ports) {
  const pids = getListeningPids(port);
  if (pids.length === 0) {
    console.log(`[port-clean] no process is listening on :${port}`);
    continue;
  }

  console.log(`[port-clean] killing :${port} -> pids: ${pids.join(", ")}`);
  for (const pid of pids) {
    try {
      process.kill(Number(pid), "SIGKILL");
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      console.warn(`[port-clean] failed to kill pid ${pid}: ${message}`);
    }
  }
}
