import { execSync } from "child_process";
import * as fs from "fs";
import * as path from "path";

const ROOT = path.resolve(__dirname, "..");
const WEBAPP_DIR = path.join(ROOT, "webapp");
const APP_PID_FILE = "/tmp/ohm-e2e-app.pid";

export default async function globalTeardown(): Promise<void> {
  if (fs.existsSync(APP_PID_FILE)) {
    const pid = parseInt(fs.readFileSync(APP_PID_FILE, "utf-8"), 10);
    try {
      process.kill(pid, "SIGINT");
    } catch {
      // process may have already exited
    }
    await new Promise((r) => setTimeout(r, 2_000));
    fs.unlinkSync(APP_PID_FILE);
  }

  execSync("docker compose stop", { cwd: WEBAPP_DIR, stdio: "inherit" });
}
