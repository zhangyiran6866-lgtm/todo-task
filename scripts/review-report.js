#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const { writeFileSync } = require("node:fs");
const { join } = require("node:path");

const repoRoot = process.cwd();
const reportPath = join(repoRoot, "CR_REPORT.md");

function run(cmd, args) {
  const result = spawnSync(cmd, args, {
    cwd: repoRoot,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });

  return {
    cmd: `${cmd} ${args.join(" ")}`.trim(),
    code: result.status ?? 1,
    stdout: result.stdout || "",
    stderr: result.stderr || "",
  };
}

function safeRun(cmd, args) {
  try {
    return run(cmd, args);
  } catch (error) {
    return {
      cmd: `${cmd} ${args.join(" ")}`.trim(),
      code: 1,
      stdout: "",
      stderr: String(error),
    };
  }
}

function short(text, max = 1800) {
  if (!text) return "";
  return text.length > max ? `${text.slice(0, max)}\n... (truncated)` : text;
}

const timestamp = new Date().toISOString();

const stagedFiles = safeRun("git", ["diff", "--cached", "--name-only"]);
const changedFiles = stagedFiles.stdout
  .split("\n")
  .map(s => s.trim())
  .filter(Boolean);

const checks = [
  { key: "frontendLint", label: "Frontend Lint", cmd: "pnpm", args: ["run", "lint:frontend"] },
  { key: "frontendTypecheck", label: "Frontend Typecheck", cmd: "pnpm", args: ["run", "typecheck:frontend"] },
  { key: "backendVet", label: "Backend Vet", cmd: "pnpm", args: ["run", "vet:backend"] },
  { key: "backendTest", label: "Backend Test", cmd: "pnpm", args: ["run", "test:backend"] },
];

const results = checks.map(item => ({ ...item, ...run(item.cmd, item.args) }));
const allPassed = results.every(item => item.code === 0);

const hasNoTestFiles = results
  .find(item => item.key === "backendTest")
  ?.stdout.includes("[no test files]");

let debugCodeOutput = "";
if (changedFiles.length > 0) {
  const debugScan = safeRun("rg", ["-n", "console\\.log\\(|fmt\\.Println\\(", ...changedFiles]);
  debugCodeOutput = debugScan.code === 0 ? debugScan.stdout.trim() : "";
}

const report = `# CR Report

- Generated At: ${timestamp}
- Repository: ${repoRoot}
- Staged Files Count: ${changedFiles.length}
- Overall Result: ${allPassed ? "PASS" : "FAIL"}

## Staged Files
${changedFiles.length ? changedFiles.map(file => `- \`${file}\``).join("\n") : "- (none)"}

## Automated Checks
${results
  .map(
    item => `### ${item.label}
- Command: \`${item.cmd}\`
- Status: ${item.code === 0 ? "PASS" : "FAIL"}

\`\`\`text
${short(`${item.stdout}${item.stderr}`.trim() || "(no output)")}
\`\`\`
`
  )
  .join("\n")}

## 8-Dimension Review Summary

- Design: Manual review required (check architecture boundary and layering).
- Complexity: Manual review required (check simplification opportunities).
- Naming: Manual review required (check naming clarity and intent).
- Functionality: Manual review required (check behavior and user value).
- Tests: ${hasNoTestFiles ? "WARN (backend has no test files currently)." : "Check command output above."}
- Comments: Manual review required (check clarity and necessity).
- Style: ${allPassed ? "PASS (lint/typecheck/vet/test passed)." : "FAIL (see failed checks above)."}
- Documentation: Manual review required (check related docs updates).

## Extra Signals

- Debug print scan in staged files (\`console.log\` / \`fmt.Println\`):
${debugCodeOutput ? `\n\`\`\`text\n${short(debugCodeOutput)}\n\`\`\`` : " none detected"}
`;

writeFileSync(reportPath, report, "utf8");
console.log(`CR report generated: ${reportPath}`);

if (!allPassed) {
  process.exit(1);
}
