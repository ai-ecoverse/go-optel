import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { pathToFileURL } from "node:url";

export function moduleMajor(goMod) {
  const match = goMod.match(/^module\s+(\S+)/m);
  if (!match) {
    throw new Error("major-guard: go.mod has no module directive");
  }
  const suffix = match[1].match(/\/v(\d+)$/);
  return { path: match[1], major: suffix ? Number(suffix[1]) : 1 };
}

export async function verifyRelease(_config, { cwd, nextRelease, logger }) {
  const { path, major } = moduleMajor(await readFile(join(cwd, "go.mod"), "utf8"));
  const next = Number(nextRelease.version.split(".")[0]);
  if (next > major) {
    throw new Error(
      `major-guard: v${nextRelease.version} needs module path ${path.replace(/\/v\d+$/, "")}/v${next}, go.mod declares ${path}`,
    );
  }
  logger.log(`major-guard: v${nextRelease.version} fits ${path}`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    await verifyRelease({}, { cwd: process.cwd(), nextRelease: { version: process.argv[2] }, logger: console });
  } catch (error) {
    console.error(error.message);
    process.exit(1);
  }
}
